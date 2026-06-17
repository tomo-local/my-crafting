package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"reverse-proxy/server"
	"strings"
)

type Args struct {
	upstreams []string
	port      string
}

func main() {
	args := parseArgs()
	fmt.Printf("Args upstreams:%s, port:%s\r\n", args.upstreams, args.port)

	addr := args.port
	if !strings.HasPrefix(args.port, ":") {
		addr = ":" + args.port
	}

	srv, err := server.NewReverseProxyServer(addr, &ReverseProxyHandler{}, args.upstreams)

	if err != nil {
		slog.Error("failed to create server", "err", err)
		os.Exit(1)
	}

	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func parseArgs() Args {
	upstreamsFlag := flag.String("upstreams", "localhost:9001,localhost:9002", "接続先のアドレス")
	port := flag.String("port", "8080", "サーバーのポート")
	flag.Parse()

	upstreams := strings.Split(*upstreamsFlag, ",")

	if len(upstreams) == 0 || upstreams[0] == "" {
		log.Fatal("no upstreams specified")
	}

	return Args{
		upstreams: upstreams,
		port:      *port,
	}
}

type ReverseProxyHandler struct {
}

func (r *ReverseProxyHandler) ServerHTTP(req server.Request, write server.Write) {}

func (r *ReverseProxyHandler) ServerReverseProxy(req server.Request, conn net.Conn, upstreamConn net.Conn) {
	removeHopByHopHeaders(req.Header)
	clientIP, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	if prior := req.Header.Get("X-Forwarded-For"); prior != "" {
		clientIP = prior + ", " + clientIP
	}
	req.Header.Set("X-Forwarded-For", clientIP)
	req.Header.Set("Via", "1.1 reverse-proxy")
	req.Write(upstreamConn)

	done := make(chan struct{}, 2)
	go func() {
		io.Copy(upstreamConn, req.Body)
		upstreamConn.(*net.TCPConn).CloseWrite()
		done <- struct{}{}
	}()

	go func() {
		io.Copy(conn, upstreamConn)
		done <- struct{}{}
	}()
	<-done
	<-done
}

var hopByHopHeaders = []string{
	"Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization",
	"TE", "Trailers", "Transfer-Encoding", "Upgrade",
}

func removeHopByHopHeaders(header http.Header) {
	connection := header.Get("Connection")
	for key := range strings.SplitSeq(connection, ",") {
		header.Del(strings.TrimSpace(key))
	}

	for _, key := range hopByHopHeaders {
		header.Del(key)
	}
}
