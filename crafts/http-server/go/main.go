package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/tomo-local/http-server/internal/request"
	"github.com/tomo-local/http-server/internal/server"
)

func main() {
	// サーバーの立ち上げ
	srv := server.NewServer(":8080", handleRequest)

	err := srv.ListenAndServe()

	if err != nil {
		log.Fatal(err)
	}
}

func handleRequest(req request.Request, conn net.Conn) {
	var status, body string
	switch {
	case req.Method == "POST" && req.ContentLength <= 0:
		status = "400 Bad Request"
		body = "missing body"
	case req.Method == "POST" && req.Path == "/echo":
		buf := make([]byte, req.ContentLength)
		if _, err := io.ReadFull(req.Body, buf); err != nil {
			fmt.Printf("err: %v", err)
			status = "500 Internal Server Error"
			body = "failed to read body"
			break
		}
		status = "200"
		body = string(buf)
	case req.Path == "/":
		status = "200 OK"
		body = "Welcome!"
	case req.Path == "/about":
		status = "200 OK"
		body = "About Path"
	default:
		status = "404 Not Found"
		body = "Not Found"
	}

	// 特定の形式に変更する
	response := "HTTP/1.1 " + status + "\r\n" +
		"Content-Length: " + strconv.Itoa(len(body)) + "\r\n" +
		"\r\n" +
		body

	_, err := conn.Write([]byte(response))

	if err != nil {
		fmt.Printf("write err: %v\r\n", err)
	}
}
