package alrtflw

import (
	"fmt"
	"io"
	"net"
)

// Receive listens on the given TCP address and receives messages
func Receive(port string) {
	// Listen for incoming TCP connections
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error starting TCP listener:", err)
		return
	}
	defer listener.Close() // Only defer listener.Close(), not individual connections

	for {
		// Accept a connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle the connection (read the message)
		handleConnection(conn)
	}
}

// handleConnection processes an incoming TCP connection
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read all data from the connection until the client closes it
	data, err := io.ReadAll(conn)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	// Display the received message
	fmt.Println("Received TCP message:", string(data))
}
