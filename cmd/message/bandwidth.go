package message

import (
	"fmt"
	"regexp"
)

type BandwidthEntry struct {
	Interval  string
	Transfer  string
	Bandwidth string
}

func ParseBandwidthOutput(output string) string {
	// Regex to match lines that contain bandwidth information
	re := regexp.MustCompile(`\[\s*\d+\]\s+(\d+\.\d+-\d+\.\d+) sec\s+([\d\.]+ [KMG]Bytes)\s+([\d\.]+ [KMG]bits/sec)\s+[\d\.]+ ms\s+\d+/\d+ \(\d+%\)`)
	match := re.FindStringSubmatch(output)

	if match == nil {
		return "No bandwitdh information found."
	}

	interval := match[1]
	transfer := match[2]
	bandwidth := match[3]

	formattedEntry := "----- Bandwidth Report -----\n"
	formattedEntry += fmt.Sprintf("Interval: %s, Transfer: %s, Bandwidth: %s", interval, transfer, bandwidth)
	formattedEntry += "\n"

	return formattedEntry
}
