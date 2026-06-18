package server

import (
	"fmt"
	"log/slog"
	"net"
)

type baseServer struct {
	Addr     string
	listener net.Listener
}

func (s *baseServer) close() error {
	if s.listener == nil {
		return nil
	}
	return s.listener.Close()
}

func (s *baseServer) trackListener(ln net.Listener, add bool) error {
	if add {
		if s.listener != nil {
			return s.close()
		}
		s.listener = ln
	} else {
		s.listener = nil
	}
	return nil
}

func (s *baseServer) listenAndServe(connHandler func(net.Conn)) error {
	if s.close() != nil {
		return ErrServerClosed
	}

	addr := s.Addr
	if addr == "" {
		addr = ":http"
	}

	ln, err := net.Listen("tcp", addr)
	slog.Info("Started server", "addr", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.Addr, err)
	}

	return s.serve(ln, connHandler)
}

func (s *baseServer) serve(l net.Listener, connHandler func(net.Conn)) error {
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
		go connHandler(conn)
	}
}
