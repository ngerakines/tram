package config

import (
	"encoding/json"
	"github.com/ngerakines/preview/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

type AppConfig struct {
	Listen  string `json:"listen"`
	LruSize uint64 `json:"lurSize"`
	Storage struct {
		Engine      string   `json:"engine"`
		BasePath    string   `json:"basePath"`
		S3Key       string   `json:"s3Key"`
		S3Secret    string   `json:"s3Secret"`
		S3Buckets   []string `json:"s3Buckets"`
		S3Host      string   `json:"s3Host"`
		S3VerifySsl bool     `json:"s3VerifySsl"`
	} `json:"storage"`
	Index struct {
		Engine        string `json:"engine"`
		LocalBasePath string `json:"localBasePath"`
	} `json:"index"`
	Source string `json:"-"`
}

func LoadAppConfig(givenPath string) (*AppConfig, error) {
	configPath := determineConfigPath(givenPath)
	if configPath == "" {
		data := NewDefaultAppConfig()
		return ParseJson([]byte(data))
	}
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	return ParseJson(data)
}

func ParseJson(data []byte) (*AppConfig, error) {
	var appConfig AppConfig
	err := json.Unmarshal(data, &appConfig)
	if err != nil {
		return nil, err
	}
	appConfig.Source = string(data)
	return &appConfig, nil
}

func determineConfigPath(givenPath string) string {
	paths := []string{
		givenPath,
		filepath.Join(util.Cwd(), "tram.conf"),
		filepath.Join(userHomeDir(), ".tram.conf"),
		"/etc/tram.conf",
	}
	for _, path := range paths {
		if util.CanLoadFile(path) {
			return path
		}
	}
	return ""
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := filepath.Join(os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"))
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
