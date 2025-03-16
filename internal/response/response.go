package response

import (
	"fmt"
	"net/http"
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
	httpWriter http.ResponseWriter
}

func NewWriter(httpWriter http.ResponseWriter) *Writer {
	return &Writer{
		httpWriter: httpWriter,
		Headers:    make(headers.Headers),
		// Initialize any other fields you need
	}
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

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	// Create the chunked format
	chunkSize := fmt.Sprintf("%x", len(p))
	// log.Printf("Chunk written to httpWriter: Size: %s, Data: %q\n", chunkSize, p)
	chunk := []byte(chunkSize + "\r\n")
	chunk = append(chunk, p...)
	chunk = append(chunk, []byte("\r\n")...)

	// If httpWriter is nil, just store the data for later use in AssembleResponse
	if w.httpWriter == nil {
		w.Body = append(w.Body, chunk...)
		return len(p), nil
	}

	// Otherwise write directly to the httpWriter
	return w.httpWriter.Write(chunk)
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	// Add the terminating chunk: 0 + CRLF + CRLF
	w.Body = append(w.Body, []byte("0\r\n")...)

	// terminatingChunk := []byte("0\r\n\r\n")

	//n, err := w.httpWriter.Write(terminatingChunk)
	//if err != nil {
	//	return n, err
	//}

	// Make sure to set the Transfer-Encoding header
	w.Headers["Transfer-Encoding"] = "chunked"

	// Remove Content-Length if it exists, as they shouldn't be used together
	delete(w.Headers, "Content-Length")

	return 0, nil
}

func (w *Writer) WriteTrailers(trailerHeaders headers.Headers) error {
	var trailerData []byte
	for key, val := range trailerHeaders {
		trailerLine := fmt.Sprintf("%s: %s\r\n", key, val)
		trailerData = append(trailerData, []byte(trailerLine)...)
	}
	trailerData = append(trailerData, []byte("\r\n")...)

	if w.httpWriter == nil {
		w.Body = append(w.Body, trailerData...)
		return nil
	}

	_, err := w.httpWriter.Write(trailerData)
	return err
}
