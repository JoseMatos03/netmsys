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
//channels fazer de informação e erros
func Receive(port string, dataChannel chan <- [] byte, erroChannel chan <- error) {
	// Listen for incoming TCP connections
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		//fmt.Println("Error starting TCP listener:", err)
		dataChannel<- err
		erroChannel <- fmt.Errorf("Error starting TCP listener:")
		return
	}
	defer listener.Close() // Only defer listener.Close(), not individual connections

	for {
		// Accept a connection
		conn, err := listener.Accept()
		if err != nil {
			//fmt.Println("Error accepting connection:", err)
			dataChannel <-err
			erroChannel <- fmt.Errorf("Error accepting connection:")
			continue
		}

		// Handle the connection (read the message)
		handleConnection(conn, erroChannel, dataChannel)
		
	}
}

// handleConnection processes an incoming TCP connection
func handleConnection(conn net.Conn, erroChannel chan <- error, dataChannel chan <- []byte) {
	defer conn.Close()
	buffer := make([]byte, 1024)

	_, err1 = conn.Read(buffer)
	if err1 != nil{
		//fmt.Println("Erro reading message:", err1)
		erroChannel <- ftm.Errorf("Erro reading message:")
		dataChannel <- err1
	}
	// Read all data from the connection until the client closes it
	data, err := io.ReadAll(conn)
	if err != nil {
		erroChannel <- ftm.Errorf("Error reading from connection:")
		dataChannel <- err
		//fmt.Println("Error reading from connection:", err)
		return
	}


	// Display the received message
	dataChannel <- data
	fmt.Println("Received TCP message:", string(data))
}
