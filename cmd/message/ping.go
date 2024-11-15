package message

import (
	"fmt"
	"regexp"
	"strconv"
)

type PingResult struct {
	Transmitted  int
	Received     int
	PacketLoss   float64
	TimeElapsed  int
	LatencyStats struct {
		Min  float64
		Avg  float64
		Max  float64
		Mdev float64
	}
}

func ParsePingOutput(pingOutput string) (string, error) {
	var result PingResult

	// Extract transmitted, received, packet loss, and time elapsed
	summaryRegexp := regexp.MustCompile(`(\d+) packets transmitted, (\d+) received, ([\d.]+)% packet loss, time (\d+)ms`)
	summaryMatches := summaryRegexp.FindStringSubmatch(pingOutput)
	if len(summaryMatches) == 5 {
		result.Transmitted, _ = strconv.Atoi(summaryMatches[1])
		result.Received, _ = strconv.Atoi(summaryMatches[2])
		result.PacketLoss, _ = strconv.ParseFloat(summaryMatches[3], 64)
		result.TimeElapsed, _ = strconv.Atoi(summaryMatches[4])
	}

	// Extract latency statistics (min, avg, max, mdev)
	latencyRegexp := regexp.MustCompile(`rtt min/avg/max/mdev = ([\d.]+)/([\d.]+)/([\d.]+)/([\d.]+) ms`)
	latencyMatches := latencyRegexp.FindStringSubmatch(pingOutput)
	if len(latencyMatches) == 5 {
		result.LatencyStats.Min, _ = strconv.ParseFloat(latencyMatches[1], 64)
		result.LatencyStats.Avg, _ = strconv.ParseFloat(latencyMatches[2], 64)
		result.LatencyStats.Max, _ = strconv.ParseFloat(latencyMatches[3], 64)
		result.LatencyStats.Mdev, _ = strconv.ParseFloat(latencyMatches[4], 64)
	}

	return fmt.Sprintf(
		"transmitted: %d\nreceived: %d\npacket loss: %.2f%%\ntime elapsed: %dms\nlatency (min/avg/max/mdev): %.3f/%.3f/%.3f/%.3f ms",
		result.Transmitted, result.Received, result.PacketLoss, result.TimeElapsed,
		result.LatencyStats.Min, result.LatencyStats.Avg, result.LatencyStats.Max, result.LatencyStats.Mdev,
	), nil
}
