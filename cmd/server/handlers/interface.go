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

// StartCLI initializes and manages the Command-Line Interface (CLI) for the server.
//
// This function provides an interactive loop for user input to manage tasks and server operations.
// It continuously waits for user commands and delegates them to specific handler functions.
//
// Supported commands:
//   - `load <path-to-json>`: Loads a task from the given JSON file and sends it to agents.
//   - `send <task-id>`: Sends a previously loaded task to agents.
//   - `help <command_name>`: Displays help information for a specific command.
//   - `man`: Lists all available commands with a brief description.
//   - `quit`: Terminates the server process gracefully.
//
// Notes:
//   - Commands are case-sensitive.
//   - If an invalid command is entered, the CLI displays an error message.
//
// Parameters:
//   - s (*Server): A reference to the Server instance.
func (s *Server) StartCLI() {
	reader := bufio.NewReader(os.Stdin)

	for {
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

// loadCommand processes the "load" command to load and send a task from a JSON file.
//
// The command expects the format "load <path-to-json>". It reads the specified JSON file,
// parses it into a Task structure, and sends the task to agents if it's not already loaded.
//
// Behavior:
//   - Checks if the JSON file path is provided in the command.
//   - Parses the JSON file into a `Task` structure.
//   - Skips loading if the task is already present in the server's task list.
//   - Sends the loaded task to agents using the `SendTask` method.
//
// Parameters:
//   - command (string): The user input command string.
//
// Returns:
//   - error: An error if the command format is invalid or the task fails to load or send.
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
	err = s.SendTask(task.TaskID)
	if err != nil {
		return err
	}
	return nil
}

// sendCommand processes the "send" command to send a preloaded task.
//
// The command expects the format "send <task-id>". It retrieves the task by its ID
// from the server's task list and sends it to agents.
//
// Parameters:
//   - command (string): The user input command string.
//
// Returns:
//   - error: An error if the command format is invalid or the task fails to send.
func (s *Server) sendCommand(command string) error {
	commandParts := strings.Split(command, " ")
	if len(commandParts) < 2 {
		return fmt.Errorf("invalid command format. Usage: send <task-id>")
	}

	taskID := commandParts[1]
	err := s.SendTask(taskID)
	if err != nil {
		return err
	}
	return nil
}

// helpCommand displays specific help information for a command.
//
// If the command format is "help <command_name>", it displays detailed usage information for
// the specified command. If no command name is provided, it suggests the proper usage format.
//
// Parameters:
//   - command (string): The user input command string.
func (s *Server) helpCommand(command string) {
	helpArgs := strings.Split(command, " ")
	if len(helpArgs) == 2 {
		printCommandHelp(helpArgs[1])
	} else {
		fmt.Println("Usage: help <command_name>")
	}
}

// manCommand lists all available commands with a brief description.
//
// This is a general overview of the server's CLI commands.
func (s *Server) manCommand() {
	fmt.Println("Available commands:")
	fmt.Println("  load <path-to-json>    - Send a task from the specified JSON file")
	fmt.Println("  send <task-id>         - Send a task that is already loaded")
	fmt.Println("  quit                   - Quit the server")
	fmt.Println("  help <command_name>    - Show specific help for a command")
}

// quitCommand gracefully shuts down the server process.
//
// The function displays a shutdown message and exits the program.
func (s *Server) quitCommand() {
	fmt.Println("Shutting down server...")
	os.Exit(0)
}

// printCommandHelp provides detailed usage information for a specific command.
//
// This helper function displays usage details and descriptions for commands supported by the CLI.
//
// Parameters:
//   - command (string): The name of the command for which help is requested.
func printCommandHelp(command string) {
	switch command {
	case "load":
		fmt.Println("Usage: load <path-to-json>")
		fmt.Println("Description: Loads a task from the specified JSON file and sends it to its targets.")
	case "send":
		fmt.Println("Usage: send <task-id>")
		fmt.Println("Description: Sends a task that is already loaded, given it's ID.")
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
