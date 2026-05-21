package request

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Request struct {
	Method        string
	Path          string
	Version       string
	ContentLength int
	Connection    string
	Body          io.Reader
}

func Parse(r io.Reader) (Request, error) {
	reader, ok := r.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return Request{}, err
	}

	fields := strings.Fields(strings.TrimRight(line, "\r\n"))
	if len(fields) != 3 {
		return Request{}, fmt.Errorf("invalid request line: %q", line)
	}

	contentLength := 0
	connection := ""
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return Request{}, err
		}
		if line == "\r\n" {
			break
		} // 空行でヘッダー終端
		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(lower, "content-length:"):
			parts := strings.SplitN(line, ":", 2)
			contentLength, err = strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return Request{}, fmt.Errorf("invalid Content-Length: %w", err)
			}
		case strings.HasPrefix(lower, "connection:"):
			parts := strings.SplitN(line, ":", 2)
			connection = strings.ToLower(strings.TrimSpace(parts[1]))
		}
	}

	return Request{
		Method:        fields[0],
		Path:          fields[1],
		Version:       fields[2],
		ContentLength: contentLength,
		Connection:    connection,
		Body:          reader,
	}, nil
}

// HTTP/1.1はデフォルトでkeep-alive、HTTP/1.0は明示的なConnection: keep-aliveが必要
func (r Request) WantsKeepAlive() bool {
	if r.Version == "HTTP/1.1" {
		return r.Connection != "close"
	}
	if r.Version == "HTTP/1.0" {
		return r.Connection == "keep-alive"
	}
	return false
}
