package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/tomo-local/http-server/internal/request"
)

type Handler func(req request.Request, conn net.Conn)

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
	log.Printf("Server status: Started listening on %s\r\n", s.addr)
	printSeparator(65)

	for {
		// 接続確率
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("accept error: %v\n", err)
			continue
		}
		fmt.Printf("conn: %v\n", conn)
		printSeparator(50)

		go s.ServeConn(conn)
	}
}

func (s *Server) ServeConn(conn net.Conn) {
	defer conn.Close()

	// addの取得
	addr := conn.RemoteAddr()
	log.Printf("addr: %v\n", addr)
	fmt.Println("====================")

	reader := bufio.NewReader(conn)
	for {
		req, err := request.Parse(reader)
		if err == io.EOF {
			log.Printf("connect close add: %v\n", addr)
			printSeparator(30)
			break
		}

		// errがある場合は break
		if err != nil {
			log.Printf("add: %v, read err: %v\n", addr, err)
			printSeparator(30)
			break
		}

		log.Printf("request: %s %s %s\n", req.Method, req.Path, req.Version)

		s.handler(req, conn)
	}
}

func printSeparator(w int) {
	fmt.Println(strings.Repeat("=", w))
}
