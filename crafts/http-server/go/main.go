package main

import (
	"fmt"
	"io"
	"log"

	"github.com/tomo-local/http-server/internal/request"
	"github.com/tomo-local/http-server/internal/response"
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

func handleRequest(req request.Request, witeResponse response.Write) {
	switch {
	case req.Method == "POST" && req.ContentLength <= 0:
		witeResponse(
			response.Status.BadRequest,
			"missing body",
		)
	case req.Method == "POST" && req.Path == "/echo":
		buf := make([]byte, req.ContentLength)
		if _, err := io.ReadFull(req.Body, buf); err != nil {
			fmt.Printf("err: %v", err)

			witeResponse(
				response.Status.InternalServerError,
				"failed to read body",
			)
			break
		}
		witeResponse(
			response.Status.OK,
			string(buf),
		)
	case req.Path == "/":
		witeResponse(
			response.Status.OK,
			"Welcome!",
		)
	case req.Path == "/about":
		witeResponse(
			response.Status.OK,
			"About Path",
		)
	default:
		witeResponse(
			response.Status.NotFound,
			"Not Found",
		)
	}
}
