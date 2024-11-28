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
	"os/exec"
	"strings"
)

// Register sends a registration message to the server to inform it of the agent's presence.
//
// Behavior:
//   - Constructs a registration message in the format "REGISTER|<AgentID>:<AgentIP>".
//   - Sends the message to the server's UDP port using the NetTask protocol.
//
// Example:
//
//	If the agent's ID is "Agent1" and its IP is "192.168.1.10", the registration message will be:
//	"REGISTER|Agent1:192.168.1.10".
//
// Dependencies:
//   - Requires the "nettsk" package for sending UDP messages.
func (a *Agent) Register() {
	registerMessage := "REGISTER|" + a.ID + ":" + a.IPAddr
	nettsk.Send(a.ServerAddr, a.UDPPort, []byte(registerMessage))
}

// ListenServer listens for incoming messages from the server on the agent's UDP port.
//
// Behavior:
//   - Continuously listens for data on the agent's UDP port using the NetTask protocol.
//   - Uses separate channels to handle incoming data and errors.
//   - Processes received messages by passing them to `handleServerMessage`.
//   - Logs errors if there are issues receiving data.
//
// Example:
//
//	The agent listens for server messages on the configured UDP port, processes tasks,
//	and handles errors during message reception.
//
// Dependencies:
//   - Requires the "nettsk" package for receiving UDP messages.
func (a *Agent) ListenServer() {
	dataChannel := make(chan []byte)
	errorChannel := make(chan error)

	// Start receiving data on UDP port with Receive function
	go nettsk.Receive(a.UDPPort, dataChannel, errorChannel)

	fmt.Println("Agent is listening for server messages on UDP port", a.UDPPort)

	for {
		select {
		case data := <-dataChannel:
			// Process received data
			a.handleServerMessage(data)
		case err := <-errorChannel:
			// Handle any errors reported by Receive
			fmt.Println("Error receiving data:", err)
		}
	}
}

// StartTCPIperfServer starts a TCP-based Iperf server in the background.
//
// Behavior:
//   - Executes the "iperf -s" command to start an Iperf server in TCP mode.
//   - Runs the command in the background with no output redirection.
//   - Logs success or returns an error if the server fails to start.
//
// Returns:
//   - error: An error if the Iperf server cannot be started.
func (a *Agent) StartTCPIperfServer() error {
	// Command to start the iperf server in the background
	cmd := exec.Command("iperf", "-s")

	// Redirect standard output and error for debugging/logging purposes
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Start the iperf server
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start tcp iperf server: %v", err)
	}

	fmt.Println("TCP Iperf server started successfully")
	return nil
}

// StartUDPIperfServer starts a UDP-based Iperf server in the background.
//
// Behavior:
//   - Executes the "iperf -s -u" command to start an Iperf server in UDP mode.
//   - Runs the command in the background with no output redirection.
//   - Logs success or returns an error if the server fails to start.
//
// Returns:
//   - error: An error if the Iperf server cannot be started.
func (a *Agent) StartUDPIperfServer() error {
	// Command to start the iperf server in the background
	cmd := exec.Command("iperf", "-s", "-u")

	// Redirect standard output and error for debugging/logging purposes
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Start the iperf server
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start udp iperf server: %v", err)
	}

	fmt.Println("UDP Iperf server started successfully")
	return nil
}

// handleServerMessage processes incoming messages from the server.
//
// Parameters:
//   - data: The raw byte array containing the server's message.
//
// Behavior:
//   - Converts the message from bytes to a string for processing.
//   - Checks if the message starts with the "TASK|" prefix.
//   - If a task is identified, extracts the task details and registers it by calling `registerTask`.
//
// Example:
//
//	A message like "TASK|{...}" will be passed to `registerTask` for deserialization and storage.
func (a *Agent) handleServerMessage(data []byte) {
	message := string(data)
	if strings.HasPrefix(message, "TASK|") {
		task := strings.TrimPrefix(message, "TASK|")
		a.registerTask(task)
	}
}

// registerTask deserializes and registers a new task for the agent.
//
// Parameters:
//   - task: The string representation of the task in JSON format.
//
// Behavior:
//   - Parses the JSON string into a `message.Task` object using the `parsers.DeserializeJSON` method.
//   - If successful, adds the task to the agent's task list and logs the task ID.
//   - Logs an error if the JSON deserialization fails.
//
// Example:
//
//	A valid task string like '{"TaskID": "Task1", ...}' will be deserialized and added to the agent's task list.
//
// Dependencies:
//   - Requires the "parsers" package for JSON deserialization.
func (a *Agent) registerTask(task string) {
	var newTask message.Task
	err := parsers.DeserializeJSON([]byte(task), &newTask)
	if err != nil {
		fmt.Println("Couldn't register task!") //nome da task
	}

	a.AddTask(newTask)
	fmt.Println("Registered new task:", newTask.TaskID)
}
