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
	TimeoutDuration = 5 * time.Second
	MaxRetransmits  = 3
	Agreement       = "AGREEMENT|%d"
	AckAgreement    = "ACK_AGREEMENT"
	AckComplete     = "ACK_COMPLETE"
	Recovery        = "RECOVERY|%d"
	FastRecovery    = "FAST_RECOVERY|%d" // Define fast recovery signal
)

// Send function sends data to the specified address and port over UDP with packet handling.
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

func sendAgreement(conn *net.UDPConn, numPackets int) bool {
	agreementPacket := []byte(fmt.Sprintf(Agreement, numPackets))
	ackReceived := false
	retransmits := 0

	for !ackReceived && retransmits < MaxRetransmits {
		conn.Write(agreementPacket)
		fmt.Printf("Sent agreement packet, expecting to send %d packets.\n", numPackets)

		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("Timeout waiting for ACK. Retransmitting (%d/%d)\n", retransmits+1, MaxRetransmits)
			retransmits++
			continue
		}

		if string(buf[:n]) == AckAgreement {
			fmt.Println("Received ACK for agreement. Sending all packets.")
			ackReceived = true
		}
	}

	if !ackReceived {
		fmt.Println("Failed to receive ACK after 3 attempts. Terminating connection.")
		return false
	}
	return true
}

func sendPackets(conn *net.UDPConn, packets [][]byte) {
	for i, packet := range packets {
		packetWithSeq := fmt.Sprintf("%d|", i) + string(packet)
		conn.Write([]byte(packetWithSeq))
		fmt.Printf("Sent packet %d/%d\n", i+1, len(packets))
	}
}

// The Send function remains the same until awaitAcknowledgment()
func awaitAcknowledgment(conn *net.UDPConn, packets [][]byte) {
	buf := make([]byte, 1024)

	for {
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

// Modified retransmitPackets function to handle both recovery modes
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
