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
	"os"
)

// Receive function listens on a given UDP port and prints received messages
func Receive(addr string, port string) {
	// Resolve the UDP address to listen on
	udpAddr, err := net.ResolveUDPAddr("udp", addr+":"+port)
	if err != nil {
		fmt.Println("Error resolving address:", err)
		os.Exit(1)
	}

	// Create a UDP socket to listen
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error setting up UDP listener:", err)
		os.Exit(1)
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	// Read incoming messages
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			continue
		}

		fmt.Printf("Received message from %s: %s\n", addr.String(), string(buffer[:n]))
	}
}
