package server

import (
	"fmt"
	"net"
	"strconv"
)

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
