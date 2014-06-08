package config

import (
	"github.com/ngerakines/testutils"
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"
)

type tempFileManager struct {
	path  string
	files map[string]string
}

func (fm *tempFileManager) initFile(name, body string) {
	path := filepath.Join(fm.path, name)
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
	return "", appConfigError{"No config file exists with that label."}
}

func initTempFileManager(path string) *tempFileManager {
	fm := new(tempFileManager)
	fm.path = path
	fm.files = make(map[string]string)
	fm.initFile("basic", `{
		"listen": ":7041",
		"lruSize": 120000,
		"index": {"engine": "local", "localBasePath": "./index"},
		"storage": {"engine": "s3", "s3Buckets": ["localhost"], "s3Key": "foo", "s3Secret": "bar", "s3Host": "localhost"}
		}`)
	return fm
}

func TestDefaultConfig(t *testing.T) {
	dm := testutils.NewDirectoryManager()
	defer dm.Close()

	appConfig, err := NewDefaultAppConfig()
	if err != nil {
		t.Error(err.Error())
		return
	}

	if appConfig.Listen() != ":7040" {
		t.Error("Invalid default for appConfig.Listen()", appConfig.Listen())
		return
	}
	if appConfig.Storage().Engine() != "local" {
		t.Error("Invalid default for appConfig.Storage().Engine()", appConfig.Storage().Engine())
		return
	}
}

func TestBasicConfig(t *testing.T) {
	dm := testutils.NewDirectoryManager()
	defer dm.Close()
	fm := initTempFileManager(dm.Path)

	path, err := fm.get("basic")
	if err != nil {
		t.Error(err.Error())
		return
	}
	appConfig, err := LoadAppConfig(path)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if appConfig.Listen() != ":7041" {
		t.Error("appConfig.Listen()", appConfig.Listen())
	}

	if appConfig.Storage().Engine() != "s3" {
		t.Error("appConfig.Storage().Engine()", appConfig.Storage().Engine())
	}
}
