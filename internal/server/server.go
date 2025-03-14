package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/boxy-pug/httpfromtcp/internal/request"
	"github.com/boxy-pug/httpfromtcp/internal/response"
)

// Contains the state of the server
type Server struct {
	State    bool
	Listener net.Listener
	Handler  Handler
}

type HandlerError struct {
	Message    string
	StatusCode response.StatusCode
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

// Creates a net.Listener and returns a new Server instance. Starts listening for requests inside a goroutine.
func Serve(port int, h Handler) (*Server, error) {

	s := &Server{}

	portString := fmt.Sprintf("127.0.0.1:%d", port)

	l, err := net.Listen("tcp", portString)
	if err != nil {
		return nil, err
	}

	s.Listener = l
	s.State = true
	s.Handler = h

	go s.listen()

	return s, nil
}

// Closes the listener and the server
func (s *Server) Close() error {
	// Mark the server as not running
	s.State = false

	// Close the listener
	if err := s.Listener.Close(); err != nil {
		return err
	}

	return nil
}

// Uses a loop to .Accept new connections as they come in, and handles each one in a new goroutine.
// I used an atomic.Bool to track whether the server is closed or not so that I can ignore connection errors after the server is closed.
func (s *Server) listen() {
	for s.State {
		conn, err := s.Listener.Accept()
		if err != nil {
			if !s.State {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn) // Handle each connection in a new goroutine
	}

}

// Handles a single connection by writing the following response and then closing the connection:
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Fprint(conn, "HTTP/1.1 400 Bad Request\r\n\r\nMalformed request")
		return
	}

	var buf bytes.Buffer

	handlerErr := s.Handler(&buf, req)
	if handlerErr != nil {
		fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\n\r\n%s", handlerErr.StatusCode, "Error", handlerErr.Message)
		return
	}

	err = response.WriteStatusLine(conn, response.OK)
	if err != nil {
		fmt.Println(err)
	}

	defaultHeaders := response.GetDefaultHeaders(buf.Len())

	// Write the HTTP response
	err = response.WriteHeaders(conn, defaultHeaders)
	if err != nil {
		log.Printf("Error writing to connection: %v", err)
		return
	}
	// Write the response body from the buffer
	_, err = buf.WriteTo(conn) // Write the buffer's contents to the connection
	if err != nil {
		fmt.Printf("Error writing response body: %v\n", err)
		return
	}

}

/*
	response :=
	//old def resp: "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 13\r\n\r\nHello World!\n"
	_, err := conn.Write([]byte(response))

	}*/
