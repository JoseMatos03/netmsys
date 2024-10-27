package nettsk

import (
	"fmt"
	"net"
	"os"
)

// Receive function listens on a given UDP port and prints received messages
func Receive(port string) {
	// Resolve the UDP address to listen on
	localAddr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		fmt.Println("Error resolving address:", err)
		os.Exit(1)
	}

	// Create a UDP socket to listen
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		fmt.Println("Error setting up UDP listener:", err)
		os.Exit(1)
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	// Read incoming messages
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			continue
		}

		fmt.Printf("Received message from %s: %s\n", addr.String(), string(buffer[:n]))
	}
}
