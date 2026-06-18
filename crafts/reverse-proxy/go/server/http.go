package server

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"net"
	"time"
)

type HttpHandler interface {
	ServerHTTP(req Request, writeResponse Write)
}

type HttpServer struct {
	baseServer
	Handler HttpHandler
}

func (s *HttpServer) ListenAndServe() error {
	return s.listenAndServe(s.serveConn)
}

func (s *HttpServer) serveConn(conn net.Conn) {
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

		res := NewResponse(conn)
		res.SetKeepAlive(keepAlive)
		s.Handler.ServerHTTP(req, res.Write)
		if !keepAlive {
			return
		}
	}
}
