package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/boxy-pug/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK            StatusCode = 200
	BadRequest    StatusCode = 400
	InternalError StatusCode = 500
)

// type Writer struct {}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case OK:
		w.Write([]byte("HTTP/1.1 200 OK\r\n"))
	case BadRequest:
		w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
	case InternalError:
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
	default:
		w.Write([]byte("HTTP/1.1 XXX \r\n"))
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

func WriteHeaders(w io.Writer, headers headers.Headers) error {

	if len(headers) == 0 {
		headers = GetDefaultHeaders(0)
	}

	for key, val := range headers {
		byteLine := fmt.Sprintf("%s: %s\r\n", key, val)
		w.Write([]byte(byteLine))
	}
	w.Write([]byte("\r\n"))

	return nil
}
