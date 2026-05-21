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
	if len(fields) < 2 {
		return Request{}, fmt.Errorf("invalid request line: %q", line)
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
		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			parts := strings.SplitN(line, ":", 2)
			contentLength, err = strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return Request{}, fmt.Errorf("invalid Content-Length: %w", err)
			}
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
