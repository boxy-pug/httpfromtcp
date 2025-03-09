package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		fmt.Printf("error resolving udp: %v\n", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Printf("error dialing udp: %v\n", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error reading: %v\n", err)
			return
		}

		_, err = conn.Write([]byte(input))
		if err != nil {
			fmt.Printf("error writing: %v\n", err)
			return
		}
	}

}
