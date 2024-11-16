package message

import (
	"fmt"
	"regexp"
)

type PacketLossEntry struct {
	Interval  string
	Transfer  string
	LostTotal string
}

func ParsePacketLossData(output string) string {
	// Regex to match the line that contains packet loss information
	re := regexp.MustCompile(`\[\s*\d+\]\s+(\d+\.\d+-\d+\.\d+) sec\s+([\d\.]+ [KMG]Bytes)\s+[\d\.]+ [KMG]bits/sec\s+[\d\.]+ ms\s+(\d+/\d+ \(\d+%\))`)
	match := re.FindStringSubmatch(output)

	if match == nil {
		return "No packet loss information found."
	}

	// Extract the required fields
	interval := match[1]
	transfer := match[2]
	lostTotal := match[3]

	// Format the output
	formattedEntry := "----- Packet Loss Report -----\n"
	formattedEntry += fmt.Sprintf("Interval: %s, Transfer: %s, Lost/Total: %s", interval, transfer, lostTotal)
	formattedEntry += "\n"

	return formattedEntry
}
