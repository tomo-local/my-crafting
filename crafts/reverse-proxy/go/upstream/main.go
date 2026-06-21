package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"reverse-proxy/server"
	"strings"
	"time"
)

type Args struct {
	id         string
	port       string
	echoHeader bool
	delay      time.Duration
}

func main() {
	args := parseArgs()
	fmt.Printf("Args id:%s, port:%s echo-header:%v\r\n", args.id, args.port, args.echoHeader)

	addr := args.port
	if !strings.HasPrefix(args.port, ":") {
		addr = ":" + args.port
	}

	srv := server.NewHTTPServer(addr, &UpstreamHandler{Id: args.id, EchoHeader: args.echoHeader, Delay: args.delay})

	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func parseArgs() Args {
	id := flag.String("id", "upstream-1", "接続先のアドレス")
	port := flag.String("port", "8080", "サーバーのポート")
	echoHeader := flag.Bool("echo-header", true, "")
	delay := flag.Duration("delay", 0, "レスポンス遅延")

	flag.Parse()

	return Args{
		id:         *id,
		port:       *port,
		echoHeader: *echoHeader,
		delay:      *delay,
	}
}

type UpstreamHandler struct {
	Id         string
	EchoHeader bool
	Delay      time.Duration
}

func (r *UpstreamHandler) ServerHTTP(req server.Request, write server.Write) {
	if r.EchoHeader {
		for key, values := range req.Header {
			fmt.Printf("%s: %s\n", key, strings.Join(values, ", "))
		}
	}

	time.Sleep(r.Delay)
	write(server.StatusOK, "Hello, "+r.Id+"!")
}
