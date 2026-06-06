package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler interface {
	ServerHTTP(req Request, writeResponse Write)
	ServerReverseProxy(req Request, conn net.Conn)
}

type Server struct {
	Addr         string
	Handler      Handler
	listener     net.Listener
	ReverseProxy bool
}

func NewHTTPServer(addr string, handler Handler) *Server {
	return &Server{
		Addr:         addr,
		Handler:      handler,
		ReverseProxy: false,
	}
}

func (s *Server) Close() error {
	if s.listener == nil {
		// The Close operation will not be executed because the server has not started yet.
		return nil
	}

	return s.listener.Close()
}

func (s *Server) SetReverseProxy() {
	s.ReverseProxy = true
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
			s.Handler.ServerReverseProxy(req, conn)
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

type Request struct {
	Method        string
	Path          string
	Version       string
	Host          string
	ContentLength int
	Connection    string
	Header        http.Header
	Body          io.Reader
}

func NewRequest(r io.Reader) (Request, error) {
	reader, ok := r.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return Request{}, err
	}

	fields := strings.Fields(strings.TrimRight(line, "\r\n"))
	if len(fields) != 3 {
		return Request{}, fmt.Errorf("invalid request line: %q", line)
	}

	contentLength := 0
	connection := ""
	host := ""
	headers := make(http.Header)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return Request{}, err
		}
		if line == "\r\n" {
			break
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers.Set(name, value)
		}

		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(lower, "content-length:"):
			parts := strings.SplitN(line, ":", 2)
			contentLength, err = strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return Request{}, fmt.Errorf("invalid Content-Length: %w", err)
			}

		case strings.HasPrefix(lower, "connection:"):
			parts := strings.SplitN(line, ":", 2)
			connection = strings.ToLower(strings.TrimSpace(parts[1]))
		case strings.HasPrefix(lower, "host:"):
			parts := strings.SplitN(line, ":", 2)
			host = strings.TrimSpace(parts[1])
		}

	}

	return Request{
		Method:        fields[0],
		Path:          fields[1],
		Version:       fields[2],
		Host:          host,
		ContentLength: contentLength,
		Connection:    connection,
		Header:        headers,
		Body:          reader,
	}, nil
}

func (r Request) WantsKeepAlive() bool {
	if r.Version == "HTTP/1.1" {
		return r.Connection != "close"
	}
	if r.Version == "HTTP/1.0" {
		return r.Connection == "keep-alive"
	}
	return false
}

func (r Request) Write(w io.Writer) error {
	fmt.Fprintf(w, "%s %s %s\r\n", r.Method, r.Path, r.Version)
	if err := r.Header.Write(w); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "\r\n")
	return err
}

type Response struct {
	conn      net.Conn
	keepAlive bool
}

type StatusCode string

const (
	StatusOK                  StatusCode = "200 OK"
	StatusBadRequest          StatusCode = "400 Bad Request"
	StatusNotFound            StatusCode = "404 Not Found"
	StatusInternalServerError StatusCode = "500 Internal Server Error"
)

func NewResponse(conn net.Conn) *Response {
	return &Response{conn: conn}
}

func (r *Response) SetKeepAlive(keepAlive bool) {
	r.keepAlive = keepAlive
}

type Write = func(status StatusCode, body string) error

func (r *Response) Write(status StatusCode, body string) error {
	connHeader := "Connection: close\r\n"
	if r.keepAlive {
		connHeader = "Connection: keep-alive\r\n"
	}
	response := "HTTP/1.1 " + string(status) + "\r\n" +
		"Content-Length: " + strconv.Itoa(len(body)) + "\r\n" +
		connHeader +
		"\r\n" +
		body

	_, err := r.conn.Write([]byte(response))
	return fmt.Errorf("fail write err: %v", err)
}
