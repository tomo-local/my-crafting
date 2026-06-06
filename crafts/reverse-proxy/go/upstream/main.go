package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"reverse-proxy/server"
	"strings"
)

type Args struct {
	id   string
	port string
}

func main() {
	args := parseArgs()
	fmt.Printf("Args id:%s, port:%s\r\n", args.id, args.port)

	addr := args.port
	if !strings.HasPrefix(args.port, ":") {
		addr = ":" + args.port
	}

	srv := server.NewHTTPServer(addr, &UpstreamHandler{Id: args.id})

	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func parseArgs() Args {
	id := flag.String("id", "upstream-1", "接続先のアドレス")
	port := flag.String("port", "8080", "サーバーのポート")
	flag.Parse()

	return Args{
		id:   *id,
		port: *port,
	}
}

type UpstreamHandler struct {
	Id string
}

func (r *UpstreamHandler) ServerHTTP(req server.Request, write server.Write) {
	write(server.StatusOK, "Hello, "+r.Id+"!")
}

func (r *UpstreamHandler) ServerReverseProxy(conn net.Conn) {}
