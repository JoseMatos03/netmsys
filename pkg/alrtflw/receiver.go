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

// Receive listens for incoming TCP connections on the specified port and forwards
// received data to a provided channel. Errors encountered during the listening
// or connection handling process are sent to a separate error channel.
//
// Parameters:
//   - port: The port on which the server will listen for incoming connections.
//   - dataChannel: A channel to which received data (as byte slices) will be sent.
//   - errorChannel: A channel to which any encountered errors will be sent.
//
// Note:
//   - The function runs indefinitely, processing connections in a loop until
//     the listener is explicitly closed.
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

// handleConnection processes an incoming TCP connection, reads data from it,
// and forwards the data or any encountered error to the respective channels.
//
// Parameters:
//   - conn: The active TCP connection to handle.
//   - dataChannel: A channel to which the received data (as a byte slice) will be sent.
//   - errorChannel: A channel to which any encountered errors will be sent.
//
// Note:
//   - The function closes the connection once processing is complete.
func handleConnection(conn net.Conn, dataChannel chan<- []byte, errorChannel chan<- error) {
	defer conn.Close()
	buffer := make([]byte, 1024)

	_, err := conn.Read(buffer)
	if err != nil {
		errorChannel <- err
	}

	dataChannel <- buffer
}
