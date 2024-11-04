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

import "fmt"

// Agent represents the agent's information
type Agent struct {
	ID         string
	TaskIDs    []string
	ServerAddr string
	UDPPort    string
	TCPPort    string
}

func NewAgent(args []string) (*Agent, error) {
	// Check if enough arguments are provided
	if len(args) < 4 {
		return nil, fmt.Errorf("insufficient arguments: ./agent <Agent ID> <Server IP> <UDP port> <TCP port>")
	}

	agentID := args[1]
	agentTaskIDs := make([]string, 10)
	agentServer := args[2]
	agentUDP := args[3]
	agentTCP := args[4]
	return &Agent{
		ID:         agentID,
		TaskIDs:    agentTaskIDs,
		ServerAddr: agentServer,
		UDPPort:    agentUDP,
		TCPPort:    agentTCP,
	}, nil
}
