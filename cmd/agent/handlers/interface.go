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

// StartCLI launches a simple Command-Line Interface (CLI) for the agent.
//
// The function displays agent information, such as its ID and the server address it is
// communicating with, and provides an interactive loop for user input.
//
// Behavior:
//   - The CLI informs the user that the agent is ready and provides its server communication details.
//   - It continuously reads user input from the console and processes commands.
//   - Supported commands:
//   - `quit`: Terminates the agent process gracefully.
//   - Any unrecognized commands are acknowledged with a message indicating that the command is unknown.
//
// Parameters:
//   - agent (*Agent): A reference to the Agent instance, providing access to its ID and server address.
//
// Notes:
//   - The function runs indefinitely until the user inputs the `quit` command.
//   - Input is read from the standard input (console) and processed after trimming any whitespace.
func (agent *Agent) StartCLI() {
	// Display agent information
	fmt.Printf("Agent %s is ready.\n", agent.ID)
	fmt.Printf("Accepting packets only from server at %s.\n", agent.ServerAddr)

	// Start interface loop
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" {
			fmt.Println("Shutting down agent...")
			os.Exit(0)
		}

		fmt.Printf("Unknown command: %s\n", input)
	}
}
