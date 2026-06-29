package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"pubsub/broker"
	"strings"
)

type Args struct {
	port string
}

func NewArgs() Args {
	port := flag.String("port", "8080", "server port")
	flag.Parse()
	return Args{
		port: *port,
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	args := NewArgs()
	ln, err := net.Listen("tcp", ":"+args.port)
	if err != nil {
		err = fmt.Errorf("failed to listen on tcp port %s: %w", args.port, err)
		logger.Error("failed to start server", "error", err)
		return
	}
	defer ln.Close()

	logger.Info("Server started", "port", args.port)

	b := broker.NewBroker()

	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.Error("failed to accept connection", "error", err)
			continue
		}
		logger.Info("connection accepted", "addr", conn.RemoteAddr())

		go handleClient(conn, b)
	}
}

func handleClient(conn net.Conn, b *broker.Broker) {
	defer conn.Close()
	sub := broker.NewSubscriber(conn)
	readLoop(sub, b)
	b.UnsubscribeAll(sub)
}

func readLoop(sub *broker.Subscriber, b *broker.Broker) {
	scanner := bufio.NewScanner(sub.Conn())

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}

		cmd := strings.ToUpper(fields[0])

		switch cmd {
		case "SUB":
			if len(fields) < 2 {
				fmt.Fprintf(sub.Conn(), "-ERR usage: SUB <topic>\r\n")
				continue
			}
			topic := fields[1]
			ch := b.Subscribe(topic, sub)
			go writeLoop(topic, ch, sub.Conn())
			fmt.Fprintf(sub.Conn(), "+OK\r\n")

		case "UNSUB":
			topic := fields[1]
			b.Unsubscribe(topic, sub)
			fmt.Fprintf(sub.Conn(), "+OK\r\n")

		case "PUB":
			if len(fields) < 3 {
				fmt.Fprintf(sub.Conn(), "-ERR usage: PUB <topic> <message>\r\n")
				continue
			}
			topic := fields[1]
			message := strings.Join(fields[2:], " ")
			b.Publish(topic, message)
			fmt.Fprintf(sub.Conn(), "+OK\r\n")

		default:
			fmt.Fprintf(sub.Conn(), "-ERR unknown command\r\n")
		}
	}
}

func writeLoop(topic string, ch chan string, conn net.Conn) {
	for msg := range ch {
		fmt.Fprintf(conn, "MSG %s %s\r\n", topic, msg)
	}
}
