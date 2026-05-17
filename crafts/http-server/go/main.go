package main

import (
	"fmt"
	"io"
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
			fmt.Printf("accept error: %v\n", err)
			continue
		}
		fmt.Printf("conn: %v\n", conn)

		handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	// addの取得
	add := conn.RemoteAddr()
	fmt.Printf("add: %v\n", add)

	// 空のメモリを用意
	buf := make([]byte, 4096)

	for {
		n, err := conn.Read(buf)

		// 送られている場合は、表示
		if n > 0 {
			fmt.Printf("n: %v\n", buf[:n])
		}

		// 接続終了の場合は表示して、break
		if err == io.EOF {
			fmt.Printf("connect close add: %v\n", add)
			break
		}

		// errがある場合は break
		if err != nil {
			fmt.Printf("add: %v, read err: %v\n", add, err)
			break
		}
	}
}
