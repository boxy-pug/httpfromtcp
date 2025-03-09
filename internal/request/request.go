package request

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	b, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}

	requestLine, err := parseRequestLine(b)
	if err != nil {
		return nil, errors.New("error parsing request line")
	}

	req := Request{
		RequestLine: *requestLine,
	}

	return &req, nil
}

func parseRequestLine(b []byte) (*RequestLine, error) {
	requestString := strings.Split(string(b), "\r\n")[0]

	requestList := strings.Split(requestString, " ")

	if len(requestList) != 3 {
		return nil, errors.New("missing args in request line")
	}

	requestLine, err := parseRequestLineElems(requestList)
	if err != nil {
		return nil, errors.New("could not parse req line elements")
	}

	return requestLine, nil
}

func parseRequestLineElems(rl []string) (*RequestLine, error) {
	httpVer := rl[2]
	httpVerNum := strings.Split(httpVer, "/")[1]
	reqTarget := rl[1]
	method := rl[0]

	fmt.Printf("httpvers: %s\nreqtarget: %s\nmethod: %s\n", httpVer, reqTarget, method)

	// Checking that method name is all uppercase
	for _, ch := range strings.TrimSpace(method) {
		if !(ch >= 'A' && ch <= 'Z') {
			return nil, errors.New("method is not all uppercase letters")
		}
	}

	// Checking that http version is "HTTP/1.1"
	if httpVer != "HTTP/1.1" {
		return nil, errors.New("wrong http version")
	}

	return &RequestLine{
		HttpVersion:   httpVerNum,
		RequestTarget: reqTarget,
		Method:        method,
	}, nil

}
