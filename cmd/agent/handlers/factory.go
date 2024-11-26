// ------------------------------------ LICENSE ------------------------------------
//
// # Copyright 2024 Ana Pires, José Matos, Rúben Oliveira
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"net"
)

// Agent represents the agent's information and connection configuration.
//
// Fields:
//   - ID: A unique identifier for the agent (e.g., "PC1").
//   - IPAddr: The local IP address of the machine running the agent.
//   - ServerAddr: The IP address of the server to which the agent will connect.
//   - UDPPort: The UDP port on which the agent will communicate with the server.
//   - TCPPort: The TCP port on which the agent will host or connect for tasks.
type Agent struct {
	ID         string
	IPAddr     string
	ServerAddr string
	UDPPort    string
	TCPPort    string
}

// NewAgent initializes a new Agent instance based on the provided command-line arguments.
//
// Parameters:
//   - args: Command-line arguments in the format:
//     ./agent <Agent ID> <Server IP> <UDP port> <TCP port>
//
// Returns:
//   - *Agent: A pointer to the newly created Agent instance.
//   - error: An error if insufficient arguments are provided or if the local IP address cannot be determined.
//
// Behavior:
//   - Validates the number of arguments provided.
//   - Determines the local IP address of the machine running the agent.
//   - Initializes an Agent struct with the provided arguments and the local IP.
//
// Dependencies:
//   - Requires the `getLocalIPAddress` function to determine the agent's IP address.
func NewAgent(args []string) (*Agent, error) {
	// Check if enough arguments are provided
	if len(args) < 4 {
		return nil, fmt.Errorf("insufficient arguments: ./agent <Agent ID> <Server IP> <UDP port> <TCP port>")
	}

	ip, err := getLocalIPAddress()
	if err != nil {
		return nil, err
	}

	agentID := args[1]
	agentServer := args[2]
	agentUDP := args[3]
	agentTCP := args[4]
	return &Agent{
		ID:         agentID,
		IPAddr:     ip,
		ServerAddr: agentServer,
		UDPPort:    agentUDP,
		TCPPort:    agentTCP,
	}, nil
}

// getLocalIPAddress retrieves the IPv4 address of the machine running the agent.
//
// Returns:
//   - string: The local IPv4 address.
//   - error: An error if no suitable IP address can be found or if network interfaces cannot be accessed.
//
// Behavior:
//   - Iterates through all network interfaces on the machine.
//   - Skips interfaces that are down or loopback interfaces.
//   - Checks each interface for an IPv4 address and returns the first suitable one.
//   - Returns an error if no IPv4 address is found.
func getLocalIPAddress() (string, error) {
	// Get a list of all available network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("could not get network interfaces: %v", err)
	}

	for _, i := range interfaces {
		// Ignore down or loopback interfaces
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Get all IP addresses assigned to this interface
		addrs, err := i.Addrs()
		if err != nil {
			return "", fmt.Errorf("could not get addresses for interface %s: %v", i.Name, err)
		}

		for _, addr := range addrs {
			// Check if the address is an IP address (ignore IPv6 link-local addresses)
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil { // Check for IPv4
					return ipNet.IP.String(), nil
				}
			}
		}
	}

	return "", fmt.Errorf("no suitable IP address found")
}
