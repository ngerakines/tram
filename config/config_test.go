package config

import (
	"github.com/ngerakines/tram/util"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

type directoryManager struct {
	path string
}

func (dm *directoryManager) Close() {
	log.Println("Removing temp path", dm.path)
	os.RemoveAll(dm.path)
}

func newDirectoryManager() *directoryManager {
	um := util.NewUidManager()
	pwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	cacheDirectory := filepath.Join(pwd, um.GenerateHex())
	os.MkdirAll(cacheDirectory, 00777)
	return &directoryManager{cacheDirectory}
}

type tempFileManager struct {
	dm    *directoryManager
	files map[string]string
}

func (fm *tempFileManager) initFile(name, body string) {
	path := filepath.Join(fm.dm.path, util.Hash([]byte(body)))
	err := ioutil.WriteFile(path, []byte(body), 00777)
	if err != nil {
		log.Fatal(err)
	}
	fm.files[name] = path
}

func (fm *tempFileManager) get(name string) (string, error) {
	path, hasPath := fm.files[name]
	if hasPath {
		return path, nil
	}
	return "", AppConfigError{"No config file exists with that label."}
}

func (fm *tempFileManager) close() {
	fm.dm.Close()
}

func initTempFileManager() *tempFileManager {
	fm := new(tempFileManager)
	fm.dm = newDirectoryManager()
	fm.files = make(map[string]string)
	fm.initFile("storage-local", `{"storageBackend": "local", "local": {"basePath": "./", "maxCacheSize": 33554432}}`)
	fm.initFile("storage-s3", `{"storageBackend": "s3", "s3": {"buckets": ["foo", "bar"]}}`)
	return fm
}

func TestLocalFile(t *testing.T) {
	fm := initTempFileManager()
	defer fm.close()

	path, err := fm.get("storage-local")
	if err != nil {
		t.Error(err.Error())
		return
	}
	appConfig, err := LoadAppConfig(path)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if appConfig.StorageBackend != "local" {
		t.Errorf("config backend set to %s", appConfig.StorageBackend)
		return
	}
	if appConfig.LocalConfig == nil {
		t.Error("app config local config should be set.")
		return
	}
	if appConfig.S3Config != nil {
		t.Error("app config s3 config should not be set.")
		return
	}
	if appConfig.LocalConfig.BasePath != "./" {
		t.Errorf("local config base path invalid: %s", appConfig.LocalConfig.BasePath)
		return
	}
	if appConfig.LocalConfig.LruSize != 33554432 {
		t.Errorf("local configlru size invalid: %s", appConfig.LocalConfig.LruSize)
		return
	}
}

func TestS3File(t *testing.T) {
	fm := initTempFileManager()
	defer fm.close()

	path, err := fm.get("storage-s3")
	if err != nil {
		t.Error(err.Error())
		return
	}
	appConfig, err := NewAppConfig(path)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if appConfig.StorageBackend != "s3" {
		t.Errorf("config backend set to %s", appConfig.StorageBackend)
		return
	}
	if appConfig.S3Config == nil {
		t.Error("app config s3 config should be set.")
		return
	}
	if appConfig.LocalConfig != nil {
		t.Error("app config local config should not be set.")
		return
	}
	if len(appConfig.S3Config.Buckets) == 0 {
		t.Errorf("app config s3 config should have buckets set.")
		return
	}
}
