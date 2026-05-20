package request

import (
	"fmt"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Version string
}

func Parse(buf []byte) (Request, error) {
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
