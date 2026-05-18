package server

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type Handler func(conn net.Conn)

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

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}
	defer ln.Close()

	logSeparator(65)
	log.Printf("Server status: Started listening on %s\r\n", s.addr)
	logSeparator(65)

	for {
		// 接続確率
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("accept error: %v\n", err)
			continue
		}
		log.Printf("conn: %v\n", conn)
		logSeparator(50)

		go s.handler(conn)
	}
}

func logSeparator(w int) {
	fmt.Println(strings.Repeat("=", w))
}
