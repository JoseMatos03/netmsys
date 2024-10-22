package main

import (
	"fmt"
	"netmsys/pkg/alrtflw"
	"netmsys/pkg/nettsk"
	"os"
)

func main() {
	// Example usage, checking for JSON file as argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: server <json-file>")
		os.Exit(1)
	}

	// jsonFile := os.Args[1]

	// Start receiving UDP messages (Nettsk protocol)
	go func() {
		fmt.Println("Server is listening for UDP messages on 8080")
		nettsk.Receive("8080")
	}()

	// Start receiving TCP messages (AlertFlow protocol)
	go func() {
		fmt.Println("Server is listening for TCP messages on 9090")
		alrtflw.Receive("9090")
	}()

	// Keep the server running
	select {} // Block forever to keep both listeners alive
}
