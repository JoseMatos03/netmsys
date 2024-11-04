package nettsk

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	RECOVERY       = "RECOVERY|%d"
	FAST_RECOVERY  = "FAST_RECOVERY|%d"
	ACK_AGREEMENT  = "ACK_AGREEMENT"
	ACK_COMPLETE   = "ACK_COMPLETE"
	MAX_RETRANSMIT = 3
	TIMEOUT        = 5 * time.Second
)

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

// Establishes agreement by receiving the number of packets to be sent
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

// Receives packets with timeout handling using a ticker
func receivePackets(conn *net.UDPConn, clientAddr *net.UDPAddr, numPackets int, receivedPackets map[int][]byte) (int, error) {
	buf := make([]byte, 1024)
	receivedSeq := -1
	expectedSeq := 0
	retransmitCount := 0

	// Set up a ticker to act as a timeout
	ticker := time.NewTicker(TIMEOUT)
	defer ticker.Stop()

	for receivedSeq < numPackets-1 && retransmitCount < MAX_RETRANSMIT {
		select {
		case <-ticker.C:
			// Timeout occurred: Send FAST_RECOVERY and increase retransmit count
			fmt.Printf("Timeout at packet %d. Sending FAST_RECOVERY.\n", expectedSeq)
			conn.WriteToUDP([]byte(fmt.Sprintf(FAST_RECOVERY, expectedSeq)), clientAddr)
			retransmitCount++

			if retransmitCount >= MAX_RETRANSMIT {
				return expectedSeq, fmt.Errorf("maximum retransmits reached, terminating")
			}

		default:
			conn.SetReadDeadline(time.Now().Add(TIMEOUT)) // Optional for blocking `ReadFromUDP`
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue // Continue to the next select cycle if no data received
			}

			// Reset the ticker on successful packet receipt
			ticker.Reset(TIMEOUT)
			retransmitCount = 0 // Reset retransmit count on successful packet receipt

			// Process the received packet
			packet := string(buf[:n])
			parts := strings.SplitN(packet, "|", 2)
			seqNum, _ := strconv.Atoi(parts[0])

			if seqNum == expectedSeq {
				receivedPackets[seqNum] = []byte(parts[1])
				receivedSeq = seqNum
				expectedSeq++
				fmt.Printf("Received in-order packet %d/%d\n", seqNum+1, numPackets)
			} else if seqNum > expectedSeq {
				receivedSeq = seqNum
				receivedPackets[seqNum] = []byte(parts[1])
				fmt.Printf("Out-of-order packet %d received. Expecting %d.\n", seqNum, expectedSeq)
			}
		}
	}

	if receivedSeq < numPackets-1 {
		return expectedSeq, fmt.Errorf("failed to receive all packets")
	}
	return expectedSeq, nil
}

// Sends RECOVERY packets for missing packets and waits for retransmissions
func handleMissingPackets(conn *net.UDPConn, clientAddr *net.UDPAddr, numPackets int, receivedPackets map[int][]byte) error {
	missingPackets := identifyMissingPackets(receivedPackets, numPackets)
	if len(missingPackets) > 0 {
		fmt.Println("Missing packets detected. Sending RECOVERY requests.")
		for _, missingSeq := range missingPackets {
			conn.WriteToUDP([]byte(fmt.Sprintf(RECOVERY, missingSeq)), clientAddr)
			buf := make([]byte, 1024)
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				return fmt.Errorf("error receiving missing packet %d: %v", missingSeq, err)
			}
			receivedPackets[missingSeq] = buf[:n]
		}
	}
	return nil
}

// Finalizes the transmission by sending ACK_COMPLETE and reassembling data
func finalizeTransmission(conn *net.UDPConn, clientAddr *net.UDPAddr, dataChannel chan<- []byte, receivedPackets map[int][]byte, numPackets int) {
	if len(receivedPackets) == numPackets {
		conn.WriteToUDP([]byte(ACK_COMPLETE), clientAddr)
		fmt.Println("Final ACK sent. Transmission complete.")
		dataChannel <- reassembleMessage(receivedPackets)
	} else {
		fmt.Println("Failed to receive all packets.")
	}
}

// Helper Function: Identify missing packets in sequence
func identifyMissingPackets(packets map[int][]byte, total int) []int {
	var missing []int
	for i := 0; i < total; i++ {
		if _, exists := packets[i]; !exists {
			missing = append(missing, i)
		}
	}
	return missing
}

// Helper Function: Reassemble packets into a full message
func reassembleMessage(packets map[int][]byte) []byte {
	var messageBuilder strings.Builder
	for i := 0; i < len(packets); i++ {
		messageBuilder.Write(packets[i])
	}
	return []byte(messageBuilder.String())
}
