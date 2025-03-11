package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/boxy-pug/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	state       int
	Headers     headers.Headers
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	stateInitialized = iota
	stateDone
	requestStateParsingHeaders
)

const bufferSize = 8 // Start with a small buffer to test chunking

func RequestFromReader(reader io.Reader) (*Request, error) {
	// Create a buffer to read data into
	buf := make([]byte, bufferSize)

	// Create a new Request with initialized state
	req := &Request{
		state: stateInitialized,
	}

	readToIndex := 0 // Track how much data we've read into our buffer

	// Continue reading and parsing until we're done
	for req.state != stateDone {
		// If buffer is full, grow it
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		// Read the next chunk of data
		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				// If we've reached EOF without completing the request, it's an error
				if req.state != stateDone {
					return nil, errors.New("incomplete request: reached EOF")
				}
				break
			}
			return nil, fmt.Errorf("error reading from reader: %v", err)
		}

		// Update how much data we've read
		readToIndex += n

		// Try to parse the data we have so far
		bytesConsumed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		// If we consumed bytes, we need to shift the buffer
		if bytesConsumed > 0 {
			// Move the remaining bytes to the beginning of the buffer
			copy(buf, buf[bytesConsumed:readToIndex])
			readToIndex -= bytesConsumed
		}
	}

	return req, nil
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {

	// Look for \r\n in the input bytes
	requestLineBytes := bytes.Index(b, []byte("\r\n"))

	// If \r\n not found, we need more data
	if requestLineBytes == -1 {
		return nil, 0, nil
	}

	// Include the \r\n in our byte count
	bytesConsumed := requestLineBytes + 2

	requestString := strings.Split(string(b), "\r\n")

	if len(requestString) == 1 {
		return nil, 0, nil
	}

	requestList := strings.Split(requestString[0], " ")

	if len(requestList) != 3 {
		return nil, 0, errors.New("missing args in request line")
	}

	requestLine, err := parseRequestLineElems(requestList)
	if err != nil {
		return nil, 0, errors.New("could not parse req line elements")
	}

	return requestLine, bytesConsumed, nil
}

func parseRequestLineElems(rl []string) (*RequestLine, error) {
	httpVer := rl[2]
	httpVerNum := strings.Split(httpVer, "/")[1]
	reqTarget := rl[1]
	method := rl[0]

	//fmt.Printf("httpvers: %s\nreqtarget: %s\nmethod: %s\n", httpVer, reqTarget, method)

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

// The parse method processes a chunk of data
func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != stateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}
		if n == 0 {
			// no more data and no error means were done
			break
		}
		totalBytesParsed += n
		// ...
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case stateInitialized:
		// Try to parse the request line from the current chunk
		requestLine, bytesConsumed, err := parseRequestLine(data)

		if err != nil {
			return 0, err
		}

		if bytesConsumed == 0 {
			// Not enough data yet, need more
			return 0, nil
		}

		// Successfully parsed the request line
		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return bytesConsumed, nil

	case requestStateParsingHeaders:

		// Create headers if they don't exist yet
		if r.Headers == nil {
			r.Headers = headers.NewHeaders()
		}

		// Parse using the existing headers object
		bytesConsumed, done, err := r.Headers.Parse(data)
		// fmt.Printf("Parse result: bytes=%d, done=%v, err=%v\n", bytesConsumed, done, err)
		if err != nil {
			return 0, err
		}

		// If we're done parsing headers, move to the next state
		if done {
			r.state = stateDone
		}

		return bytesConsumed, nil

	case stateDone:
		return 0, errors.New("error: trying to read data in a done state")

	default:
		return 0, errors.New("error: unknown state")
	}
}
