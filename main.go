package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {

	file, err := os.Open("messages.txt")
	if err != nil {
		fmt.Errorf("could not open file: %w", err)
	}

	defer file.Close()

	ch := getLinesChannel(file)

	for line := range ch {
		fmt.Printf("read: %s\n", line)
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
