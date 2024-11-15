package message

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Struct to hold parsed bandwidth data
type BandwidthData struct {
	ID        string
	Interval  string
	Transfer  string
	Bandwidth string
	Jitter    string
	LostTotal string
}

func ParseIperfOutput(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var data []BandwidthData
	scanner := bufio.NewScanner(file)

	// Regex pattern to match the relevant lines
	linePattern := regexp.MustCompile(`\[\s*(\d+)\] (\d+\.\d+)-(\d+\.\d+) sec\s+([\d\.]+\s\w+)\s+([\d\.]+\s\w+/sec)(?:\s+([\d\.]+ ms)\s+(\d+/\d+ \(\d+%\)))?`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := linePattern.FindStringSubmatch(line)

		if len(matches) > 0 {
			data = append(data, BandwidthData{
				ID:        matches[1],
				Interval:  fmt.Sprintf("%s-%s sec", matches[2], matches[3]),
				Transfer:  matches[4],
				Bandwidth: matches[5],
				Jitter:    matches[6],
				LostTotal: matches[7],
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	var formattedStrings []string

	for _, entry := range data {
		formatted := fmt.Sprintf("ID: %s\nInterval: %s\nTransfer: %s\nBandwidth: %s",
			entry.ID, entry.Interval, entry.Transfer, entry.Bandwidth)

		if entry.Jitter != "" && entry.LostTotal != "" {
			formatted += fmt.Sprintf("\nJitter: %s\nLost/Total: %s", entry.Jitter, entry.LostTotal)
		}

		formattedStrings = append(formattedStrings, formatted)
	}

	return strings.Join(formattedStrings, "\n\n"), nil
}
