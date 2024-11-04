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
	"netmsys/pkg/nettsk"
)

func (agent *Agent) Start() {
	dataChannel := make(chan []byte)
	errorChannel := make(chan error)
	go func() {
		for {
			nettsk.Receive(agent.UDPPort, dataChannel, errorChannel)
		}
	}()
	for {
		select {
		case data := <-dataChannel:
			fmt.Printf("Client: Received task data:\n%s\n", string(data))
		case err := <-errorChannel:
			fmt.Printf("Client: Error receiving data: %v\n", err)
		}
	}
}
