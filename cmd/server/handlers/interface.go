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

// CommandLineInterface runs an infinite loop to handle user commands.
func CommandLineInterface() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ") // Command-line prompt
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		// Trim the input to avoid issues with newlines
		command := strings.TrimSpace(input)

		// Handle commands
		switch {
		case strings.HasPrefix(command, "send"):
			SendCommand(command)

		case strings.HasPrefix(command, "help"):
			HelpCommand(command)

		case command == "quit":
			QuitCommand()

		default:
			fmt.Println("Unknown command. Available commands: load_task <json-file>, quit")
		}
	}
}
