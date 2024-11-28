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

// Package nettsk provides functionality for sending data over UDP with
// mechanisms to handle packet retransmission and recovery in case of packet loss.
package nettsk

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// Send transmits data to a specified UDP address and port with mechanisms to
// handle packet retransmission and recovery in case of packet loss.
//
// Parameters:
//   - addr: The IP address of the target server.
//   - port: The port of the target server.
//   - data: The data to be sent as a byte slice.
//
// Returns:
//   - An error if any step in the process fails, otherwise nil.
func Send(addr string, port string, data []byte) error {
	serverAddr, err := net.ResolveUDPAddr("udp", addr+":"+port)
	if err != nil {
		fmt.Println("Error resolving server address:", err)
		return err
	}

	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		fmt.Println("Error dialing UDP:", err)
		return err
	}
	defer conn.Close()

	numPackets, packets := calculatePackets(data)
	newPort, err := handleAgreement(conn, serverAddr, numPackets)
	if err != nil {
		return err
	}

	newAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", addr, newPort))
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = sendPackets(conn, packets, newAddr)
	if err != nil {
		return err
	}
	err = awaitAcknowledgment(conn, packets, newAddr)
	if err != nil {
		return err
	}

	return nil
}

// calculatePackets splits the given data into smaller packets based on
// the defined maximum packet size.
//
// Parameters:
//   - data: The data to be split as a byte slice.
//
// Returns:
//   - The number of packets.
//   - A slice of byte slices representing the packets.
func calculatePackets(data []byte) (int, [][]byte) {
	numPackets := (len(data) + MAX_PACKET_SIZE - 1) / MAX_PACKET_SIZE
	var packets [][]byte

	for i := 0; i < len(data); i += MAX_PACKET_SIZE {
		end := i + MAX_PACKET_SIZE
		if end > len(data) {
			end = len(data)
		}
		packets = append(packets, data[i:end])
	}

	return numPackets, packets
}

// handleAgreement sends an agreement packet to the server, indicating the
// number of packets to be transmitted, and negotiates a new port for communication.
//
// Parameters:
//   - conn: The UDP connection.
//   - serverAddr: The server's address.
//   - numPackets: The number of packets to be transmitted.
//
// Returns:
//   - The newly negotiated port number.
//   - An error if the agreement process fails.
func handleAgreement(conn *net.UDPConn, serverAddr *net.UDPAddr, numPackets int) (int, error) {
	agreementPacket := []byte(fmt.Sprintf(AGREEMENT, numPackets))
	retransmits := 0

	for retransmits < MAX_RETRANSMIT {
		conn.WriteToUDP(agreementPacket, serverAddr)

		conn.SetReadDeadline(time.Now().Add(TIMEOUT))
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			retransmits++
			continue
		}

		ackAgreement := string(buf[:n])
		if strings.HasPrefix(ackAgreement, "ACK_AGREEMENT") {
			parts := strings.Split(ackAgreement, "|")
			port, _ := strconv.Atoi(parts[1])
			return port, nil
		}
	}
	return -1, fmt.Errorf("connection timed-out trying to send agreement")
}

// sendPackets transmits a series of packets to the specified server address.
//
// Parameters:
//   - conn: The UDP connection.
//   - packets: A slice of packets to be sent.
//   - serverAddr: The server's address.
//
// Returns:
//   - An error if any packet fails to send, otherwise nil.
func sendPackets(conn *net.UDPConn, packets [][]byte, serverAddr *net.UDPAddr) error {
	for i, packet := range packets {
		packetWithSeq := fmt.Sprintf("%d|", i) + string(packet)
		_, err := conn.WriteToUDP([]byte(packetWithSeq), serverAddr)
		if err != nil {
			fmt.Println(err)
			return fmt.Errorf("failed to send packets")
		}
	}
	return nil
}

// awaitAcknowledgment waits for acknowledgments from the server and handles
// retransmissions if packets are lost.
//
// Parameters:
//   - conn: The UDP connection.
//   - packets: A slice of packets to be retransmitted if necessary.
//   - serverAddr: The server's address.
//
// Returns:
//   - An error if the acknowledgment process fails, otherwise nil.
func awaitAcknowledgment(conn *net.UDPConn, packets [][]byte, serverAddr *net.UDPAddr) error {
	buf := make([]byte, 1024)

	for {
		conn.SetReadDeadline(time.Now().Add(TIMEOUT))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			return fmt.Errorf("timeout waiting for ACK. assuming successful transmission")
		}

		ack := string(buf[:n])
		if ack == ACK_COMPLETE {
			return nil
		}

		if strings.HasPrefix(ack, "RECOVERY") {
			parts := strings.Split(ack, "|")
			recoverySeq, _ := strconv.Atoi(parts[1])
			err := retransmitPackets(conn, serverAddr, packets, recoverySeq, false) // false indicates normal recovery
			if err != nil {
				return err
			}
		}

		if strings.HasPrefix(ack, "FAST_RECOVERY") {
			parts := strings.Split(ack, "|")
			recoverySeq, _ := strconv.Atoi(parts[1])
			err := retransmitPackets(conn, serverAddr, packets, recoverySeq, true) // true indicates fast recovery
			if err != nil {
				return err
			}
		}
	}
}

// retransmitPackets retransmits packets starting from a specific sequence
// number, based on the recovery type.
//
// Parameters:
//   - conn: The UDP connection.
//   - serverAddr: The server's address.
//   - packets: A slice of packets to be retransmitted.
//   - startSeq: The sequence number to start retransmitting from.
//   - fast: A boolean indicating whether to use fast recovery.
//
// Returns:
//   - An error if retransmission fails, otherwise nil.
func retransmitPackets(conn *net.UDPConn, serverAddr *net.UDPAddr, packets [][]byte, startSeq int, fast bool) error {
	if fast {
		// Fast recovery: retransmit all packets from startSeq onward
		for i := startSeq; i < len(packets); i++ {
			packetWithSeq := fmt.Sprintf("%d|", i) + string(packets[i])
			_, err := conn.WriteToUDP([]byte(packetWithSeq), serverAddr)
			if err != nil {
				return fmt.Errorf("failed to retransmit with FAST_RECOVERY")
			}
		}
	} else {
		// Normal recovery: retransmit only the missing packet
		packetWithSeq := fmt.Sprintf("%d|", startSeq) + string(packets[startSeq])
		_, err := conn.WriteToUDP([]byte(packetWithSeq), serverAddr)
		if err != nil {
			return fmt.Errorf("failed to retransmit with RECOVERY")
		}
	}

	return nil
}
