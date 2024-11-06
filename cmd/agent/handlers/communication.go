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
	"netmsys/pkg/nettsk"
	"netmsys/tools/parsers"
	"strings"
)

func (a *Agent) Register() {
	registerMessage := "REGISTER|" + a.ID + ":" + a.IPAddr
	nettsk.Send(a.ServerAddr, a.UDPPort, []byte(registerMessage))
}

func (a *Agent) ListenServer() {
	dataChannel := make(chan []byte)
	errorChannel := make(chan error)

	// Start receiving data on UDP port with Receive function
	go func() {
		for {
			nettsk.Receive(a.UDPPort, dataChannel, errorChannel)
		}
	}()

	fmt.Println("Agent is listening for server messages on UDP port", a.UDPPort)

	for {
		select {
		case data := <-dataChannel:
			// Process received data
			a.handleServerMessage(data)
		case err := <-errorChannel:
			// Handle any errors reported by Receive
			fmt.Println("Error receiving data:", err)
		}
	}
}

func (a *Agent) handleServerMessage(data []byte) {
	message := string(data)
	if strings.HasPrefix(message, "TASK|") {
		task := strings.TrimPrefix(message, "TASK|")
		a.registerTask(task)
	}
}

func (a *Agent) registerTask(task string) {
	var newTask message.Task
	err := parsers.DeserializeJSON([]byte(task), &newTask)
	if err != nil {
		fmt.Println("Couldn't register task!")
	}

	a.Tasks = append(a.Tasks, newTask)
	fmt.Println("Registered new task:", newTask.TaskID)
}
