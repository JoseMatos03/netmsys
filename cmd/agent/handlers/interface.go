// ------------------------------------ LICENSE ------------------------------------
//
// Copyright 2024 Ana Pires, José Matos, Rúben Oliveira
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// ---------------------------------------------------------------------------------

package handlers

import (
	"fmt"
	"netmsys/pkg/alrtflw"
	"netmsys/pkg/nettsk"
	"os"
)

// Agent represents the agent's information
type Agent struct {
	ID         string
	IPAddr     string
	TaskID     string // Holds the current task ID if any, empty if no task assigned
	UDPPort    string
	TCPPort    string
	ServerAddr string // Only receive packets from this server IP address
}

// Create a new agent with the given ID and IP address
func NewAgent(id, ip, serverAddr, udpPort, tcpPort string) *Agent {
	return &Agent{
		ID:         id,
		IPAddr:     ip,
		TaskID:     "",
		ServerAddr: serverAddr,
		UDPPort:    udpPort,
		TCPPort:    tcpPort,
	}
}

func Start(args []string) {
	// Check if enough arguments are provided
	if len(args) < 4 {
		fmt.Println("Usage: ./agent <Agent ID> <Server IP> <UDP port> <TCP port>")
		return
	}

	// Parse arguments
	agentID := args[1]
	serverAddr := args[2]
	udpPort := args[3]
	tcpPort := args[4]

	// Automatically get the local IP address of the agent
	agentIP, err := GetLocalIP()
	if err != nil {
		fmt.Println("Error getting local IP address:", err)
		os.Exit(1)
	}

	// Create a new agent with the provided ID, local IP address, and server address
	agent := NewAgent(agentID, agentIP, serverAddr, udpPort, tcpPort)

	// Display agent information
	fmt.Printf("Agent %s (IP: %s) is ready.\n", agent.ID, agent.IPAddr)
	fmt.Printf("Listening for UDP on port %s and TCP on port %s. Accepting packets only from server at %s.\n", agent.UDPPort, agent.TCPPort, agent.ServerAddr)

	// Start listening for UDP messages (Nettsk protocol)
	go func() {
		fmt.Println("Agent is listening for UDP messages on port", agent.UDPPort)
		nettsk.Receive(agent.UDPPort) // Receive UDP messages only from the server
	}()

	// Start listening for TCP messages (AlertFlow protocol)
	go func() {
		fmt.Println("Agent is listening for TCP messages on port", agent.TCPPort)
		alrtflw.Receive(agent.TCPPort) // Receive TCP messages only from the server
	}()

	// Block the agent to keep it running indefinitely
	select {} // Infinite loop, agent runs indefinitely
}
