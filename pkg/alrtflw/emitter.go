package alrtflw

import (
	"fmt"
	"net"
)

// Send sends a message to the given TCP server address
func Send(message string, serverAddr string) {
	// Establish a TCP connection
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Send the message
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}

	fmt.Println("Message sent:", message)
}
