// ------------------------------------ LICENSE ------------------------------------
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
// ---------------------------------------------------------------------------------

// Package parsers provides methods for parsing external data.
// This module provices specific parsing methods for JSON data.
package parsers

import (
	"encoding/json"
	"os"
)

// ReadJSONFile reads a JSON file from the specified filename and parses it into the struct passed as 'result'.
// The result parameter must be a pointer to the struct that matches the JSON format.
func ReadJSONFile(filename string, result interface{}) error {
	// Read the JSON file using os.ReadFile
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Parse the JSON into the provided struct
	err = json.Unmarshal(data, result)
	if err != nil {
		return err
	}

	return nil
}

// WriteJSONFile serializes the provided struct 'data' into JSON format and writes it to the specified filename.
// The JSON file will be created or overwritten if it already exists.
func WriteJSONFile(filename string, data interface{}) error {
	// Serialize the data into JSON format with pretty printing (using MarshalIndent)
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON data to the file with appropriate permissions (0644)
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// SerializeJSONFile reads a JSON file from the specified filename and returns its contents as a byte slice.
// It is useful for sending or handling raw JSON data without parsing it.
func SerializeJSONFile(filename string) ([]byte, error) {
	// Read the JSON file using os.ReadFile
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// DeserializeJSONFile takes a byte slice of JSON data and writes it to the specified filename.
// It creates a new JSON file or overwrites an existing one with the provided data.
func DeserializeJSONFile(filename string, input []byte) error {
	// Write the input JSON data to the specified file
	err := os.WriteFile(filename, input, 0644) // Set appropriate file permissions
	if err != nil {
		return err
	}

	return nil
}

// SerializeJSON serializes the provided struct 'input' into a JSON byte slice.
// It is useful for transmitting or storing data in JSON format.
func SerializeJSON(input interface{}) ([]byte, error) {
	// Serialize the task struct into JSON format
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	return jsonData, err
}

// DeserializeJSON parses the given JSON byte slice into the provided struct 'result'.
// The result parameter must be a pointer to the struct that will hold the parsed data.
func DeserializeJSON(input []byte, result interface{}) error {
	// Deserialize the JSON data into the provided struct
	err := json.Unmarshal(input, result)
	if err != nil {
		return err
	}

	return nil
}
