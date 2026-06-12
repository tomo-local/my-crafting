package server

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Request struct {
	Method        string
	Path          string
	Version       string
	Host          string
	ContentLength int
	Connection    string
	Header        http.Header
	Body          io.Reader
}

func NewRequest(r io.Reader) (Request, error) {
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
	host := ""
	headers := make(http.Header)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return Request{}, err
		}
		if line == "\r\n" {
			break
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers.Set(name, value)
		}

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
		case strings.HasPrefix(lower, "host:"):
			parts := strings.SplitN(line, ":", 2)
			host = strings.TrimSpace(parts[1])
		}

	}

	return Request{
		Method:        fields[0],
		Path:          fields[1],
		Version:       fields[2],
		Host:          host,
		ContentLength: contentLength,
		Connection:    connection,
		Header:        headers,
		Body:          reader,
	}, nil
}

func (r Request) WantsKeepAlive() bool {
	if r.Version == "HTTP/1.1" {
		return r.Connection != "close"
	}
	if r.Version == "HTTP/1.0" {
		return r.Connection == "keep-alive"
	}
	return false
}

func (r Request) Write(w io.Writer) error {
	fmt.Fprintf(w, "%s %s %s\r\n", r.Method, r.Path, r.Version)
	if err := r.Header.Write(w); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "\r\n")
	return err
}
