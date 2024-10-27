// -------------------------------------------- LICENSE --------------------------------------------
//
// Copyright 2024 Ana Pires, José Matos, Rúben Oliveira
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// -------------------------------------------------------------------------------------------------

package message

type Task struct {
	TaskID    string `json:"task_id"`
	Frequency int    `json:"frequency"`
	Targets   []struct {
		TargetID  string `json:"target_id"`
		IPAddress string `json:"ip_address"`
		UDPPort   string `json:"udp_port"`
		TCPPort   string `json:"tcp_port"`
	} `json:"targets"`
	DeviceOptions []struct {
		DeviceID       string `json:"device_id"`
		IPAddress      string `json:"ip_address"`
		UDPPort        string `json:"udp_port"`
		TCPPort        string `json:"tcp_port"`
		MonitorOptions struct {
			MonitorCPU bool     `json:"monitor_cpu"`
			MonitorRAM bool     `json:"monitor_ram"`
			Interfaces []string `json:"intercaces"`
		} `json:"monitor_options"`
		LinkOptions struct {
			Bandwidth struct {
				Protocol        string `json:"protocol"`
				Duration        int    `json:"duration"`
				ClientPort      int    `json:"client_port"`
				ParallelStreams int    `json:"parallel_streams"`
				Interval        int    `json:"interval"`
			} `json:"bandwidth"`
			Jitter struct {
				Protocol   string `json:"protocol"`
				Bandwidth  string `json:"bandwidth"`
				Duration   int    `json:"duration"`
				ClientPort int    `json:"client_port"`
				PacketSize int    `json:"packet_size"`
				Interval   int    `json:"interval"`
			} `json:"jitter"`
			PacketLoss struct {
				Protocol   string `json:"protocol"`
				Bandwidth  string `json:"bandwidth"`
				Duration   int    `json:"duration"`
				ClientPort int    `json:"client_port"`
				PacketSize int    `json:"packet_size"`
				Interval   int    `json:"interval"`
			} `json:"packet_loss"`
			Latency struct {
				PacketCount int `json:"packet_count"`
				PacketSize  int `json:"packet_size"`
				Interval    int `json:"interval"`
				Timeout     int `json:"timeout"`
			} `json:"latency"`
		} `json:"link_options"`
		AlertFlowConditions struct {
			CPUUsage       float64 `json:"cpu_usage"`
			RAMUsage       float64 `json:"ram_usage"`
			PacketLoss     float64 `json:"packet_loss"`
			InterfaceStats int     `json:"interface_stats"`
			Jitter         int     `json:"jitter"`
		} `json:"alertflow_conditions"`
	} `json:"device_options"`
}
