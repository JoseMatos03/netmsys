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
	MaxPacketSize   = 512
	TimeoutDuration = 4 * time.Second
	MaxRetransmits  = 5
	Agreement       = "AGREEMENT|%d"
	AckAgreement    = "ACK_AGREEMENT"
	AckComplete     = "ACK_COMPLETE"
	Recovery        = "RECOVERY|%d"
	FastRecovery    = "FAST_RECOVERY|%d" // Define fast recovery signal
)

// Send transmits data to a specific address and port over UDP.
// It splits the data into packets, initiates an agreement with the receiver,
// sends all packets, and handles recovery if packets are lost or timeouts occur.
//
// Parameters:
// - addr: The IP address to send data to.
// - port: The port number to use for sending data.
// - data: The data to be sent.
func Send(addr string, port string, data []byte) {
	// Resolve the UDP address to send to
	udpAddr, err := net.ResolveUDPAddr("udp", addr+":"+port)
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	// Create a UDP connection
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("Error dialing UDP:", err)
		return
	}
	defer conn.Close()

	numPackets, packets := calculatePackets(data)

	if sendAgreement(conn, numPackets) {
		sendPackets(conn, packets)
		awaitAcknowledgment(conn, packets)
	}
}

// calculatePackets splits the input data into multiple packets based on MaxPacketSize.
//
// Parameters:
// - data: The data to be split.
//
// Returns:
// - The total number of packets.
// - A slice of byte slices, each containing a packet's data.
func calculatePackets(data []byte) (int, [][]byte) {
	numPackets := (len(data) + MaxPacketSize - 1) / MaxPacketSize
	var packets [][]byte

	for i := 0; i < len(data); i += MaxPacketSize {
		end := i + MaxPacketSize
		if end > len(data) {
			end = len(data)
		}
		packets = append(packets, data[i:end])
	}

	return numPackets, packets
}

// sendAgreement initiates the communication by sending an agreement packet with
// the number of packets to be transmitted. Retransmits if no acknowledgment is received.
//
// Parameters:
// - conn: The UDP connection to send and receive packets on.
// - numPackets: The total number of packets to be sent.
//
// Returns:
// - True if the agreement was acknowledged, false otherwise.
func sendAgreement(conn *net.UDPConn, numPackets int) bool {
	agreementPacket := []byte(fmt.Sprintf(Agreement, numPackets))
	ackReceived := false
	retransmits := 0

	for !ackReceived && retransmits < MaxRetransmits {
		conn.Write(agreementPacket)
		fmt.Printf("Sent agreement packet, expecting to send %d packets.\n", numPackets)

		conn.SetReadDeadline(time.Now().Add(TimeoutDuration))
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("Timeout waiting for ACK. Retransmitting (%d/%d)\n", retransmits+1, MaxRetransmits)
			retransmits++
			continue
		}

		ack := string(buf[:n])
		if ack == AckAgreement {
			fmt.Println("Received ACK for agreement. Sending all packets.")
			ackReceived = true
		}

		if strings.HasPrefix(ack, "FAST_RECOVERY") {
			fmt.Println("Received FAST_RECOVERY for agreement. Sending all packets.")
			ackReceived = true
		}
	}

	if !ackReceived {
		fmt.Println("Failed to receive ACK after 3 attempts. Terminating connection.")
		return false
	}
	return true
}

// sendPackets transmits the sequence of data packets over the UDP connection.
//
// Parameters:
// - conn: The UDP connection to send packets on.
// - packets: A slice of byte slices, each containing a packet's data.
func sendPackets(conn *net.UDPConn, packets [][]byte) {
	for i, packet := range packets {
		packetWithSeq := fmt.Sprintf("%d|", i) + string(packet)
		conn.Write([]byte(packetWithSeq))
		fmt.Printf("Sent packet %d/%d\n", i+1, len(packets))
	}
}

// awaitAcknowledgment waits for an acknowledgment indicating all packets have
// been received by the receiver. If a recovery request is received, retransmits
// the requested packets.
//
// Parameters:
// - conn: The UDP connection to receive acknowledgments and send recovery packets on.
// - packets: The original sequence of packets sent.
func awaitAcknowledgment(conn *net.UDPConn, packets [][]byte) {
	buf := make([]byte, 1024)

	for {
		conn.SetReadDeadline(time.Now().Add(TimeoutDuration))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Timeout waiting for ACK. Assuming successful transmission.")
			return
		}

		ack := string(buf[:n])
		if ack == AckComplete {
			fmt.Println("Received ACK_COMPLETE, all packets delivered.")
			return
		}

		if strings.HasPrefix(ack, "RECOVERY") {
			parts := strings.Split(ack, "|")
			recoverySeq, _ := strconv.Atoi(parts[1])
			fmt.Printf("Received RECOVERY ACK for packet %d. Retransmitting missing packet.\n", recoverySeq)
			retransmitPackets(conn, packets, recoverySeq, false) // false indicates normal recovery
		}

		if strings.HasPrefix(ack, "FAST_RECOVERY") {
			parts := strings.Split(ack, "|")
			recoverySeq, _ := strconv.Atoi(parts[1])
			fmt.Printf("Received FAST_RECOVERY for packet %d. Retransmitting from that sequence.\n", recoverySeq)
			retransmitPackets(conn, packets, recoverySeq, true) // true indicates fast recovery
		}
	}
}

// retransmitPackets handles the retransmission of packets based on the type of recovery requested.
//
// Parameters:
// - conn: The UDP connection to send retransmitted packets on.
// - packets: A slice of byte slices, each containing a packet's data.
// - startSeq: The sequence number to start retransmission from.
// - fast: Indicates if fast recovery mode is used (retransmits all packets from startSeq).
func retransmitPackets(conn *net.UDPConn, packets [][]byte, startSeq int, fast bool) {
	if fast {
		// Fast recovery: retransmit all packets from startSeq onward
		for i := startSeq; i < len(packets); i++ {
			packetWithSeq := fmt.Sprintf("%d|", i) + string(packets[i])
			conn.Write([]byte(packetWithSeq))
			fmt.Printf("Retransmitted packet %d/%d (Fast Recovery)\n", i+1, len(packets))
		}
	} else {
		// Normal recovery: retransmit only the missing packet
		packetWithSeq := fmt.Sprintf("%d|", startSeq) + string(packets[startSeq])
		conn.Write([]byte(packetWithSeq))
		fmt.Printf("Retransmitted missing packet %d (Normal Recovery)\n", startSeq)
	}
}
