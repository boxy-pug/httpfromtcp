package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/boxy-pug/httpfromtcp/internal/request"
	"github.com/boxy-pug/httpfromtcp/internal/server"
)

const port = 42069

func main() {

	// Instantiate your handler function
	myHandler := func(w io.Writer, req *request.Request) *server.HandlerError {
		// Check request path and handle appropriately
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return &server.HandlerError{StatusCode: 400, Message: "Your problem is not my problem\n"}
		case "/myproblem":
			return &server.HandlerError{StatusCode: 500, Message: "Woopsie, my bad\n"}
		default:
			// Write a default response to the writer
			fmt.Fprint(w, "All good, frfr\n")
			return nil
		}
	}

	// Create the server with the custom handler
	s := &server.Server{
		Handler: myHandler,
	}

	httpServer, err := server.Serve(port, s.Handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer httpServer.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
