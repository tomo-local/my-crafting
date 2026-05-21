package response

import (
	"net"
	"strconv"
)

type Response struct {
	conn net.Conn
}

type StatusCode string

type statusCodes struct {
	OK                  StatusCode
	BadRequest          StatusCode
	NotFound            StatusCode
	InternalServerError StatusCode
}

var Status = statusCodes{
	OK:                  "200 OK",
	BadRequest:          "400 Bad Request",
	NotFound:            "404 Not Found",
	InternalServerError: "500 Internal Server Error",
}

func NewResponse(conn net.Conn) *Response {
	return &Response{
		conn: conn,
	}
}

type Write = func(status StatusCode, body string) error

func (r *Response) Write(status StatusCode, body string) error {
	response := "HTTP/1.1 " + string(status) + "\r\n" +
		"Content-Length: " + strconv.Itoa(len(body)) + "\r\n" +
		"\r\n" +
		body

	_, err := r.conn.Write([]byte(response))
	return err
}
