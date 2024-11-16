package nettsk

import "time"

const (
	// RECOVERY represents the format for recovery requests for a specific packet.
	RECOVERY = "RECOVERY|%d"

	// FAST_RECOVERY is the format for fast recovery requests to re-transmit missing packets.
	FAST_RECOVERY = "FAST_RECOVERY|%d"

	// AGREEMENT is the format for an agreement packet.
	AGREEMENT = "AGREEMENT|%d"

	// ACK_AGREEMENT is sent to confirm initial agreement for packet transmission.
	ACK_AGREEMENT = "ACK_AGREEMENT"

	// ACK_COMPLETE is sent to indicate successful completion of transmission.
	ACK_COMPLETE = "ACK_COMPLETE"

	// MAX_RETRANSMIT defines the maximum number of retransmission attempts.
	MAX_RETRANSMIT = 5

	// TIMEOUT defines the duration to wait before declaring a packet timeout.
	TIMEOUT = 1 * time.Second

	// MAX_PACKET_SIZE defines the maximum size of each packet that is sent.
	MAX_PACKET_SIZE = 512
)
