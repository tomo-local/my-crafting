package response

import (
	"fmt"
	"net"
	"strconv"
)

type Response struct {
	conn net.Conn
}

type StatusCode string

const (
	StatusOK                  StatusCode = "200 OK"
	StatusBadRequest          StatusCode = "400 Bad Request"
	StatusNotFound            StatusCode = "404 Not Found"
	StatusInternalServerError StatusCode = "500 Internal Server Error"
)

func NewResponse(conn net.Conn) *Response {
	return &Response{
		conn: conn,
	}
}

type Write = func(status StatusCode, body string) error

func (r *Response) Write(status StatusCode, body string) error {
	response := "HTTP/1.1 " + string(status) + "\r\n" +
		"Content-Length: " + strconv.Itoa(len(body)) + "\r\n" +
		"Connection: keep-alive\r\n" +
		"\r\n" +
		body

	_, err := r.conn.Write([]byte(response))
	return fmt.Errorf("fail write err: %v", err)
}
