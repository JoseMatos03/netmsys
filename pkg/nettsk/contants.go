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

import "time"

const (
	// RECOVERY represents the format for recovery requests for a specific packet.
	RECOVERY = "RECOVERY|%d"

	// FAST_RECOVERY is the format for fast recovery requests to re-transmit missing packets.
	FAST_RECOVERY = "FAST_RECOVERY|%d"

	// AGREEMENT is the format for an agreement packet.
	AGREEMENT = "AGREEMENT|%d"

	// ACK_AGREEMENT is sent to confirm initial agreement for packet transmission.
	ACK_AGREEMENT = "ACK_AGREEMENT|%d"

	// ACK_COMPLETE is sent to indicate successful completion of transmission.
	ACK_COMPLETE = "ACK_COMPLETE"

	// MAX_RETRANSMIT defines the maximum number of retransmission attempts.
	MAX_RETRANSMIT = 5

	// TIMEOUT defines the duration to wait before declaring a packet timeout.
	TIMEOUT = 2 * time.Second

	// MAX_PACKET_SIZE defines the maximum size of each packet that is sent.
	MAX_PACKET_SIZE = 512
)
