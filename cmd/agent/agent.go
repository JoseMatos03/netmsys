package main

import (
	"fmt"
	"netmsys/pkg/alrtflw"
	"netmsys/pkg/nettsk"
	"os"
)

func main() {
	// Get server address from command line argument
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./agent <UDP server address> <TCP server address>")
		return
	}

	udpServerAddr := os.Args[1]
	tcpServerAddr := os.Args[2]

	// Sending a message using the Nettsk (UDP) protocol
	fmt.Println("Agent is sending a UDP message to", udpServerAddr)
	nettsk.Send("Hello from Agent via UDP", udpServerAddr)

	// Sending a message using the AlertFlow (TCP) protocol
	fmt.Println("Agent is sending a TCP message to", tcpServerAddr)
	alrtflw.Send("Hello from Agent via TCP", tcpServerAddr)
}
