// ------------------------------------ LICENSE ------------------------------------
//
// # Copyright 2024 Ana Pires, José Matos, Rúben Oliveira
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// ---------------------------------------------------------------------------------
package handlers

import (
	"bytes"
	"fmt"
	"log"
	"netmsys/cmd/message"
	"netmsys/pkg/alrtflw"
	"netmsys/pkg/nettsk"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type Task = message.Task
type DeviceOptions = message.DeviceOptions
type MonitorOptions = message.MonitorOptions
type LinkOptions = message.LinkOptions
type BandwidthOptions = message.BandwidthOptions
type JitterOptions = message.JitterOptions
type PacketLossOptions = message.PacketLossOptions
type LatencyOptions = message.LatencyOptions
type AlertFlowConditions = message.AlertFlowConditions

// RunAgentTasks starts monitoring tasks in separate goroutines
var (
	tasks      = []Task{}              // Slice of tasks to be processed
	taskChan   = make(chan Task)       // Channel to add new tasks
	runningMap = make(map[string]bool) // Map to track running tasks
	mutex      sync.Mutex              // Mutex to protect the runningMap
)

func (a *Agent) RunAgentTasks() {
	var wg sync.WaitGroup

	// Goroutine to listen for new tasks and run them
	go func() {
		for task := range taskChan {
			mutex.Lock()
			if runningMap[task.TaskID] {
				// Task is already running, skip to avoid duplication
				mutex.Unlock()
				continue
			}
			// Mark the task as running
			runningMap[task.TaskID] = true
			mutex.Unlock()

			// Add to WaitGroup and run the task
			wg.Add(1)
			go func(t Task) {
				defer wg.Done()
				a.runTask(t)

				// Mark task as completed
				mutex.Lock()
				delete(runningMap, t.TaskID)
				mutex.Unlock()
			}(task)
		}
	}()

	// Initial load of tasks
	for _, task := range tasks {
		taskChan <- task
	}

	// Wait for all tasks to finish before exiting
	wg.Wait()
}

// Function to add a new task
func (a *Agent) AddTask(newTask Task) {
	taskChan <- newTask
}

// runTask handles the execution of a single task
func (a *Agent) runTask(task Task) {
	fmt.Printf("Starting task: %s, running every %d seconds\n", task.TaskID, task.Frequency)
	ticker := time.NewTicker(time.Duration(task.Frequency) * time.Second)
	iteration := 0
	defer ticker.Stop()

	// Monitor CPU usage if needed
	if task.MonitorOptions.MonitorCPU {
		go a.monitorCPU(task.AlertFlowConditions.CPUUsage, task.Frequency)
	}

	// Monitor RAM usage if needed
	if task.MonitorOptions.MonitorRAM {
		go a.monitorRAM(task.AlertFlowConditions.RAMUsage, task.Frequency)
	}

	// Monitor interfaces
	for _, iface := range task.MonitorOptions.Interfaces {
		go a.monitorInterface(iface.Name, iface.IP, task.AlertFlowConditions.InterfaceStats, task.AlertFlowConditions.Jitter, task.AlertFlowConditions.PacketLoss, task.Frequency)
	}

	for range ticker.C {
		iteration++

		for _, device := range task.DeviceOptions {
			fmt.Printf("Running task for device: %s\n", device.DeviceID)

			// Run bandwidth tests
			go a.runBandwidthTest(device.LinkOptions.Bandwidth, device.IPAddress, task.TaskID, iteration)

			// Run jitter tests
			go a.runJitterTest(device.LinkOptions.Jitter, device.IPAddress, task.TaskID, iteration)

			// Run packet loss tests
			go a.runPacketLossTest(device.LinkOptions.PacketLoss, device.IPAddress, task.TaskID, iteration)

			// Run latency tests
			go a.runLatencyTest(device.LinkOptions.Latency, device.IPAddress, task.TaskID, iteration)
		}
	}
}

// Below are functions that initiate the "iperf" and "ping" commands

func (a *Agent) runBandwidthTest(options BandwidthOptions, targetIP string, taskID string, iteration int) error {
	cmd := exec.Command("iperf", "-c", targetIP, "-p", fmt.Sprintf("%d", options.ClientPort),
		"-t", fmt.Sprintf("%d", options.Duration), "-P", fmt.Sprintf("%d", options.ParallelStreams),
		"-i", fmt.Sprintf("%d", options.Interval))
	if options.Protocol == "udp" {
		cmd.Args = append(cmd.Args, "-u")
	}

	// Execute the ping command
	output, errCmd := cmd.CombinedOutput()
	if errCmd != nil {
		return errCmd
	}

	// Parse the command output
	parsedMessage := message.ParseBandwidthOutput(string(output))

	// Send the command output
	outputMessage := fmt.Sprintf("OUTPUT-%s-I%d|%s", taskID, iteration, parsedMessage)
	err := nettsk.Send(a.ServerAddr, a.UDPPort, []byte(outputMessage))
	if err != nil {
		return err
	}
	return nil
}

func (a *Agent) runJitterTest(options JitterOptions, targetIP string, taskID string, iteration int) error {
	cmd := exec.Command("iperf", "-c", targetIP, "-u", "-b", options.Bandwidth,
		"-t", fmt.Sprintf("%d", options.Duration), "-p", fmt.Sprintf("%d", options.ClientPort),
		"-l", fmt.Sprintf("%d", options.PacketSize), "-i", fmt.Sprintf("%d", options.Interval))

	// Execute the ping command
	output, errCmd := cmd.CombinedOutput()
	if errCmd != nil {
		return errCmd
	}

	// Parse the command output
	parsedMessage := message.ParseJitterData(string(output))

	// Send the command output
	outputMessage := fmt.Sprintf("OUTPUT-%s-I%d|%s", taskID, iteration, parsedMessage)
	err := nettsk.Send(a.ServerAddr, a.UDPPort, []byte(outputMessage))
	if err != nil {
		return err
	}
	return nil
}

func (a *Agent) runPacketLossTest(options PacketLossOptions, targetIP string, taskID string, iteration int) error {
	cmd := exec.Command("iperf", "-c", targetIP, "-u", "-b", options.Bandwidth,
		"-t", fmt.Sprintf("%d", options.Duration), "-p", fmt.Sprintf("%d", options.ClientPort),
		"-l", fmt.Sprintf("%d", options.PacketSize), "-i", fmt.Sprintf("%d", options.Interval))

	// Execute the ping command
	output, errCmd := cmd.CombinedOutput()
	if errCmd != nil {
		return errCmd
	}

	// Parse the command output
	parsedMessage := message.ParsePacketLossData(string(output))

	// Send the command output
	outputMessage := fmt.Sprintf("OUTPUT-%s-I%d|%s", taskID, iteration, parsedMessage)
	err := nettsk.Send(a.ServerAddr, a.UDPPort, []byte(outputMessage))
	if err != nil {
		return err
	}
	return nil
}

func (a *Agent) runLatencyTest(options LatencyOptions, targetIP string, taskID string, iteration int) error {
	cmd := exec.Command("ping", "-c", fmt.Sprintf("%d", options.PacketCount), "-s", fmt.Sprintf("%d", options.PacketSize),
		"-i", fmt.Sprintf("%d", options.Interval), "-W", fmt.Sprintf("%d", options.Timeout), targetIP)

	// Execute the ping command
	output, errCmd := cmd.CombinedOutput()
	if errCmd != nil {
		return errCmd
	}

	// Parse the command output
	parsedMessage, errParser := message.ParsePingOutput(string(output))
	if errParser != nil {
		return errParser
	}

	// Send the command output
	outputMessage := fmt.Sprintf("OUTPUT-%s-I%d|%s", taskID, iteration, parsedMessage)
	err := nettsk.Send(a.ServerAddr, a.UDPPort, []byte(outputMessage))
	if err != nil {
		return err
	}
	return nil
}

// monitorCPU checks the current CPU usage and prints/logs the data
func (a *Agent) monitorCPU(usage float64, frequency int) {
	ticker := time.NewTicker(time.Duration(frequency) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Get CPU usage percentage over 1 second
		usagePercentages, err := cpu.Percent(1*time.Second, false)
		if err != nil {
			fmt.Printf("Failed to monitor CPU: %v\n", err)
			return
		}

		if usagePercentages[0] >= usage {
			cpuAlert := []byte("CPU usage above desired values.")
			alrtflw.Send(a.ServerAddr, a.TCPPort, cpuAlert)
		}
	}
}

// monitorRAM checks the current RAM usage and prints/logs the data
func (a *Agent) monitorRAM(usage float64, frequency int) {
	ticker := time.NewTicker(time.Duration(frequency) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Get memory usage statistics
		memStats, err := mem.VirtualMemory()
		if err != nil {
			fmt.Printf("Failed to monitor RAM: %v\n", err)
			return
		}

		if memStats.UsedPercent >= usage {
			cpuAlert := []byte("RAM usage above desired values.")
			alrtflw.Send(a.ServerAddr, a.TCPPort, cpuAlert)
		}
	}
}

func (a *Agent) monitorInterface(iface string, ifaceIP string, stats int, jitterThreshold int, packetLossThreshold float64, frequency int) {
	ticker := time.NewTicker(time.Duration(frequency) * time.Second)
	defer ticker.Stop()

	var prevPacketsSent uint64
	var prevPacketsRecv uint64
	var prevTime time.Time

	for range ticker.C {
		// Get network I/O stats for the specified interface
		ioStats, err := net.IOCounters(true)
		if err != nil {
			log.Printf("Failed to get I/O stats for interface %s: %v", iface, err)
			continue
		}

		// Find the stats for the specified interface
		var currentIO *net.IOCountersStat
		for _, stat := range ioStats {
			if stat.Name == iface {
				currentIO = &stat
				break
			}
		}

		if currentIO == nil {
			log.Printf("Interface %s not found", iface)
			continue
		}

		// Calculate packets per second (PPS) if this isn't the first iteration
		if !prevTime.IsZero() {
			duration := time.Since(prevTime).Seconds()
			packetsSentPerSec := float64(currentIO.PacketsSent-prevPacketsSent) / duration
			packetsRecvPerSec := float64(currentIO.PacketsRecv-prevPacketsRecv) / duration

			// Check if PPS exceeds the threshold
			if packetsSentPerSec > float64(stats) || packetsRecvPerSec > float64(stats) {
				ppsAlert := fmt.Sprintf("ALERT: High packets per second on %s: Sent=%.2f, Recv=%.2f", iface, packetsSentPerSec, packetsRecvPerSec)
				alrtflw.Send(a.ServerAddr, a.TCPPort, []byte(ppsAlert))
			}
		}

		// Run Iperf to measure jitter and packet loss
		jitter, packetLoss, err := a.runIperfForInterface(ifaceIP, 10) // Duration set to 10 seconds
		if err != nil {
			log.Printf("Failed to run Iperf for interface %s: %v", iface, err)
			continue
		}

		// Check if jitter exceeds the threshold
		if jitter > float64(jitterThreshold) {
			jitterAlert := fmt.Sprintf("ALERT: High jitter on %s: %.2f ms", iface, jitter)
			alrtflw.Send(a.ServerAddr, a.TCPPort, []byte(jitterAlert))
		}

		// Check if packet loss exceeds the threshold
		if packetLoss > packetLossThreshold {
			packetLossAlert := fmt.Sprintf("ALERT: High packet loss on %s: %.2f%%", iface, packetLoss)
			alrtflw.Send(a.ServerAddr, a.TCPPort, []byte(packetLossAlert))
		}

		// Update previous stats and time
		prevPacketsSent = currentIO.PacketsSent
		prevPacketsRecv = currentIO.PacketsRecv
		prevTime = time.Now()
	}
}

func (a *Agent) runIperfForInterface(ifaceIP string, duration int) (float64, float64, error) {
	// Build the Iperf command
	cmd := exec.Command("iperf", "-c", ifaceIP, "-u", "-b", "1M", "-t", strconv.Itoa(duration))

	// Capture the output
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return 0, 0, err
	}

	// Get the output string
	output := out.String()

	// Updated regex patterns for jitter and packet loss
	jitterRegex := regexp.MustCompile(`(\d+\.\d+)\s*ms`)
	packetLossRegex := regexp.MustCompile(`(\d+)/(\d+)\s+\((\d+(\.\d+)?)%\)`)

	// Match and extract jitter
	jitterMatch := jitterRegex.FindStringSubmatch(output)
	if jitterMatch == nil {
		log.Printf("Failed to parse jitter from Iperf output: %s", output)
		return 0, 0, fmt.Errorf("failed to parse jitter")
	}
	fmt.Println("Jitter Match:", jitterMatch) // Debug: Print the match
	jitter, err := strconv.ParseFloat(jitterMatch[1], 64)
	if err != nil {
		return 0, 0, err
	}

	// Match and extract packet loss
	packetLossMatch := packetLossRegex.FindStringSubmatch(output)
	if packetLossMatch == nil {
		log.Printf("Failed to parse packet loss from Iperf output: %s", output)
		return 0, 0, fmt.Errorf("failed to parse packet loss")
	}
	fmt.Println("Packet Loss Match:", packetLossMatch) // Debug: Print the match
	packetLoss, err := strconv.ParseFloat(packetLossMatch[3], 64)
	if err != nil {
		return 0, 0, err
	}

	return jitter, packetLoss, nil
}
