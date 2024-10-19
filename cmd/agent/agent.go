package main

import (
	"fmt"
	"netmsys/pkg/nettsk"
	"os"
)

func main() {
	// Check if the server address argument is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: agent <server-address>")
		os.Exit(1)
	}

	// Get the server address from the arguments
	serverAddress := os.Args[1]

	fmt.Println("Agent is sending a message to", serverAddress)
	// Sending message to the server at the provided address
	nettsk.Send("Hello from Agent", serverAddress)
}
