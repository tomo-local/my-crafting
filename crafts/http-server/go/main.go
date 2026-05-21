package main

import (
	"io"
	"log/slog"
	"os"

	"github.com/tomo-local/http-server/internal/request"
	"github.com/tomo-local/http-server/internal/response"
	"github.com/tomo-local/http-server/internal/server"
)

func main() {
	srv := server.NewServer(":8080", handleRequest)

	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func handleRequest(req request.Request, writeResponse response.Write) {
	switch {
	case req.Method == "POST" && req.ContentLength <= 0:
		if err := writeResponse(response.StatusBadRequest, "missing body"); err != nil {
			slog.Error("failed to write response", "err", err)
		}
	case req.Method == "POST" && req.Path == "/echo":
		buf := make([]byte, req.ContentLength)
		if _, err := io.ReadFull(req.Body, buf); err != nil {
			slog.Error("failed to read body", "err", err)
			if err := writeResponse(response.StatusInternalServerError, "failed to read body"); err != nil {
				slog.Error("failed to write response", "err", err)
			}
			break
		}
		if err := writeResponse(response.StatusOK, string(buf)); err != nil {
			slog.Error("failed to write response", "err", err)
		}
	case req.Path == "/":
		if err := writeResponse(response.StatusOK, "Welcome!"); err != nil {
			slog.Error("failed to write response", "err", err)
		}
	case req.Path == "/about":
		if err := writeResponse(response.StatusOK, "About Path"); err != nil {
			slog.Error("failed to write response", "err", err)
		}
	default:
		if err := writeResponse(response.StatusNotFound, "Not Found"); err != nil {
			slog.Error("failed to write response", "err", err)
		}
	}
}
