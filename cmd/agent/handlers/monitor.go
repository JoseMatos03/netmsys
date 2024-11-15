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
	"fmt"
	"log"
	"netmsys/cmd/message"
	"netmsys/pkg/nettsk"
	"os/exec"
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

	for range ticker.C {
		iteration++

		// Monitor CPU usage if needed
		if task.MonitorOptions.MonitorCPU {
			go monitorCPU(task.AlertFlowConditions.CPUUsage)
		}

		// Monitor RAM usage if needed
		if task.MonitorOptions.MonitorRAM {
			go monitorRAM(task.AlertFlowConditions.RAMUsage)
		}

		// Monitor interfaces
		for _, iface := range task.MonitorOptions.Interfaces {
			go monitorInterface(iface, task.AlertFlowConditions.InterfaceStats, task.AlertFlowConditions.Jitter, task.AlertFlowConditions.PacketLoss)
		}

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
	parsedMessage, errParser := message.ParseIperfOutput(string(output))
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
	parsedMessage, errParser := message.ParseIperfOutput(string(output))
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
	parsedMessage, errParser := message.ParseIperfOutput(string(output))
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
func monitorCPU(usage float64) {
	// Get CPU usage percentage over 1 second
	usagePercentages, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		fmt.Printf("Failed to monitor CPU: %v\n", err)
		return
	}

	if usagePercentages[0] >= usage {
		// TODO SEND ALERT
	}
}

// monitorRAM checks the current RAM usage and prints/logs the data
func monitorRAM(usage float64) {
	// Get memory usage statistics
	memStats, err := mem.VirtualMemory()
	if err != nil {
		fmt.Printf("Failed to monitor RAM: %v\n", err)
		return
	}

	if memStats.UsedPercent >= usage {
		// TODO SEND ALERT
	}
}

func monitorInterface(iface string, stats int, jitter int, packetloss float64) {
	// Get network I/O stats
	_, err := net.IOCounters(true)
	if err != nil {
		log.Printf("Failed to get I/O stats for interface %s: %v", iface, err)
	}

	// TODO compare Packets per second, jitter and packet loss
	// THEN send an alert
}
