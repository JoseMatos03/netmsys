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
	"os"
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
		fmt.Printf("communication.SendTask(): Task with ID %s not found.\n", taskID)
		return
	}

	// Send task to each target
	for _, targetID := range task.Targets {
		targetAddr, exists := s.Agents[targetID]
		if !exists {
			fmt.Printf("communication.SendTask(): Target %s is not registered.\n", targetID)
			continue
		}

		serializedTask, _ := parsers.SerializeJSON(task)
		message := "TASK|" + string(serializedTask)
		go func() {
			err := nettsk.Send(targetAddr, "8080", []byte(message))
			if err != nil {
				fmt.Printf("communication.SendTask(): failed to send task.")
				return
			}
			fmt.Printf("Task sent successfully.")
		}()
	}
}

func (s *Server) ListenAgents() {
	// UDP Channels
	dataChannel := make(chan []byte)
	errorChannel := make(chan error)

	// Start receiving data on UDP port with Receive function
	go func() {
		for {
			nettsk.Receive(s.UDPPort, dataChannel, errorChannel)
		}
	}()

	fmt.Printf("Listening for agent messages on UDP port: %s\n", s.UDPPort)

	for {
		select {
		case data := <-dataChannel:
			// Process received data
			err := s.handleAgentMessage(data)
			if err != nil {
				fmt.Printf("communication.ListenAgents(): Error handling message.\n")
			}
		case err := <-errorChannel:
			// Handle any errors reported by Receive
			fmt.Printf("communication.ListenAgents(): Error receiving data: %s\n", err)
		}
	}
}

// handleAgentMessage checks if a message is a wake-up signal and registers the agent.
func (s *Server) handleAgentMessage(data []byte) error {
	message := string(data)
	if strings.HasPrefix(message, "REGISTER|") {
		registrationData := strings.TrimPrefix(message, "REGISTER|")
		parts := strings.Split(registrationData, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid registration message format: %s", message)
		}

		clientID := parts[0]
		clientIP := parts[1]

		// Register agent with parsed ID and IP address
		err := s.registerAgent(clientID, clientIP)
		if err != nil {
			return err
		}
	}
	if strings.HasPrefix(message, "OUTPUT|") {
		// Extract the output data
		outputData := strings.TrimPrefix(message, "OUTPUT|")

		// Call the logOutput function to handle writing to the log file
		err := s.logOutput(outputData)
		if err != nil {
			return err
		}
	}

	// Other cases will go here after...
	return nil
}

// registerAgent adds the agent to the map of AgentIDs if not already registered.
func (s *Server) registerAgent(agentID, ipAddr string) error {
	// Check if agent is already registered
	if _, exists := s.Agents[agentID]; exists {
		return fmt.Errorf("agent %s is already registered", agentID)
	}

	// Register the agent
	s.Agents[agentID] = ipAddr
	fmt.Printf("Registered new agent: ID=%s, IP=%s\n", agentID, ipAddr)

	// Loop through tasks and send those that target this agent
	for _, task := range s.Tasks {
		// Check if this task targets the newly registered agent
		for _, targetID := range task.Targets {
			if targetID == agentID {
				s.SendTask(task.TaskID)
				break
			}
		}
	}
	return nil
}

func (s *Server) logOutput(outputData string) error {
	// Define the output directory and log file path
	outputDir := "outputs"
	logFilePath := outputDir + "/output_log.txt" // TODO: Change to date or smth idk

	// Ensure the output directory exists
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Open the log file in append mode, create it if it doesn't exist
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open or create log file: %v", err)
	}
	defer file.Close()

	// Write the output data to the file
	_, err = file.WriteString(outputData + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to log file: %v", err)
	}

	return nil
}
