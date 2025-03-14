package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/boxy-pug/httpfromtcp/internal/headers"
	"github.com/boxy-pug/httpfromtcp/internal/request"
	"github.com/boxy-pug/httpfromtcp/internal/response"
	"github.com/boxy-pug/httpfromtcp/internal/server"
)

const port = 42069

func main() {

	// Instantiate your handler function
	myHandler := func(w *response.Writer, req *request.Request) {
		// Check request path and handle appropriately
		switch req.RequestLine.RequestTarget {
		case "/":
			body := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
			w.WriteStatusLine(response.OK)
			w.WriteHeaders(headers.Headers{
				"Content-Type":   "text/html",
				"Content-Length": strconv.Itoa(len(body)),
			})
			w.WriteBody([]byte(body))

		case "/yourproblem":
			body := `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
			w.WriteStatusLine(response.BadRequest)
			w.WriteHeaders(headers.Headers{
				"Content-Type":   "text/html",
				"Content-Length": strconv.Itoa(len(body)),
			})
			w.WriteBody([]byte(body))

		case "/myproblem":
			body := `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
			w.WriteStatusLine(response.InternalError)
			w.WriteHeaders(headers.Headers{
				"Content-Type":   "text/html",
				"Content-Length": strconv.Itoa(len(body)),
			})
			w.WriteBody([]byte(body))

		default:
			// Write a default response to the writer
			w.WriteStatusLine(response.BadRequest)
			return
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
