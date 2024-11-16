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
	re := regexp.MustCompile(`\[\s*\d+\] (\d+\.\d+)-(\d+\.\d+) sec\s+([\d\.]+ [KMG]Bytes)\s+([\d\.]+ [KMG]bits/sec)`)
	matches := re.FindAllStringSubmatch(output, -1)

	formattedEntries := "----- Bandwidth Report -----\n"
	for _, match := range matches {
		interval := match[1] + "-" + match[2] + " sec"
		transfer := match[3]
		bandwidth := match[4]

		entry := fmt.Sprintf("Interval: %s, Transfer: %s, Bandwidth: %s", interval, transfer, bandwidth)
		formattedEntries += entry + "\n"
	}

	return formattedEntries
}
