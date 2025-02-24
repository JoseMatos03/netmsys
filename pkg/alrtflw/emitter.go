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
	"net"
)

// Send establishes a TCP connection to a specified address and port, transmits data,
// and ensures the connection is closed properly after the transmission.
//
// Parameters:
//   - addr: The IP address of the target server.
//   - port: The port of the target server.
//   - data: The data to be sent as a byte slice.
//
// Returns:
//   - An error if the connection or data transmission fails, otherwise nil.
func Send(addr string, port string, data []byte) {
	// Establish a TCP connection
	conn, err := net.Dial("tcp", addr+":"+port)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Send the message
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}
}
