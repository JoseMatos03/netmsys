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
	"netmsys/cmd/message"
	"netmsys/pkg/nettsk"
	"netmsys/tools/parsers"
	"strings"
)

func (s *Server) SendTask(taskID string) {
	// Find the task by ID
	var task *message.Task
	for i := range s.Tasks {
		if s.Tasks[i].TaskID == taskID {
			task = &s.Tasks[i]
			break
		}
	}

	// Check if the task was found
	if task == nil {
		fmt.Printf("Task with ID %s not found\n", taskID)
		return
	}

	// Send task to each target
	for _, targetID := range task.Targets {
		targetAddr, exists := s.Agents[targetID]
		if !exists {
			continue
		}

		serializedTask, _ := parsers.SerializeJSON(task)
		message := "TASK|" + string(serializedTask)
		go nettsk.Send(targetAddr, "8080", []byte(message))
	}
}

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
		registrationData := strings.TrimPrefix(message, "REGISTER|")
		parts := strings.Split(registrationData, ":")
		if len(parts) != 2 {
			fmt.Println("Invalid registration message format:", message)
			return
		}

		clientID := parts[0]
		clientIP := parts[1]

		// Register agent with parsed ID and IP address
		s.registerAgent(clientID, clientIP)
	}
}

// registerAgent adds the agent to the map of AgentIDs if not already registered.
func (s *Server) registerAgent(agentID, ipAddr string) {
	// Check if agent is already registered
	if _, exists := s.Agents[agentID]; exists {
		fmt.Println("Agent", agentID, "is already registered.")
		return
	}

	// Register the agent
	s.Agents[agentID] = ipAddr
	fmt.Printf("Registered new agent: ID=%s, IP=%s\n", agentID, ipAddr)

	// Loop through tasks and send those that target this agent
	for _, task := range s.Tasks {
		// Check if this task targets the newly registered agent
		for _, targetID := range task.Targets {
			if targetID == agentID {
				fmt.Printf("Sending task %s to newly registered agent %s\n", task.TaskID, agentID)
				s.SendTask(task.TaskID)
				break
			}
		}
	}
}
