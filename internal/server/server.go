package server

import (
	"fmt"
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

type Handler func(w *response.Writer, req *request.Request)

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

	// var buf bytes.Buffer
	writer := &response.Writer{}

	s.Handler(writer, req)

	// Write statusline
	conn.Write(writer.AssembleResponse())

}
