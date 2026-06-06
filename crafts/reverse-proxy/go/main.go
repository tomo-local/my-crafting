package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"reverse-proxy/server"
)

type Args struct {
	upstream string
	port     string
}

func main() {
	args := parseArgs()
	fmt.Printf("Args upstream:%s, port:%s\r\n", args.upstream, args.port)

	srv := server.NewHTTPServer(args.port, &ReverseProxyHandler{Upstream: args.upstream})

	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func parseArgs() Args {
	var port = ":8080"
	upstream := flag.String("upstream", "localhost:9001", "接続先のアドレス")
	port := flag.String("port", ":8080", "サーバーのポート")

	return Args{
		upstream: *upstream,
		port:     *port,
	}
}

type ReverseProxyHandler struct {
	Upstream string
}

func (r *ReverseProxyHandler) ServerHTTP(req server.Request, write server.Write) {
	write(server.StatusOK, "Hello, world!")
}
