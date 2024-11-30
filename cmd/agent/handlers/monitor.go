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

var (
	tasks      = []Task{}              // Slice of tasks to be processed
	taskChan   = make(chan Task)       // Channel to add new tasks
	runningMap = make(map[string]bool) // Map to track running tasks
	mutex      sync.Mutex              // Mutex to protect the runningMap
)

// RunAgentTasks manages the execution of agent tasks by processing tasks from a shared channel,
// ensuring tasks are not duplicated, and handling concurrency. It marks tasks as running,
// executes them, and cleans up once completed.
//
// Behavior:
//   - Launches a goroutine to listen for tasks from a channel and runs them.
//   - Prevents duplication of tasks by using a shared map to track running tasks.
//   - Uses a WaitGroup to ensure all tasks complete before exiting.
//
// Note:
//   - Tasks are executed concurrently, and task lifecycle management is handled automatically.
//   - The function runs indefinitely, processing tasks as they are received in the channel.
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

// AddTask sends a new task to the task channel for execution.
//
// Parameters:
//   - newTask: The task to be added to the execution queue.
//
// Behavior:
//   - Queues the new task for processing by the RunAgentTasks function.
//
// Note:
//   - Tasks added using this function are handled asynchronously.
func (a *Agent) AddTask(newTask Task) {
	taskChan <- newTask
}

// runTask executes a single task, handling its periodic execution and associated monitoring activities.
//
// Parameters:
//   - task: The Task object containing configuration details for execution.
//
// Behavior:
//   - Schedules and runs periodic executions of device tests (bandwidth, jitter, packet loss, latency).
//   - Starts monitoring for CPU, RAM, and interface statistics as defined in the task configuration.
//   - Ensures task execution is tied to a frequency and handles lifecycle management.
//
// Note:
//   - Monitoring and testing are handled concurrently using goroutines.
//   - Periodic task execution uses a ticker based on the specified frequency.
//   - Ensures monitoring and testing are scoped to each device in the task.
func (a *Agent) runTask(task Task) {
	fmt.Printf("Starting task: %s, running every %d seconds\n", task.TaskID, task.Frequency)
	ticker := time.NewTicker(time.Duration(task.Frequency) * time.Second)
	iteration := 1
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

	// Function to execute tests
	executeTests := func() {
		for _, device := range task.DeviceOptions {
			fmt.Printf("Running task for device: %s\n", device.DeviceID)

			var wg sync.WaitGroup
			wg.Add(4)

			// Run bandwidth test
			go func() {
				defer wg.Done()
				a.runBandwidthTest(device.LinkOptions.Bandwidth, device.IPAddress, task.TaskID, iteration)
			}()

			// Run jitter test
			go func() {
				defer wg.Done()
				a.runJitterTest(device.LinkOptions.Jitter, device.IPAddress, task.TaskID, iteration)
			}()

			go func() {
				defer wg.Done()
				// Run packet loss test
				a.runPacketLossTest(device.LinkOptions.PacketLoss, device.IPAddress, task.TaskID, iteration)
			}()

			go func() {
				defer wg.Done()
				// Run latency test
				a.runLatencyTest(device.LinkOptions.Latency, device.IPAddress, task.TaskID, iteration)
			}()

			wg.Wait()
		}
	}

	// Execute tests instantly
	executeTests()

	// Frequency loop of executions
	for range ticker.C {
		iteration++
		executeTests()
	}
}

// Below are the functions that execute each of the tests with the given configurations.

// runBandwidthTest executes a bandwidth test using iperf for the specified target IP and task configuration,
// parses the output, and sends the results to the server.
//
// Behavior:
//   - Constructs and executes an iperf command based on the specified options.
//   - Parses the command output to extract bandwidth data.
//   - Sends the parsed output to the server using a UDP connection.
//
// Parameters:
//   - options: Configuration options for the bandwidth test (duration, protocol, ports, etc.).
//   - targetIP: The IP address of the target device.
//   - taskID: The identifier of the task associated with the test.
//   - iteration: The current iteration number of the task.
//
// Returns:
//   - An error if the command execution, parsing, or sending of the output fails, otherwise nil.
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

// runJitterTest performs a jitter test using iperf with the given configuration,
// parses the output, and sends the results to the server.
//
// Behavior:
//   - Executes an iperf command configured for jitter testing.
//   - Parses the output to extract jitter metrics.
//   - Sends the parsed results to the server via a UDP connection.
//
// Parameters:
//   - options: Configuration options for the jitter test (duration, bandwidth, packet size, etc.).
//   - targetIP: The IP address of the target device.
//   - taskID: The identifier of the task associated with the test.
//   - iteration: The current iteration number of the task.
//
// Returns:
//   - An error if the command execution, parsing, or sending of the output fails, otherwise nil.
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

// runPacketLossTest executes a packet loss test using iperf with the specified options,
// parses the results, and sends the metrics to the server.
//
// Behavior:
//   - Runs an iperf command configured to measure packet loss.
//   - Extracts packet loss data from the command output.
//   - Sends the parsed data to the server over a UDP connection.
//
// Parameters:
//   - options: Configuration options for the packet loss test (bandwidth, duration, packet size, etc.).
//   - targetIP: The IP address of the target device.
//   - taskID: The identifier of the task associated with the test.
//   - iteration: The current iteration number of the task.
//
// Returns:
//   - An error if the command execution, parsing, or sending of the output fails, otherwise nil.
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

// runLatencyTest measures latency by executing a ping command based on the provided configuration,
// parses the output, and sends the latency data to the server.
//
// Behavior:
//   - Constructs and runs a ping command based on the specified options.
//   - Parses the ping output to extract latency metrics.
//   - Sends the parsed latency data to the server using a UDP connection.
//
// Parameters:
//   - options: Configuration options for the latency test (packet size, interval, timeout, etc.).
//   - targetIP: The IP address of the target device.
//   - taskID: The identifier of the task associated with the test.
//   - iteration: The current iteration number of the task.
//
// Returns:
//   - An error if the command execution, output parsing, or sending of the metrics fails, otherwise nil.
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

// monitorCPU continuously monitors CPU usage at regular intervals and sends an alert if usage exceeds a specified threshold.
//
// Parameters:
//   - usage: The CPU usage threshold percentage (e.g., 80.0 for 80%).
//   - frequency: The frequency in seconds at which to check the CPU usage.
//
// Behavior:
//   - Periodically checks CPU usage using the "github.com/shirou/gopsutil/cpu" package.
//   - If CPU usage exceeds the specified threshold, sends an alert message to the server using the TCP protocol.
//   - Terminates the monitoring process if an error occurs during CPU usage retrieval.
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

// monitorRAM continuously monitors RAM usage at regular intervals and sends an alert if usage exceeds a specified threshold.
//
// Parameters:
//   - usage: The RAM usage threshold percentage (e.g., 75.0 for 75%).
//   - frequency: The frequency in seconds at which to check the RAM usage.
//
// Behavior:
//   - Periodically checks RAM usage using the "github.com/shirou/gopsutil/mem" package.
//   - If RAM usage exceeds the specified threshold, sends an alert message to the server using the TCP protocol.
//   - Terminates the monitoring process if an error occurs during memory usage retrieval.
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

// monitorInterface monitors network performance on a specified interface, including packet rates, jitter, and packet loss.
//
// Parameters:
//   - iface: The name of the network interface to monitor (e.g., "eth0").
//   - ifaceIP: The IP address associated with the interface for testing with Iperf.
//   - stats: Threshold for packets per second (PPS).
//   - jitterThreshold: Maximum allowable jitter in milliseconds.
//   - packetLossThreshold: Maximum allowable packet loss percentage.
//   - frequency: The frequency in seconds at which to monitor network performance.
//
// Behavior:
//   - Periodically retrieves network I/O stats using the "github.com/shirou/gopsutil/net" package.
//   - Computes packets sent/received per second and checks if they exceed the specified threshold.
//   - Uses Iperf to measure jitter and packet loss and validates them against specified thresholds.
//   - Sends alerts to the server if any of the metrics exceed the specified thresholds.
//   - Updates previous stats for incremental monitoring.
func (a *Agent) monitorInterface(iface string, ifaceIP string, stats int, jitterThreshold int, packetLossThreshold float64, frequency int) {
	ticker := time.NewTicker(time.Duration(frequency) * time.Second)
	defer ticker.Stop()

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

		packetsSent := currentIO.PacketsSent

		// Check if PPS exceeds the threshold
		if packetsSent > uint64(stats) {
			ppsAlert := fmt.Sprintf("ALERT: High packets per second on %s: Sent=%d", iface, packetsSent)
			alrtflw.Send(a.ServerAddr, a.TCPPort, []byte(ppsAlert))
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
	}
}

// runIperfForInterface executes an Iperf test on the specified interface IP to measure jitter and packet loss.
//
// Parameters:
//   - ifaceIP: The IP address of the target interface for the test.
//   - duration: The duration of the Iperf test in seconds.
//
// Returns:
//   - jitter: The measured jitter in milliseconds.
//   - packetLoss: The measured packet loss percentage.
//   - error: An error if the command execution or output parsing fails.
//
// Behavior:
//   - Constructs and executes an Iperf command for UDP testing.
//   - Parses the command output using regex patterns to extract jitter and packet loss metrics.
//   - Returns the parsed metrics or an error if parsing fails.
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
		log.Printf("Failed to parse jitter from Iperf output")
		return 0, 0, fmt.Errorf("failed to parse jitter")
	}
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
	packetLoss, err := strconv.ParseFloat(packetLossMatch[3], 64)
	if err != nil {
		return 0, 0, err
	}

	return jitter, packetLoss, nil
}
