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
	"fmt"
	"io"
	"net"
)

// Receive listens on the given TCP address and receives messages
func Receive(addr string, port string) {
	// Listen for incoming TCP connections
	listener, err := net.Listen("tcp", addr+":"+port)
	if err != nil {
		fmt.Println("Error starting TCP listener:", err)
		return
	}
	defer listener.Close() // Only defer listener.Close(), not individual connections

	for {
		// Accept a connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle the connection (read the message)
		handleConnection(conn)
	}
}

// handleConnection processes an incoming TCP connection
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read all data from the connection until the client closes it
	data, err := io.ReadAll(conn)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	// Display the received message
	fmt.Println("Received TCP message:", string(data))
}
