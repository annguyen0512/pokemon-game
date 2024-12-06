package utils

import (
	"encoding/json"
	"os"
)

func SaveToFile(filePath string, data interface{}) error {
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, file, 0644)
}

func LoadFromFile(filePath string, data interface{}) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(file, data)
}
