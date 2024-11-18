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

package main

import (
	"fmt"
	"netmsys/cmd/server/handlers"
	"os"
)

func main() {
	server, err := handlers.NewServer(os.Args)
	if err != nil {
		fmt.Println("Usage: ./server <UDP Port> <TCP Port>")
		os.Exit(1)
	}

	go server.ListenAgents()
	go server.ListenAlerts()
	go server.StartUDPIperfServer()
	server.StartCLI()

	// The program will continue running until the user enters the "quit" command
}
