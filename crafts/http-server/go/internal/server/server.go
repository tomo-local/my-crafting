package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/tomo-local/http-server/internal/request"
	"github.com/tomo-local/http-server/internal/response"
)

type Handler func(req request.Request, writeResponse response.Write)

type Server struct {
	addr    string
	handler Handler
}

func NewServer(addr string, handler Handler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
	}
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}
	defer ln.Close()

	printSeparator(65)
	slog.Info("server started", "addr", s.addr)
	printSeparator(65)

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("failed to accept connection", "err", err)
			continue
		}
		slog.Info("connection accepted", "remote_addr", conn.RemoteAddr())
		printSeparator(50)

		go s.ServeConn(conn)
	}
}

func (s *Server) ServeConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	addr := conn.RemoteAddr()
	slog.Info("start serving connection", "addr", addr)

	// timeoutを指定
	const idleTimeout = 30 * time.Second

	for {
		conn.SetReadDeadline(time.Now().Add(idleTimeout))
		req, err := request.Parse(reader)

		if err != nil {
			var netErr net.Error
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || (errors.As(err, &netErr) && netErr.Timeout()) {
				return
			}
			slog.Error("failed to parse request", "addr", addr, "err", err)
			printSeparator(30)
			return
		}

		slog.Info("request received", "method", req.Method, "path", req.Path, "version", req.Version)
		keepAlive := req.WantsKeepAlive()
		res := response.NewResponse(conn)
		res.SetKeepAlive(keepAlive)
		s.handler(req, res.Write)
		if !keepAlive {
			return
		}
	}
}

func printSeparator(w int) {
	fmt.Println(strings.Repeat("=", w))
}
