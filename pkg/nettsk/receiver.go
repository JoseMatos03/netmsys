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

package nettsk

import (
	"fmt"
	"net"
)

// Receive function listens on a given UDP port and prints received messages
func Receive(port string, dataChannel chan<- []byte, errorChannel chan<- error) {
	// Resolve the UDP address to listen on
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		errorChannel <- fmt.Errorf("error resolving address: %w", err)
		return
	}

	// Create a UDP socket to listen
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		errorChannel <- fmt.Errorf("error setting up UDP listener: %w", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	// Continuous loop to receive messages
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			errorChannel <- fmt.Errorf("error reading from UDP: %w", err)
			continue
		}

		fmt.Printf("Received message from %s\n", addr.String())

		// Send received data to the channel
		dataChannel <- buffer[:n]
	}
}
