package request

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type Request struct {
	Method        string
	Path          string
	Version       string
	ContentLength int
	Body          io.Reader
}

func Parse(conn net.Conn) (Request, error) {
	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')
	if err != nil {
		return Request{}, err
	}

	fields := strings.Fields(strings.TrimRight(line, "\r\n"))
	if len(fields) < 2 {
		return Request{}, fmt.Errorf("invalid request line: %q", line[0])
	}

	contentLength := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return Request{}, err
		}
		if line == "\r\n" {
			break
		} // 空行でヘッダー終端
		if strings.HasPrefix(line, "Content-Length:") {
			parts := strings.SplitN(line, ":", 2)
			contentLength, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
		}
	}

	return Request{
		Method:        fields[0],
		Path:          fields[1],
		Version:       fields[2],
		ContentLength: contentLength,
		Body:          reader,
	}, nil
}
