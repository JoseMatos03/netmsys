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
)

// Server represents a server that facilitates communication with agents and handles tasks.
//
// Fields:
//   - UDPPort: The port on which the server listens for agent messages via UDP.
//   - TCPPort: The port on which the server listens for alert messages via TCP.
//   - Agents: A map of registered agents, where the key is the agent ID and the value is the agent's IP address.
//   - Tasks: A slice of tasks to be distributed to agents.
type Server struct {
	UDPPort string
	TCPPort string
	Agents  map[string]string
	Tasks   []message.Task
}

// NewServer initializes and returns a new Server instance.
//
// Parameters:
//   - args: A slice of command-line arguments. The first argument is the program name,
//     the second is the UDP port, and the third is the TCP port.
//
// Returns:
//   - (*Server, error): A pointer to a newly created Server instance, or an error
//     if the arguments are insufficient.
//
// Behavior:
//   - Validates that the required arguments (UDP and TCP ports) are provided.
//   - Initializes the `Agents` map as empty.
//   - Initializes the `Tasks` slice as empty.
//   - Assigns the specified ports to the server.
//
// Error Handling:
//   - Returns an error if fewer than two arguments are provided, indicating insufficient arguments.
//
// Notes:
//   - The `NewServer` function assumes that ports are passed as string arguments.
func NewServer(args []string) (*Server, error) {
	// Check if enough arguments are provided
	if len(args) < 2 {
		return nil, fmt.Errorf("insufficient arguments: ./server <UDP port> <TCP port>")
	}

	serverUDP := args[1]
	serverTCP := args[2]
	return &Server{
		UDPPort: serverUDP,
		TCPPort: serverTCP,
		Agents:  make(map[string]string),
		Tasks:   []message.Task{},
	}, nil
}
