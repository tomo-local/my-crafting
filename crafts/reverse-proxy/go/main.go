package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"reverse-proxy/server"
	"strings"
)

type Args struct {
	upstream string
	port     string
}

func main() {
	args := parseArgs()
	fmt.Printf("Args upstream:%s, port:%s\r\n", args.upstream, args.port)

	addr := args.port
	if !strings.HasPrefix(args.port, ":") {
		addr = ":" + args.port
	}

	srv := server.NewReverseProxyServer(addr, &ReverseProxyHandler{Upstream: args.upstream})

	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func parseArgs() Args {
	upstream := flag.String("upstream", "localhost:9001", "接続先のアドレス")
	port := flag.String("port", "8080", "サーバーのポート")
	flag.Parse()

	return Args{
		upstream: *upstream,
		port:     *port,
	}
}

type ReverseProxyHandler struct {
	Upstream string
}

func (r *ReverseProxyHandler) ServerHTTP(req server.Request, write server.Write) {}

func (r *ReverseProxyHandler) ServerReverseProxy(req server.Request, conn net.Conn) {
	upstreamConn, err := net.Dial("tcp", r.Upstream)
	if err != nil {
		slog.Error("Reverse Proxy Error")
		return
	}
	defer upstreamConn.Close()

	// カスタムした reqなのでSetで書き換わる
	req.Host = "upstream"
	req.Header.Set("Host", r.Upstream)

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
