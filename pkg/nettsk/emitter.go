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

func Send(addr string, port string, data []byte) error {
	serverAddr, err := net.ResolveUDPAddr("udp", addr+":"+port)
	if err != nil {
		fmt.Println("Error resolving server address:", err)
		return err
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println("Error dialing UDP:", err)
		return err
	}
	defer conn.Close()

	numPackets, packets := calculatePackets(data)
	newPort, err := handleAgreement(conn, numPackets)
	if err != nil {
		return err
	}

	// Create a new UDP connection to the server using the new port
	newServerAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", addr, newPort))
	if err != nil {
		fmt.Println("Error resolving new server address:", err)
		return err
	}

	newConn, err := net.DialUDP("udp", nil, newServerAddr)
	if err != nil {
		fmt.Println("Error dialing new UDP connection:", err)
		return err
	}
	defer newConn.Close()

	err = sendPackets(newConn, packets)
	if err != nil {
		return err
	}
	err = awaitAcknowledgment(newConn, packets)
	if err != nil {
		return err
	}

	return nil
}

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

func handleAgreement(conn *net.UDPConn, numPackets int) (int, error) {
	agreementPacket := []byte(fmt.Sprintf(AGREEMENT, numPackets))
	retransmits := 0

	for retransmits < MAX_RETRANSMIT {
		conn.Write(agreementPacket)

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

func awaitAcknowledgment(conn *net.UDPConn, packets [][]byte) error {
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
