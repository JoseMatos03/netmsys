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
	"bufio"
	"fmt"
	"os"
	"strings"
)

func (agent *Agent) StartCLI() {
	// Display agent information
	fmt.Printf("Agent %s is ready.\n", agent.ID)
	fmt.Printf("Listening for UDP on port %s and TCP on port %s. Accepting packets only from server at %s.\n", agent.UDPPort, agent.TCPPort, agent.ServerAddr)

	// Start interface loop
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Type a command: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		// Handle other commands here if needed
		fmt.Printf("Unknown command: %s\n", input)
	}
}
