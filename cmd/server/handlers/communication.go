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
	"netmsys/pkg/nettsk"
	"strings"
)

func (s *Server) ListenAgents() {
	dataChannel := make(chan []byte)
	errorChannel := make(chan error)

	// Start receiving data on UDP port with Receive function
	go func() {
		for {
			nettsk.Receive(s.UDPPort, dataChannel, errorChannel)
		}
	}()

	fmt.Println("Server is listening for agent messages on UDP port", s.UDPPort)

	for {
		select {
		case data := <-dataChannel:
			// Process received data
			s.handleAgentMessage(data)
		case err := <-errorChannel:
			// Handle any errors reported by Receive
			fmt.Println("Error receiving data:", err)
		}
	}
}

// handleAgentMessage checks if a message is a wake-up signal and registers the agent.
func (s *Server) handleAgentMessage(data []byte) {
	message := string(data)
	if strings.HasPrefix(message, "REGISTER|") {
		// Extract client ID from message after "WAKE_UP|"
		clientID := strings.TrimPrefix(message, "REGISTER|")
		s.registerAgent(clientID)
	}
}

// registerAgent adds a new agent to the list of agent IDs if not already registered.
func (s *Server) registerAgent(agentID string) {
	for _, id := range s.AgentIDs {
		if id == agentID {
			fmt.Println("Agent", agentID, "is already registered.")
			return
		}
	}
	s.AgentIDs = append(s.AgentIDs, agentID)
	fmt.Println("Registered new agent:", agentID)
}

func (s *Server) sendTasks() {

}