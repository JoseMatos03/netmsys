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

// Agent represents the agent's information
type Agent struct {
	ID         string
	IPAddr     string
	ServerAddr string
	UDPPort    string
	TCPPort    string
}

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
