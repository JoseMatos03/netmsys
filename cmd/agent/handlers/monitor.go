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
	"netmsys/cmd/message"
	"netmsys/pkg/nettsk"
	"os/exec"
	"sync"
	"time"
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
	defer ticker.Stop()

	for range ticker.C {
		for _, device := range task.DeviceOptions {
			fmt.Printf("Running task for device: %s\n", device.DeviceID)

			// Monitor CPU usage if needed
			if device.MonitorOptions.MonitorCPU {
				go monitorCPU(device)
			}

			// Monitor RAM usage if needed
			if device.MonitorOptions.MonitorRAM {
				go monitorRAM(device)
			}

			// Monitor interfaces
			for _, iface := range device.MonitorOptions.Interfaces {
				go monitorInterface(iface, device)
			}

			// Run bandwidth tests
			go a.runBandwidthTest(device.LinkOptions.Bandwidth, device.IPAddress)

			// Run jitter tests
			go a.runJitterTest(device.LinkOptions.Jitter, device.IPAddress)

			// Run packet loss tests
			go a.runPacketLossTest(device.LinkOptions.PacketLoss, device.IPAddress)

			// Run latency tests
			go a.runLatencyTest(device.LinkOptions.Latency, device.IPAddress)
		}
	}
}

// Below are functions that initiate the "iperf" and "ping" commands

func (a *Agent) runBandwidthTest(options BandwidthOptions, targetIP string) {
	cmd := exec.Command("iperf", "-c", targetIP, "-p", fmt.Sprintf("%d", options.ClientPort),
		"-t", fmt.Sprintf("%d", options.Duration), "-P", fmt.Sprintf("%d", options.ParallelStreams),
		"-i", fmt.Sprintf("%d", options.Interval))
	if options.Protocol == "udp" {
		cmd.Args = append(cmd.Args, "-u")
	}
	a.runCommand(cmd, "Bandwidth Test")
}

func (a *Agent) runJitterTest(options JitterOptions, targetIP string) {
	cmd := exec.Command("iperf", "-c", targetIP, "-u", "-b", options.Bandwidth,
		"-t", fmt.Sprintf("%d", options.Duration), "-p", fmt.Sprintf("%d", options.ClientPort),
		"-l", fmt.Sprintf("%d", options.PacketSize), "-i", fmt.Sprintf("%d", options.Interval))
	a.runCommand(cmd, "Jitter Test")
}

func (a *Agent) runPacketLossTest(options PacketLossOptions, targetIP string) {
	cmd := exec.Command("iperf", "-c", targetIP, "-u", "-b", options.Bandwidth,
		"-t", fmt.Sprintf("%d", options.Duration), "-p", fmt.Sprintf("%d", options.ClientPort),
		"-l", fmt.Sprintf("%d", options.PacketSize), "-i", fmt.Sprintf("%d", options.Interval))
	a.runCommand(cmd, "Packet Loss Test")
}

func (a *Agent) runLatencyTest(options LatencyOptions, targetIP string) {
	cmd := exec.Command("ping", "-c", fmt.Sprintf("%d", options.PacketCount), "-s", fmt.Sprintf("%d", options.PacketSize),
		"-i", fmt.Sprintf("%d", options.Interval), "-W", fmt.Sprintf("%d", options.Timeout), targetIP)
	a.runCommand(cmd, "Latency Test")
}

func monitorCPU(device DeviceOptions) {
	// Simulate CPU monitoring logic
}

func monitorRAM(device DeviceOptions) {
	// Simulate RAM monitoring logic
}

func monitorInterface(iface string, device DeviceOptions) {
	// Simulate network interface monitoring logic
}

func (a *Agent) runCommand(cmd *exec.Cmd, testName string) {
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s failed: %v\n", testName, err)
		return
	}

	outputMessage := "OUTPUT|" + string(output)
	nettsk.Send(a.ServerAddr, a.UDPPort, []byte(outputMessage))
}
