package main

import (
	"fmt"
	"netmsys/pkg/nettsk"
	"netmsys/tools/parsers"
	"os"
)

func main() {
	// Example usage, checking for JSON file as argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: server <json-file>")
		os.Exit(1)
	}

	jsonFile := os.Args[1]

	// Use the parser to read and parse the JSON file
	taskData, err := parsers.ReadAndParseJSON(jsonFile)
	if err != nil {
		fmt.Println("Error reading or parsing JSON:", err)
		os.Exit(1)
	}

	// For now, just print the parsed data
	fmt.Printf("Parsed Task Data: %+v\n", taskData)

	// Open the connection for receiving packets
	fmt.Println("Server is waiting for messages...")
	nettsk.Receive("8080")
}
