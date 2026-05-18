package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/tomo-local/http-server/internal/server"
)

func main() {
	// サーバーの立ち上げ
	srv := server.NewServer(":8080", handleConn)

	err := srv.Start()

	if err != nil {
		log.Fatal(err)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	// addの取得
	add := conn.RemoteAddr()
	fmt.Printf("add: %v\n", add)
	fmt.Println("====================")

	// 空のメモリを用意
	buf := make([]byte, 4096)

	for {
		n, err := conn.Read(buf)

		// 送られている場合は、表示
		if n > 0 {
			fmt.Println("request:")
			fmt.Printf("%v", string(buf[:n]))
			fmt.Println("====================")

			rl, err := parseRequest(buf)
			if err != nil {
				fmt.Printf("error: %v\r\n", err)
			}

			var status, body string
			switch rl.Path {
			case "/":
				status = "200 OK"
				body = "Welcome!"
			case "/about":
				status = "200 OK"
				body = "About Path"
			default:
				status = "400 Not Found"
				body = "Not Found"
			}

			body += "\r\n" +
				string(buf[:n])

			// 特定の形式に変更する
			response := "HTTP/1.1" + status + "\r\n" +
				"Content-Length: " + strconv.Itoa(len(body)) + "\r\n" +
				"\r\n" +
				body

			_, err = conn.Write([]byte(response))

			if err != nil {
				fmt.Printf("write err: %v\r\n", err)
				break
			}

		}

		// 接続終了の場合は表示して、break
		if err == io.EOF {
			fmt.Printf("connect close add: %v\n", add)
			fmt.Println("====================")
			break
		}

		// errがある場合は break
		if err != nil {
			fmt.Printf("add: %v, read err: %v\n", add, err)
			fmt.Println("====================")
			break
		}
	}
}

type Request struct {
	Method  string
	Path    string
	Version string
}

func parseRequest(buf []byte) (Request, error) {
	lines := strings.Split(string(buf), "\r\n")
	fields := strings.Fields(lines[0])
	//1行目は、 Method, Path, Http Versionの3種類
	if len(fields) != 3 {
		return Request{}, fmt.Errorf("invalid request line: %q", lines[0])
	}

	return Request{
		Method:  fields[0],
		Path:    fields[1],
		Version: fields[2],
	}, nil
}
