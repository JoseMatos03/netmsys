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

// Send function sends a message to a given UDP address
func Send(addr string, port string, data []byte) {
	// Resolve the UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", addr+":"+port)
	if err != nil {
		fmt.Println("Error resolving address:", err)
		os.Exit(1)
	}

	// Create a UDP connection
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("Error dialing UDP:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Send the message
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("Error sending message:", err)
		os.Exit(1)
	}

	fmt.Println("Message sent to", addr)
}
