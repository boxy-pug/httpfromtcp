package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {

	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Printf("error listening to port: %v", err)
		return
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("error accepting connection: %v", err)
			continue
		}
		if conn != nil {
			fmt.Printf("Connection established\n")
		}
		ch := getLinesChannel(conn)
		for line := range ch {
			fmt.Printf("%s\n", line)
		}
		fmt.Printf("Connection has been closed\n")
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer f.Close()
		defer close(ch)

		buffer := make([]byte, 8)
		str := ""

		for {
			n, err := f.Read(buffer)
			if err != nil {
				if err == io.EOF {
					if str != "" {
						ch <- str
					}
					break
				}
				log.Fatalf("error reading file: %v", err)
			}

			str += string(buffer[:n])

			for {
				idx := strings.Index(str, "\n")
				if idx == -1 {
					break
				}
				ch <- str[:idx]
				str = str[idx+1:]
			}

		}

	}()
	return ch

}
