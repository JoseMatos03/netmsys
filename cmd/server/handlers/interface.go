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
	"netmsys/tools/parsers"
	"os"
	"strings"
)

// CommandLineInterface runs an infinite loop to handle user commands.
func (s *Server) StartCLI() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ") // Command-line prompt
		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		// Trim the input to avoid issues with newlines
		command := strings.TrimSpace(input)
		var commErr error

		// Handle commands
		switch {
		case strings.HasPrefix(command, "load"):
			commErr = s.loadCommand(command)

		case strings.HasPrefix(command, "send"):
			commErr = s.sendCommand(command)

		case strings.HasPrefix(command, "help"):
			s.helpCommand(command)

		case command == "man":
			s.manCommand()

		case command == "quit":
			s.quitCommand()

		default:
			fmt.Println("Unknown command. Available commands: load_task <json-file>, quit")
		}

		if commErr != nil {
			fmt.Printf("interface.StartCLI(): Error executing command.\n%v\n", commErr)
		}
	}
}

func (s *Server) loadCommand(command string) error {
	// Extract the path to the JSON file from the command
	// Assuming the format is "send <path-to-json>"
	commandParts := strings.Split(command, " ")
	if len(commandParts) < 2 {
		return fmt.Errorf("invalid command format. Usage: load <path-to-json>")
	}

	jsonFile := commandParts[1]

	// Read the JSON file into a Task struct
	var task message.Task
	err := parsers.ReadJSONFile(jsonFile, &task)
	if err != nil {
		return fmt.Errorf("failed to read task JSON file: %v", err)
	}

	// Check if the task is already in the array
	for _, existingTask := range s.Tasks {
		if existingTask.TaskID == task.TaskID {
			fmt.Printf("Task %s is already loaded. Skipping.\n", task.TaskID)
			return nil
		}
	}

	// Add the task to the list and send it
	s.Tasks = append(s.Tasks, task)
	fmt.Printf("Loaded task %s.\n", task.TaskID)
	s.SendTask(task.TaskID)
	return nil
}

func (s *Server) sendCommand(command string) error {
	// Extract the path to the JSON file from the command
	// Assuming the format is "send <path-to-json>"
	commandParts := strings.Split(command, " ")
	if len(commandParts) < 2 {
		return fmt.Errorf("invalid command format. Usage: load <path-to-json>")
	}

	jsonFile := commandParts[1]

	// Read the JSON file into a Task struct
	var task message.Task
	err := parsers.ReadJSONFile(jsonFile, &task)
	if err != nil {
		return fmt.Errorf("failed to read task JSON file: %v", err)
	}

	// Check if the task is already in the array and send it
	for _, existingTask := range s.Tasks {
		if existingTask.TaskID == task.TaskID {
			s.SendTask(task.TaskID)
			return nil
		}
	}

	return fmt.Errorf("task %s is not loaded", task.TaskID)
}

func (s *Server) helpCommand(command string) {
	helpArgs := strings.Split(command, " ")
	if len(helpArgs) == 2 {
		printCommandHelp(helpArgs[1])
	} else {
		fmt.Println("Usage: help <command_name>")
	}
}

func (s *Server) manCommand() {
	fmt.Println("Available commands:")
	fmt.Println("  send <json-file>       - Send a task from the specified JSON file")
	fmt.Println("  quit                   - Quit the server")
	fmt.Println("  help <command_name>    - Show specific help for a command")
}

func (s *Server) quitCommand() {
	fmt.Println("Shutting down server...")
	os.Exit(0)
}

// printCommandHelp displays specific help for a given command
func printCommandHelp(command string) {
	switch command {
	case "load_task":
		fmt.Println("Usage: load_task <json-file>")
		fmt.Println("Description: Loads a task from the specified JSON file and sends it to its targets.")
	case "send":
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
