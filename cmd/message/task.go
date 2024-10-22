package message

type Task struct {
	TaskID        string   `json:"task_id"`
	Frequency     int      `json:"frequency"`
	Devices       []string `json:"devices"`
	DeviceOptions []struct {
		DeviceID       string `json:"device_id"`
		IPAddress      string `json:"ip_address"`
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
