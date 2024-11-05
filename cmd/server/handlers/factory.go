package handlers

import (
	"fmt"
	"netmsys/cmd/message"
)

type Server struct {
	UDPPort  string
	TCPPort  string
	AgentIDs []string
	Tasks    []message.Task
}

func NewServer(args []string) (*Server, error) {
	// Check if enough arguments are provided
	if len(args) < 2 {
		return nil, fmt.Errorf("insufficient arguments: ./server <UDP port> <TCP port>")
	}

	serverUDP := args[1]
	serverTCP := args[2]
	return &Server{
		UDPPort:  serverUDP,
		TCPPort:  serverTCP,
		AgentIDs: make([]string, 10),
		Tasks:    make([]message.Task, 10),
	}, nil
}
