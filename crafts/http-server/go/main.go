package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	// サーバーの立ち上げ
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Listening on :8080")

	for {
		// 接続確率
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("conn: %v\n", conn)

		add := conn.RemoteAddr()
		fmt.Printf("add: %v\n", add)
	}
}
