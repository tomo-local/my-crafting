package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"reverse-proxy/balancer"
	"time"
)

type Handler interface {
	ServerHTTP(req Request, writeResponse Write)
	ServerReverseProxy(req Request, conn net.Conn, upstreamConn net.Conn)
}

type Server struct {
	Addr         string
	Handler      Handler
	listener     net.Listener
	ReverseProxy bool
	RoundRobin   balancer.RoundRobin
}

func NewHTTPServer(addr string, handler Handler) *Server {
	return &Server{
		Addr:         addr,
		Handler:      handler,
		ReverseProxy: false,
	}
}

func NewReverseProxyServer(addr string, handler Handler, upstreams []string) (*Server, error) {
	r, err := balancer.NewRoundRobin(upstreams)
	if err != nil {
		return nil, fmt.Errorf("failed to create round robin: %w", err)
	}

	return &Server{
		Addr:         addr,
		Handler:      handler,
		ReverseProxy: true,
		RoundRobin:   *r,
	}, nil
}

func (s *Server) Close() error {
	if s.listener == nil {
		// The Close operation will not be executed because the server has not started yet.
		return nil
	}

	return s.listener.Close()
}

var ErrServerClosed = errors.New("server: Server closed")
var ErrAlreadyStarted = errors.New("server: Already started server")

func (s *Server) ListenAndServe() error {
	if s.Close() != nil {
		return ErrServerClosed
	}

	addr := s.Addr
	if addr == "" {
		addr = ":http"
	}

	if s.ReverseProxy {
		s.RoundRobin.StartHealthCheck(10 * time.Second)
	}

	ln, err := net.Listen("tcp", addr)
	slog.Info("Stated server addr", "addr", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.Addr, err)
	}

	return s.serve(ln)
}

func (s *Server) serve(l net.Listener) error {
	if s.trackListener(l, true) != nil {
		return ErrServerClosed
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			slog.Error("failed to accept connection", "err", err)
			continue
		}
		slog.Info("connection accepted", "remote_addr", conn.RemoteAddr())
		go s.ServeConn(conn)
	}
}

func (s *Server) trackListener(ln net.Listener, add bool) error {
	if add {
		if s.listener != nil {
			return s.Close()
		}
		s.listener = ln
	} else {
		s.listener = nil
	}

	return nil
}

func (s *Server) ServeConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	addr := conn.RemoteAddr()
	slog.Info("start serving connection", "addr", addr)

	const idleTimeout = 30 * time.Second

	for {
		conn.SetReadDeadline(time.Now().Add(idleTimeout))
		req, err := NewRequest(reader)
		if err != nil {
			var netErr net.Error
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || (errors.As(err, &netErr) && netErr.Timeout()) {
				return
			}
			slog.Error("failed to parse request", "addr", addr, "err", err)
			return
		}

		slog.Info("request received", "method", req.Method, "path", req.Path, "version", req.Version)
		keepAlive := req.WantsKeepAlive()

		if s.ReverseProxy {
			upstream, err := s.RoundRobin.Next()
			if err != nil {
				slog.Error("no available upstream:", "err", err)
				res := NewResponse(conn)
				res.SetKeepAlive(keepAlive)
				res.Write(StatusBadGateway, "Bad Gateway")
				return
			}

			upstreamConn, err := net.Dial("tcp", upstream)
			if err != nil {
				slog.Error("Reverse Proxy Error")
				return
			}
			defer upstreamConn.Close()

			req.Host = "upstream"
			req.Header.Set("Host", upstream)
			s.Handler.ServerReverseProxy(req, conn, upstreamConn)
			return
		}

		res := NewResponse(conn)
		res.SetKeepAlive(keepAlive)
		s.Handler.ServerHTTP(req, res.Write)
		if !keepAlive {
			return
		}
	}

}
