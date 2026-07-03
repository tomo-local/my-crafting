package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
)

type Args struct {
	port  string
	topic string
	count int
}

func NewArgs() Args {
	port := flag.String("port", "8080", "server port")
	topic := flag.String("topic", "", "publish topic name")
	count := flag.Int("count", 0, "number of messages to publish (0 = read from stdin)")
	flag.Parse()

	return Args{
		port:  *port,
		topic: *topic,
		count: *count,
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

	serverReader := bufio.NewReader(conn)

	publish := func(msg string) bool {
		fmt.Fprintf(conn, "PUB %s %s\r\n", args.topic, msg)
		resp, err := serverReader.ReadString('\n')
		if err != nil {
			logger.Error("read response error", "error", err)
			return false
		}
		resp = strings.TrimRight(resp, "\r\n")
		if strings.HasPrefix(resp, "-ERR") {
			logger.Error("publish failed", "response", resp)
		}
		return true
	}

	if args.count > 0 {
		for i := range args.count {
			if !publish(fmt.Sprintf("message %d", i+1)) {
				return
			}
		}
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if !publish(line) {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Error("stdin read error", "error", err)
	}
}
