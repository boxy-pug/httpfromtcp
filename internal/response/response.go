package response

import (
	"fmt"
	"strconv"

	"github.com/boxy-pug/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK            StatusCode = 200
	BadRequest    StatusCode = 400
	InternalError StatusCode = 500
)

type Writer struct {
	Headers    headers.Headers
	StatusCode StatusCode
	StatusLine []byte
	Body       []byte
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	switch statusCode {
	case OK:
		w.StatusLine = []byte("HTTP/1.1 200 OK\r\n")
	case BadRequest:
		w.StatusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case InternalError:
		w.StatusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default:
		w.StatusLine = []byte("HTTP/1.1 XXX \r\n")
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	strLen := strconv.Itoa(contentLen)
	return headers.Headers{
		"Content-Length": strLen,
		"Connection":     "close",
		"Content-Type":   "text/plain",
	}
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	w.Headers = headers
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	w.Body = p
	// Implementation here
	return len(p), nil
}

func (w *Writer) AssembleResponse() []byte {
	var resp []byte

	// write statusline
	resp = append(resp, w.StatusLine...)

	// write headers
	for key, val := range w.Headers {
		headerLine := fmt.Sprintf("%s: %s\r\n", key, val)
		resp = append(resp, []byte(headerLine)...)
	}
	resp = append(resp, []byte("\r\n")...)

	// write body
	resp = append(resp, w.Body...)

	return resp

}
