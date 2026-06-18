package server

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"net"
	b "reverse-proxy/balancer"
	"time"
)

type ReverseProxyHandler interface {
	ServerReverseProxy(req Request, conn net.Conn, upstreamConn net.Conn)
}

type ReverseProxyServer struct {
	baseServer
	Handler  ReverseProxyHandler
	balancer b.Balancer
}

func (s *ReverseProxyServer) ListenAndServe() error {
	s.balancer.StartHealthCheck()
	return s.listenAndServe(s.serveConn)
}

func (s *ReverseProxyServer) serveConn(conn net.Conn) {
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

		upstream, err := s.balancer.Next()
		if err != nil {
			slog.Error("no available upstream:", "err", err)
			res := NewResponse(conn)
			res.SetKeepAlive(keepAlive)
			res.Write(StatusBadGateway, "Bad Gateway")
			return
		}

		ok := func() bool {
			upstreamConn, err := net.Dial("tcp", upstream)
			if err != nil {
				slog.Error("failed to connect upstream", "upstream",
					upstream, "err", err)
				res := NewResponse(conn)
				res.SetKeepAlive(keepAlive)
				res.Write(StatusBadGateway, "Bad Gateway")
				return false
			}

			req.Host = "upstream"
			req.Header.Set("Host", upstream)

			defer upstreamConn.Close()
			s.Handler.ServerReverseProxy(req, conn, upstreamConn)
			return true
		}()

		if !ok || !keepAlive {
			return
		}
	}
}
