package parsers

import (
	"encoding/json"
	"os"
)

// Define a placeholder struct for the JSON structure
type TaskData struct {
	TaskID   string `json:"task_id"`
	TaskType string `json:"task_type"`
	Payload  string `json:"payload"`
}

// ReadAndParseJSON reads a JSON file and parses it into a TaskData struct
func ReadAndParseJSON(filename string) (*TaskData, error) {
	// Read the JSON file using os.ReadFile
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Parse the JSON into the TaskData struct
	var task TaskData
	err = json.Unmarshal(data, &task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}
