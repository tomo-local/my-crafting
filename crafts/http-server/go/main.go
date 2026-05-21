package main

import (
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

func handleRequest(req request.Request, writeResponse response.Write) {
	switch {
	case req.Method == "POST" && req.ContentLength <= 0:
		if err := writeResponse(response.StatusBadRequest, "missing body"); err != nil {
			log.Printf("write err: %v", err)
		}
	case req.Method == "POST" && req.Path == "/echo":
		buf := make([]byte, req.ContentLength)
		if _, err := io.ReadFull(req.Body, buf); err != nil {
			log.Printf("err: %v", err)
			if err := writeResponse(response.StatusInternalServerError, "failed to read body"); err != nil {
				log.Printf("write err: %v", err)
			}
			break
		}
		if err := writeResponse(response.StatusOK, string(buf)); err != nil {
			log.Printf("write err: %v", err)
		}
	case req.Path == "/":
		if err := writeResponse(response.StatusOK, "Welcome!"); err != nil {
			log.Printf("write err: %v", err)
		}
	case req.Path == "/about":
		if err := writeResponse(response.StatusOK, "About Path"); err != nil {
			log.Printf("write err: %v", err)
		}
	default:
		if err := writeResponse(response.StatusNotFound, "Not Found"); err != nil {
			log.Printf("write err: %v", err)
		}
	}
}
