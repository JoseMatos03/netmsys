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
func Send(addr string, port string, data []byte) error {
	// Resolve the UDP address to send to
	udpAddr, err := net.ResolveUDPAddr("udp", addr+":"+port)
	if err != nil {
		return fmt.Errorf("error resolving address")
	}

	// Create a UDP connection
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return fmt.Errorf("error dialing UDP")
	}
	defer conn.Close()

	numPackets, packets := calculatePackets(data)

	succ, err := sendAgreement(conn, numPackets)
	if succ {
		err = sendPackets(conn, packets)
		if err != nil {
			return err
		}
		err = awaitAcknowledgment(conn, packets)
		if err != nil {
			return err
		}
	}
	return err
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
func sendAgreement(conn *net.UDPConn, numPackets int) (bool, error) {
	agreementPacket := []byte(fmt.Sprintf(Agreement, numPackets))
	ackReceived := false
	retransmits := 0

	for !ackReceived && retransmits < MaxRetransmits {
		conn.Write(agreementPacket)

		conn.SetReadDeadline(time.Now().Add(TimeoutDuration))
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			retransmits++
			continue
		}

		ack := string(buf[:n])
		if ack == AckAgreement {
			ackReceived = true
		}

		if strings.HasPrefix(ack, "FAST_RECOVERY") {
			ackReceived = true
		}
	}

	if !ackReceived {
		return false, fmt.Errorf("failed to receive ACK")
	}
	return true, nil
}

// sendPackets transmits the sequence of data packets over the UDP connection.
//
// Parameters:
// - conn: The UDP connection to send packets on.
// - packets: A slice of byte slices, each containing a packet's data.
func sendPackets(conn *net.UDPConn, packets [][]byte) error {
	for i, packet := range packets {
		packetWithSeq := fmt.Sprintf("%d|", i) + string(packet)
		_, err := conn.Write([]byte(packetWithSeq))
		if err != nil {
			return fmt.Errorf("failed to send packets")
		}
	}
	return nil
}

// awaitAcknowledgment waits for an acknowledgment indicating all packets have
// been received by the receiver. If a recovery request is received, retransmits
// the requested packets.
//
// Parameters:
// - conn: The UDP connection to receive acknowledgments and send recovery packets on.
// - packets: The original sequence of packets sent.
func awaitAcknowledgment(conn *net.UDPConn, packets [][]byte) error {
	buf := make([]byte, 1024)

	for {
		conn.SetReadDeadline(time.Now().Add(TimeoutDuration))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			return fmt.Errorf("timeout waiting for ACK. assuming successful transmission")
		}

		ack := string(buf[:n])
		if ack == AckComplete {
			return nil
		}

		if strings.HasPrefix(ack, "RECOVERY") {
			parts := strings.Split(ack, "|")
			recoverySeq, _ := strconv.Atoi(parts[1])
			err := retransmitPackets(conn, packets, recoverySeq, false) // false indicates normal recovery
			if err != nil {
				return err
			}
		}

		if strings.HasPrefix(ack, "FAST_RECOVERY") {
			parts := strings.Split(ack, "|")
			recoverySeq, _ := strconv.Atoi(parts[1])
			err := retransmitPackets(conn, packets, recoverySeq, true) // true indicates fast recovery
			if err != nil {
				return err
			}
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
func retransmitPackets(conn *net.UDPConn, packets [][]byte, startSeq int, fast bool) error {
	if fast {
		// Fast recovery: retransmit all packets from startSeq onward
		for i := startSeq; i < len(packets); i++ {
			packetWithSeq := fmt.Sprintf("%d|", i) + string(packets[i])
			_, err := conn.Write([]byte(packetWithSeq))
			if err != nil {
				return fmt.Errorf("failed to retransmit with FAST_RECOVERY")
			}
		}
	} else {
		// Normal recovery: retransmit only the missing packet
		packetWithSeq := fmt.Sprintf("%d|", startSeq) + string(packets[startSeq])
		_, err := conn.Write([]byte(packetWithSeq))
		if err != nil {
			return fmt.Errorf("failed to retransmit with RECOVERY")
		}
	}

	return nil
}
