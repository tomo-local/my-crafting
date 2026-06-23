package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	lb "reverse-proxy/balancer"
	"reverse-proxy/server"
	"strings"
	"time"
)

type Args struct {
	upstreams []string
	port      string
	interval  int64
	balancer  int64
	maxIdle   int
}

func parseArgs() Args {
	port := flag.String("port", "8080", "サーバーのポート")
	upstreamsFlag := flag.String("upstreams", "localhost:9001,localhost:9002", "接続先のアドレス")
	interval := flag.Int64("interval", 10, "ヘルスチェックのインターバル")
	b := flag.Int64("balancer", 0, "バランシングを使うのか(0:RoundRobin,1:LeastConn)")
	maxIdle := flag.Int("pool-size", 10, "")

	flag.Parse()

	upstreams := strings.Split(*upstreamsFlag, ",")

	if len(upstreams) == 0 || upstreams[0] == "" {
		log.Fatal("no upstreams specified")
	}

	return Args{
		port:      *port,
		upstreams: upstreams,
		interval:  *interval,
		balancer:  *b,
		maxIdle:   *maxIdle,
	}
}

func main() {
	args := parseArgs()
	fmt.Printf("Args upstreams:%s, port:%s\r\n", args.upstreams, args.port)

	addr := args.port
	if !strings.HasPrefix(args.port, ":") {
		addr = ":" + args.port
	}
	interval := time.Duration(args.interval) * time.Second
	balancer, err := lb.NewBalancer(lb.Kind(args.balancer), args.upstreams, interval, args.maxIdle)

	if err != nil {
		slog.Error("failed to create balancer", "err", err)
		os.Exit(1)
	}

	srv := server.NewReverseProxyServer(addr, &ReverseProxyHandler{}, balancer)

	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

type ReverseProxyHandler struct {
}

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
		io.Copy(upstreamConn, io.LimitReader(req.Body, int64(req.ContentLength)))
		done <- struct{}{}
	}()

	go func() {
		reader := bufio.NewReader(upstreamConn)

		statusLine, err := reader.ReadString('\n')
		if err != nil {
			done <- struct{}{}
			return
		}
		conn.Write([]byte(statusLine))

		contentLength := 0
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				done <- struct{}{}
				return
			}
			conn.Write([]byte(line))
			if line == "\r\n" {
				break
			}
			lower := strings.ToLower(line)
			if strings.HasPrefix(lower, "content-length:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					contentLength, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
				}
			}
		}

		io.Copy(conn, io.LimitReader(reader, int64(contentLength)))
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
