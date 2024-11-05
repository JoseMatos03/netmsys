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
	"netmsys/cmd/message"
	"netmsys/pkg/nettsk"
	"netmsys/tools/parsers"
	"os"
	"strings"
)

// CommandLineInterface runs an infinite loop to handle user commands.
func (server *Server) StartCLI() {
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
			sendCommand(command)

		case strings.HasPrefix(command, "help"):
			helpCommand(command)

		case command == "quit":
			quitCommand()

		default:
			fmt.Println("Unknown command. Available commands: load_task <json-file>, quit")
		}
	}
}

func sendCommand(command string) {
	// Extract the path to the JSON file from the command
	// Assuming the format is "send <path-to-json>"
	commandParts := strings.Split(command, " ")
	if len(commandParts) < 2 {
		fmt.Println("Invalid command format. Usage: send <path-to-json>")
		return
	}

	jsonFile := commandParts[1]

	// Step 1: Read the JSON file into a Task struct
	var task message.Task
	err := parsers.ReadJSONFile(jsonFile, &task)
	if err != nil {
		fmt.Printf("Failed to read task JSON file: %v\n", err)
		return
	}

	// Step 2: Analyze the targets (assuming Task struct has a 'Targets' field)
	if len(task.Targets) == 0 {
		fmt.Println("No targets specified in the task.")
		return
	}

	// Step 3: Serialize the task into JSON bytes
	taskData, err := parsers.SerializeJSON(task)
	if err != nil {
		fmt.Printf("Failed to serialize task: %v\n", err)
		return
	}

	// Step 4: Send the serialized task to each target via the Nettsk protocol
	for _, target := range task.Targets {
		fmt.Printf("Sending task to target %s via Nettsk...\n", target)
		go nettsk.Send(target.IPAddress, target.UDPPort, taskData)
	}
}

func helpCommand(command string) {
	helpArgs := strings.Split(command, " ")
	if len(helpArgs) == 1 {
		printGeneralHelp()
	} else if len(helpArgs) == 2 {
		printCommandHelp(helpArgs[1])
	} else {
		fmt.Println("Usage: help <command_name>")
	}
}

func quitCommand() {
	fmt.Println("Shutting down server...")
	os.Exit(0)
}

// printGeneralHelp displays a general help message
func printGeneralHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  send <json-file>       - Send a task from the specified JSON file")
	fmt.Println("  quit                   - Quit the server")
	fmt.Println("  help                   - Show general help information")
	fmt.Println("  help <command_name>    - Show specific help for a command")
}

// printCommandHelp displays specific help for a given command
func printCommandHelp(command string) {
	switch command {
	case "load_task":
		fmt.Println("Usage: send <json-file>")
		fmt.Println("Description: Sends a task from the specified JSON file to its targets.")
	case "quit":
		fmt.Println("Usage: quit")
		fmt.Println("Description: Shuts down the server and exits the program.")
	case "help":
		fmt.Println("Usage: help <command_name>")
		fmt.Println("Description: Provides detailed information for a specific command.")
	default:
		fmt.Println("No help available for the specified command.")
	}
}
