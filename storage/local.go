package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorageManager struct {
	basePath string
}

type LocalCachedFile struct {
	contentHash string
	path        string
	urls        []string
	aliases     []string
}

func NewLocalStorageManager(basePath string) StorageManager {
	return &LocalStorageManager{basePath}
}

func (sm *LocalStorageManager) Load(callback chan CachedFile) {
	walkFn := func(path string, _ os.FileInfo, err error) error {
		stat, err := os.Stat(path)
		if err != nil {
			return err
		}

		if stat.IsDir() && path != sm.basePath {
			return filepath.SkipDir
		}

		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".metadata") {
			cachedFile, err := sm.unpackLocalCachedFile(path)
			if err != nil {
				callback <- cachedFile
			} else {
				fmt.Println(err.Error())
			}
		}
		return nil
	}
	err := filepath.Walk(sm.basePath, walkFn)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (sm *LocalStorageManager) unpackLocalCachedFile(path string) (*LocalCachedFile, error) {
	return nil, StorageError{"LocalStorageManager.unpackLocalCachedFile(path) not implemented yet."}
}

func (sm *LocalStorageManager) Store(payload []byte, sourceUrl string, contentHash string, aliases []string, callback chan CachedFile) {
	path := filepath.Join(sm.basePath, contentHash)

	cachedFile := NewLocalCachedFile(contentHash, path, []string{sourceUrl}, aliases)
	err1 := sm.persistCachedFileToDisk(cachedFile, payload)
	if err1 != nil {
		fmt.Println(err1.Error())
		return
	}
	err2 := sm.persistMetaDaToDisk(cachedFile)
	if err1 != nil {
		fmt.Println(err2.Error())
		return
	}

	callback <- cachedFile
}

func (sm *LocalStorageManager) persistCachedFileToDisk(cachedFile *LocalCachedFile, payload []byte) error {
	location := cachedFile.Location()
	err := ioutil.WriteFile(location, payload, 00777)
	return err
}

func (sm *LocalStorageManager) persistMetaDaToDisk(cachedFile *LocalCachedFile) error {
	location := cachedFile.Location()
	data, err := cachedFile.Serialize()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(location+".metadata", data, 00777)
	return err
}

func NewLocalCachedFile(contentHash, path string, urls []string, aliases []string) *LocalCachedFile {
	return &LocalCachedFile{contentHash, path, urls, aliases}
}

func (cf *LocalCachedFile) LocationType() string {
	return CachedFile_Local
}

func (cf *LocalCachedFile) Location() string {
	return cf.path
}

func (cf *LocalCachedFile) Size() int {
	stat, err := os.Stat(cf.path)
	if err != nil {
		return 0
	}
	return int(stat.Size())
}

func (cf *LocalCachedFile) Delete() error {
	err := os.Remove(cf.path)
	if err != nil {
		return err
	}
	err = os.Remove(cf.path + ".metadata")
	if err != nil {
		return err
	}
	return nil
}

func (cf *LocalCachedFile) ContentHash() string {
	return cf.contentHash
}

func (cf *LocalCachedFile) Urls() []string {
	return cf.urls
}

func (cf *LocalCachedFile) Aliases() []string {
	return cf.aliases
}

func (cf *LocalCachedFile) Serialize() ([]byte, error) {
	data, err := json.Marshal(cf)
	if err != nil {
		return nil, err
	}
	return data, nil
}
