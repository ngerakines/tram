package config

import (
	"github.com/ngerakines/preview/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

type appConfigError struct {
	message string
}

type AppConfig interface {
	Listen() string
	LruSize() uint64
	Storage() StorageAppConfig
	Index() IndexAppConfig
	Source() string
}

type IndexAppConfig interface {
	Engine() string
	LocalBasePath() string
}

type StorageAppConfig interface {
	Engine() string
	BasePath() string
	S3Key() (string, error)
	S3Secret() (string, error)
	S3Buckets() ([]string, error)
	S3Host() (string, error)
	S3VerifySsl() (bool, error)
}

func LoadAppConfig(givenPath string) (AppConfig, error) {
	configPath := determineConfigPath(givenPath)
	if configPath == "" {
		return NewDefaultAppConfig()
	}
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	return NewUserAppConfig(content)
}

func (err appConfigError) Error() string {
	return err.message
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
