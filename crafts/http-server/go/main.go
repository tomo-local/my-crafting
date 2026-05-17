package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()

	fmt.Println("Listening on :8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("conn: %v\n", conn)
	}
}
