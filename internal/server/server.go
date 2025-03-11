package server

import (
	"fmt"
	"log"
	"net"
)

// Contains the state of the server
type Server struct {
	State    bool
	Listener net.Listener
}

// Creates a net.Listener and returns a new Server instance. Starts listening for requests inside a goroutine.
func Serve(port int) (*Server, error) {

	s := &Server{}

	portString := fmt.Sprintf(":%d", port)

	l, err := net.Listen("tcp", portString)
	if err != nil {
		return nil, err
	}

	s.Listener = l
	s.State = true

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

	// Write the HTTP response
	response := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 13\r\n\r\nHello World!\n"
	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error writing to connection: %v", err)
		return
	}
}
