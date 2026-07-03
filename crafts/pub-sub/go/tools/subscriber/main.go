package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"
)

type Args struct {
	port  string
	topic string
	delay time.Duration
}

func NewArgs() Args {
	port := flag.String("port", "8080", "server port")
	topic := flag.String("topic", "", "subscribe topic name")
	delay := flag.Duration("delay", 0, "processing delay per message (simulate slow subscriber)")
	flag.Parse()

	return Args{
		port:  *port,
		topic: *topic,
		delay: *delay,
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	args := NewArgs()
	if args.topic == "" {
		logger.Error("-topic is required")
		os.Exit(1)
	}

	conn, err := net.Dial("tcp", "localhost:"+args.port)
	if err != nil {
		logger.Error("failed to connect", "port", args.port, "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	logger.Info("connected", "port", args.port)

	fmt.Fprintf(conn, "SUB %s\r\n", args.topic)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "+OK"):
			logger.Info("subscribed", "topic", args.topic)
		case strings.HasPrefix(line, "-ERR"):
			logger.Error("server error", "msg", line)
		case strings.HasPrefix(line, "MSG "):
			parts := strings.SplitN(line, " ", 3)
			if len(parts) >= 3 {
				fmt.Printf("[%s] %s\n", parts[1], parts[2])
			}
			if args.delay > 0 {
				time.Sleep(args.delay)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Error("read error", "error", err)
	}
}
