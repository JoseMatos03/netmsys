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

// Package nettsk provides functionality for receiving UDP messages
// with packet retransmission and recovery mechanisms to handle
// potential packet loss in unreliable networks.
package nettsk

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Receive listens for incoming UDP messages on the specified port and handles
// client communication with packet retransmission and recovery mechanisms.
//
// Parameters:
//   - port: The port to listen for incoming UDP connections.
//   - dataChannel: A channel used to send the final reassembled message.
//   - errorChannel: A channel used to report errors during communication.
//
// Behavior:
//   - Listens for AGREEMENT messages from clients, determines the number of expected packets,
//     and dynamically assigns a unique port for each client.
//   - Delegates data reception and recovery to a separate goroutine per client.
func Receive(port string, dataChannel chan<- []byte, errorChannel chan<- error) {
	mainAddr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		errorChannel <- err
		return
	}

	mainConn, err := net.ListenUDP("udp", mainAddr)
	if err != nil {
		errorChannel <- err
		return
	}
	defer mainConn.Close()

	var mu sync.Mutex
	clientPorts := make(map[string]int)

	// Goroutine to handle each incoming client
	for {
		// Establish agreement
		numPackets, clientAddr, err := establishAgreement(mainConn)
		if err != nil {
			errorChannel <- err
		}

		// Check if the client is new
		mu.Lock()
		if _, exists := clientPorts[clientAddr.String()]; !exists {
			newPort := 50000 + len(clientPorts) // Assign a new port dinamically
			clientPorts[clientAddr.String()] = newPort

			// Inform the client of the new port
			_, err = mainConn.WriteToUDP([]byte(fmt.Sprintf(ACK_AGREEMENT, newPort)), clientAddr)
			if err != nil {
				errorChannel <- err
				mu.Unlock()
				continue
			}

			// Handle communication with the new client on the new port
			go handleClient(mainConn, clientAddr, newPort, dataChannel, errorChannel, numPackets)
		}
		mu.Unlock()
	}
}

// establishAgreement handles the initial communication with a client
// to agree on the number of packets expected.
//
// Parameters:
//   - mainConn: The UDP connection used for initial communication.
//
// Returns:
//   - int: The number of packets expected from the client.
//   - *net.UDPAddr: The address of the client.
//   - error: An error if the agreement could not be established.
func establishAgreement(mainConn *net.UDPConn) (int, *net.UDPAddr, error) {
	buffer := make([]byte, 1024)
	n, clientAddr, err := mainConn.ReadFromUDP(buffer)
	if err != nil {
		return 0, nil, fmt.Errorf("error reading from UDP: %v", err)
	}

	agreementMsg := string(buffer[:n])
	if strings.HasPrefix(agreementMsg, "AGREEMENT") {
		parts := strings.Split(agreementMsg, "|")
		numPackets, _ := strconv.Atoi(parts[1])
		return numPackets, clientAddr, nil
	}
	return 0, nil, fmt.Errorf("unrecognizable packet")
}

// handleClient manages communication with an individual client, ensuring
// all packets are received and recovered if necessary.
//
// Parameters:
//   - oldConn: The main UDP connection.
//   - clientAddr: The client's address.
//   - port: The new port assigned to the client.
//   - dataChannel: The channel to send the final reassembled message.
//   - errorChannel: The channel to report errors.
//   - numPackets: The total number of packets expected from the client.
func handleClient(oldConn *net.UDPConn, clientAddr *net.UDPAddr, port int, dataChannel chan<- []byte, errorChannel chan<- error, numPackets int) {
	clientAddrPort, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		errorChannel <- err
		return
	}

	conn, err := net.ListenUDP("udp", clientAddrPort)
	if err != nil {
		errorChannel <- err
		return
	}
	defer conn.Close()

	// Process packet reception with recovery mechanisms
	receivedPackets := make(map[int][]byte)
	expectedSeq := receivePackets(oldConn, port, conn, clientAddr, numPackets, receivedPackets)
	if expectedSeq < numPackets {
		// Handle any missing packets after initial reception
		err = handleMissingPackets(conn, clientAddr, numPackets, receivedPackets)
		if err != nil {
			errorChannel <- err
			return
		}
	}

	// Finalize transmission by sending the final ACK and reassemble message
	finalizeTransmission(conn, clientAddr, dataChannel, receivedPackets, numPackets)
}

// receivePackets handles the reception of packets from a client, ensuring proper sequencing
// and initiating retransmission requests for missing packets.
//
// Parameters:
//   - oldConn: The main UDP connection used for initial communication.
//   - newPort: The dynamically assigned port for the client.
//   - conn: The dedicated UDP connection for communication with the client.
//   - clientAddr: The address of the client.
//   - numPackets: The total number of packets expected from the client.
//   - receivedPackets: A map to store received packets, keyed by their sequence numbers.
//
// Returns:
//   - int: The next expected sequence number, which indicates the progress of packet reception.
func receivePackets(oldConn *net.UDPConn, newPort int, conn *net.UDPConn, clientAddr *net.UDPAddr, numPackets int, receivedPackets map[int][]byte) int {
	buf := make([]byte, 1024)
	expectedSeq := 0
	retransmitCount := 0

	for expectedSeq < numPackets && retransmitCount < MAX_RETRANSMIT {
		conn.SetReadDeadline(time.Now().Add(TIMEOUT))
		n, _, err := conn.ReadFromUDP(buf)

		if err != nil { // Timeout occurred
			conn.WriteToUDP([]byte(fmt.Sprintf(FAST_RECOVERY, expectedSeq)), clientAddr)
			retransmitCount++
			continue
		}

		packet := string(buf[:n])
		if strings.HasPrefix(packet, "AGREEMENT") {
			oldConn.WriteToUDP([]byte(fmt.Sprintf(ACK_AGREEMENT, newPort)), clientAddr)
			continue
		}
		parts := strings.SplitN(packet, "|", 2)
		seqNum, _ := strconv.Atoi(parts[0])

		if seqNum == expectedSeq {
			receivedPackets[seqNum] = []byte(parts[1])
			expectedSeq++
			retransmitCount = 0
		} else if seqNum > expectedSeq {
			receivedPackets[seqNum] = []byte(parts[1])
		}
	}
	return expectedSeq
}

// handleMissingPackets sends RECOVERY requests for any missing packets
// and waits for retransmissions from the client.
//
// Parameters:
//   - conn: The UDP connection to communicate with the client.
//   - clientAddr: The client’s address.
//   - numPackets: The total number of packets expected.
//   - receivedPackets: A map to store received packets by sequence number.
//
// Returns:
//   - error: An error if retransmissions fail for missing packets.
func handleMissingPackets(conn *net.UDPConn, clientAddr *net.UDPAddr, numPackets int, receivedPackets map[int][]byte) error {
	missingPackets := identifyMissingPackets(receivedPackets, numPackets)
	if len(missingPackets) > 0 {

		for _, missingSeq := range missingPackets {
			retries := 0
			packetReceived := false

			for !packetReceived && retries < MAX_RETRANSMIT {
				// Send RECOVERY request for the missing packet
				conn.WriteToUDP([]byte(fmt.Sprintf(RECOVERY, missingSeq)), clientAddr)
				// Set a read deadline for the response
				conn.SetReadDeadline(time.Now().Add(TIMEOUT))
				buf := make([]byte, 1024)
				n, _, err := conn.ReadFromUDP(buf)

				// Check if the missing packet was received
				if err == nil {
					receivedPackets[missingSeq] = buf[:n]
					packetReceived = true
				} else {
					// Timeout or other error, increase retry count
					retries++
				}
			}

			// If packet was not received after all retries, return an error
			if !packetReceived {
				return fmt.Errorf("failed to receive missing packet %d after %d attempts", missingSeq, MAX_RETRANSMIT)
			}
		}
	}

	return nil
}

// finalizeTransmission concludes the data reception by sending an ACK_COMPLETE
// message to the client and reassembling the received data.
//
// Parameters:
//   - conn: The UDP connection for communication.
//   - clientAddr: The client’s address.
//   - dataChannel: Channel where the reassembled message is sent.
//   - receivedPackets: A map of received packets by sequence number.
//   - numPackets: The total number of expected packets.
func finalizeTransmission(conn *net.UDPConn, clientAddr *net.UDPAddr, dataChannel chan<- []byte, receivedPackets map[int][]byte, numPackets int) {
	if len(receivedPackets) == numPackets {
		conn.WriteToUDP([]byte(ACK_COMPLETE), clientAddr)
		dataChannel <- reassembleMessage(receivedPackets)
	}
}

// identifyMissingPackets returns a list of sequence numbers for packets
// that have not been received.
//
// Parameters:
//   - packets: A map of received packets by sequence number.
//   - total: The total number of packets expected.
//
// Returns:
//   - []int: A slice of sequence numbers for missing packets.
func identifyMissingPackets(packets map[int][]byte, total int) []int {
	var missing []int
	for i := 0; i < total; i++ {
		if _, exists := packets[i]; !exists {
			missing = append(missing, i)
		}
	}
	return missing
}

// reassembleMessage reconstructs the full message from received packets.
//
// Parameters:
//   - packets: A map of received packets by sequence number.
//
// Returns:
//   - []byte: The reassembled message as a byte slice.
func reassembleMessage(packets map[int][]byte) []byte {
	var messageBuilder strings.Builder
	for i := 0; i < len(packets); i++ {
		messageBuilder.Write(packets[i])
	}
	return []byte(messageBuilder.String())
}
