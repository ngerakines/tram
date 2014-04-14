package app

import (
	"encoding/json"
	"os"
)

type ConfigurationFile struct {
	Users    []string
	Groups   []string
}

func LoadConfiguration(path string) (*ConfigurationFile, error) {
	file, fileError := os.Open(path)
	if fileError != nil {
		return nil, fileError
	}
	decoder := json.NewDecoder(file)
	configuration := ConfigurationFile{}
	decodeError := decoder.Decode(&configuration)
	if decodeError != nil {
		return nil, decodeError
	}
	return &configuration, nil
}
