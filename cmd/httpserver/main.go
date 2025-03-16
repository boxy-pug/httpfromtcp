package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
			if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
				path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
				targetURL := fmt.Sprintf("https://httpbin.org/%s", path)

				proxyReq, err := http.NewRequest("GET", targetURL, nil)
				if err != nil {
					fmt.Println("Error making request to httpbin: ", err)
					return
				}
				// Make the proxy request
				resp, err := http.DefaultClient.Do(proxyReq)
				if err != nil {
					fmt.Println("Error making request to httpbin:", err)
					return
				}
				defer resp.Body.Close()

				w.Headers = headers.NewHeaders()
				w.WriteStatusLine(response.OK)

				w.Headers.Remove("Content-Length")

				w.WriteHeaders(headers.Headers{
					"Transfer-Encoding": "chunked",          // Clearly states chunks are coming
					"Content-Type":      "application/json", // Expected from httpbin
					"Trailer":           "X-Content-Sha256, X-Content-Length",
				})

				var fullBody []byte
				buf := make([]byte, 100) // Adjust buffer size if needed
				for {
					n, err := resp.Body.Read(buf)
					if n > 0 {
						fullBody = append(fullBody, buf[:n]...)
						fmt.Printf("Chunk to write: %q, size: %d\n", buf[:n], n)
						if _, writeErr := w.WriteChunkedBody(buf[:n]); writeErr != nil {
							fmt.Println("Error writing chunk:", writeErr)
							return
						}
					}
					if err != nil {
						if err == io.EOF {
							break // End of response body
						}
						fmt.Println("Error reading response body:", err)
						return
					}
				}
				// Signal end of chunked response
				if _, err := w.WriteChunkedBodyDone(); err != nil {
					fmt.Println("Error finishing chunked response:", err)
				}

				// Calculate SHA256 hash and content length
				hash := sha256.Sum256(fullBody)
				hashString := hex.EncodeToString(hash[:])
				contentLength := strconv.Itoa(len(fullBody))

				// Write trailers
				w.WriteTrailers(headers.Headers{
					"X-Content-Sha256": hashString,
					"X-Content-Length": contentLength,
				})

			} else {
				// Write a default response to the writer
				w.WriteStatusLine(response.BadRequest)
				return
			}

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
