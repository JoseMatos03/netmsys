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
	"time"
)

const (
	// RECOVERY represents the format for recovery requests for a specific packet.
	RECOVERY = "RECOVERY|%d"

	// FAST_RECOVERY is the format for fast recovery requests to re-transmit missing packets.
	FAST_RECOVERY = "FAST_RECOVERY|%d"

	// ACK_AGREEMENT is sent to confirm initial agreement for packet transmission.
	ACK_AGREEMENT = "ACK_AGREEMENT"

	// ACK_COMPLETE is sent to indicate successful completion of transmission.
	ACK_COMPLETE = "ACK_COMPLETE"

	// MAX_RETRANSMIT defines the maximum number of retransmission attempts.
	MAX_RETRANSMIT = 5

	// TIMEOUT defines the duration to wait before declaring a packet timeout.
	TIMEOUT = 1 * time.Second
)

// Receive listens for UDP packets on the specified port, handling
// packet loss and retransmissions. Received data is sent through dataChannel,
// and any errors are sent through errorChannel.
//
// Parameters:
//   - port: The port to listen on for incoming UDP packets.
//   - dataChannel: A channel where the reassembled message is sent upon successful reception.
//   - errorChannel: A channel where errors are sent if transmission fails.
func Receive(port string, dataChannel chan<- []byte, errorChannel chan<- error) {
	addr, _ := net.ResolveUDPAddr("udp", ":"+port)
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()

	buf := make([]byte, 1024)

	// Wait indefinitely for an initial agreement message
	numPackets, clientAddr, err := establishAgreement(conn, buf)
	if err != nil {
		fmt.Println("Failed to establish agreement:", err)
		return
	}

	// Process packet reception with recovery mechanisms
	receivedPackets := make(map[int][]byte)
	expectedSeq, err := receivePackets(conn, clientAddr, numPackets, receivedPackets)
	if err != nil {
		errorChannel <- err
		return
	}
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

// establishAgreement waits for an "AGREEMENT" message from the client to
// determine the number of packets to be sent. It responds with ACK_AGREEMENT.
//
// Parameters:
//   - conn: The UDP connection to listen on.
//   - buf: A buffer to store received data.
//
// Returns:
//   - int: The number of packets expected to receive.
//   - *net.UDPAddr: The client address.
//   - error: An error if the agreement cannot be established.
func establishAgreement(conn *net.UDPConn, buf []byte) (int, *net.UDPAddr, error) {
	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return 0, nil, fmt.Errorf("error reading from UDP: %v", err)
		}

		agreementMsg := string(buf[:n])
		if strings.HasPrefix(agreementMsg, "AGREEMENT") {
			parts := strings.Split(agreementMsg, "|")
			numPackets, _ := strconv.Atoi(parts[1])
			conn.WriteToUDP([]byte(ACK_AGREEMENT), clientAddr)
			return numPackets, clientAddr, nil
		}
	}
}

// receivePackets handles the reception of UDP packets. If a packet timeout occurs,
// it sends a FAST_RECOVERY request to re-transmit the missing packet.
//
// Parameters:
//   - conn: The UDP connection for receiving packets.
//   - clientAddr: The client's address.
//   - numPackets: The total number of packets expected.
//   - receivedPackets: A map to store received packets by sequence number.
//
// Returns:
//   - int: The next expected sequence number, or an error if packet reception fails.
func receivePackets(conn *net.UDPConn, clientAddr *net.UDPAddr, numPackets int, receivedPackets map[int][]byte) (int, error) {
	buf := make([]byte, 1024)
	receivedSeq := -1
	expectedSeq := 0
	retransmitCount := 0

	for receivedSeq < numPackets-1 && retransmitCount < MAX_RETRANSMIT {
		conn.SetReadDeadline(time.Now().Add(TIMEOUT))
		n, _, err := conn.ReadFromUDP(buf)

		if err != nil { // Timeout occurred
			fmt.Printf("Timeout at packet %d. Sending FAST_RECOVERY.\n", expectedSeq)
			conn.WriteToUDP([]byte(fmt.Sprintf(FAST_RECOVERY, expectedSeq)), clientAddr)
			retransmitCount++
			continue
		}

		packet := string(buf[:n])
		if strings.HasPrefix(packet, "AGREEMENT") {
			continue
		}
		parts := strings.SplitN(packet, "|", 2)
		seqNum, _ := strconv.Atoi(parts[0])

		if seqNum == expectedSeq {
			receivedPackets[seqNum] = []byte(parts[1])
			receivedSeq = seqNum
			expectedSeq++
			retransmitCount = 0
			fmt.Printf("Received in-order packet %d/%d\n", seqNum+1, numPackets)
		} else if seqNum > expectedSeq {
			receivedSeq = seqNum
			receivedPackets[seqNum] = []byte(parts[1])
			fmt.Printf("Out-of-order packet %d received. Expecting %d.\n", seqNum, expectedSeq)
		}
	}
	if receivedSeq < numPackets-1 {
		return expectedSeq, fmt.Errorf("failed to receive all packets")
	}
	return expectedSeq, nil
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
		fmt.Println("Missing packets detected. Sending RECOVERY requests.")

		for _, missingSeq := range missingPackets {
			retries := 0
			packetReceived := false

			for !packetReceived && retries < MaxRetransmits {
				// Send RECOVERY request for the missing packet
				conn.WriteToUDP([]byte(fmt.Sprintf(Recovery, missingSeq)), clientAddr)
				fmt.Printf("Sent RECOVERY request for packet %d (Attempt %d/%d)\n", missingSeq, retries+1, MaxRetransmits)

				// Set a read deadline for the response
				conn.SetReadDeadline(time.Now().Add(TimeoutDuration))
				buf := make([]byte, 1024)
				n, _, err := conn.ReadFromUDP(buf)

				// Check if the missing packet was received
				if err == nil {
					receivedPackets[missingSeq] = buf[:n]
					packetReceived = true
					fmt.Printf("Successfully received missing packet %d\n", missingSeq)
				} else {
					// Timeout or other error, increase retry count
					fmt.Printf("Timeout waiting for missing packet %d. Retrying...\n", missingSeq)
					retries++
				}
			}

			// If packet was not received after all retries, return an error
			if !packetReceived {
				return fmt.Errorf("failed to receive missing packet %d after %d attempts", missingSeq, MaxRetransmits)
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
		fmt.Println("Final ACK sent. Transmission complete.")
		dataChannel <- reassembleMessage(receivedPackets)
	} else {
		fmt.Println("Failed to receive all packets.")
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
