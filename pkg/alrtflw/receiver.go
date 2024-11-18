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

package alrtflw

import (
	"net"
)

func Receive(port string, dataChannel chan<- []byte, errorChannel chan<- error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		errorChannel <- err
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			errorChannel <- err
			continue
		}

		// Handle the connection
		handleConnection(conn, dataChannel, errorChannel)
	}
}

// handleConnection processes an incoming TCP connection
func handleConnection(conn net.Conn, dataChannel chan<- []byte, errorChannel chan<- error) {
	defer conn.Close()
	buffer := make([]byte, 1024)

	_, err := conn.Read(buffer)
	if err != nil {
		errorChannel <- err
	}

	dataChannel <- buffer
}
