package message

import (
	"fmt"
	"regexp"
)

type JitterEntry struct {
	Interval string
	Transfer string
	Jitter   string
}

func ParseJitterData(output string) string {
	// Regex to match the line that contains jitter information
	re := regexp.MustCompile(`\[\s*\d+\]\s+(\d+\.\d+-\d+\.\d+) sec\s+([\d\.]+ [KMG]Bytes)\s+[\d\.]+ [KMG]bits/sec\s+([\d\.]+ ms)\s+\d+/\d+ \(\d+%\)`)
	match := re.FindStringSubmatch(output)

	if match == nil {
		return "No jitter information found."
	}

	// Extract the required fields
	interval := match[1]
	transfer := match[2]
	jitter := match[3]

	// Format the output
	formattedEntry := "----- Jitter Report -----\n"
	formattedEntry += fmt.Sprintf("Interval: %s, Transfer: %s, Jitter: %s", interval, transfer, jitter)
	formattedEntry += "\n"

	return formattedEntry
}
