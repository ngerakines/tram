package config

import (
	"encoding/json"
	"github.com/ngerakines/tram/util"
	"io/ioutil"
	"os"
	"path/filepath"
)

type AppConfig struct {
	StorageBackend string          `json:"storageBackend"`
	LocalConfig    *LocalAppConfig `json:"local,omitempty"`
	S3Config       *S3AppConfig    `json:"s3,omitempty"`
}

type LocalAppConfig struct {
	BasePath string `json:"basePath"`
	LruSize  uint64 `json:"maxCacheSize"`
}

type S3AppConfig struct {
	Buckets []string `json:"buckets"`
}

type AppConfigError struct {
	message string
}

var defaultAppConfig = &AppConfig{
	StorageBackend: "local",
	LocalConfig: &LocalAppConfig{
		BasePath: func() string {
			pwd, err := os.Getwd()
			if err != nil {
				panic(err.Error())
			}
			cacheDirectory := filepath.Join(pwd, ".cache")
			os.MkdirAll(cacheDirectory, 00777)
			return cacheDirectory
		}(),
		LruSize: 33554432,
	},
}

func LoadAppConfig(givenPath string) (*AppConfig, error) {
	configPath := determineConfigPath(givenPath)
	if configPath == "" {
		return defaultAppConfig, nil
	}
	return NewAppConfig(configPath)
}

func NewAppConfig(path string) (*AppConfig, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var appConfig AppConfig
	err = json.Unmarshal(content, &appConfig)
	if err != nil {
		return nil, err
	}
	return &appConfig, nil
}

func (err AppConfigError) Error() string {
	return err.message
}

func determineConfigPath(givenPath string) string {
	paths := []string{
		givenPath,
		filepath.Join(util.CWD(), "tram.conf"),
		filepath.Join(util.UserHomeDir(), "tram.conf"),
		"/etc/tram.conf",
	}
	for _, path := range paths {
		if canLoadFile(path) {
			return path
		}
	}
	return ""
}

func canLoadFile(path string) bool {
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		return false
	}
	return true
}
