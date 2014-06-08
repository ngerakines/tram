package storage

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type LocalStorageManager struct {
	basePath string
	index    Index
}

type LocalCachedFile struct {
	contentHash string
	path        string
	urls        []string
	aliases     []string
}

func NewLocalStorageManager(basePath string, index Index) StorageManager {
	return &LocalStorageManager{basePath, index}
}

func (sm *LocalStorageManager) Store(payload []byte, sourceUrl string, contentHash string, aliases []string, callback chan CachedFile) {
	path := filepath.Join(sm.basePath, contentHash)

	cachedFile := NewLocalCachedFile(contentHash, path, []string{sourceUrl}, aliases)
	err := sm.persistCachedFileToDisk(cachedFile, payload)
	if err != nil {
		log.Println(err)
		return
	}

	err = sm.index.Update(contentHash, aliases, []string{sourceUrl}, len(payload))
	if err != nil {
		log.Println(err)
		return
	}

	callback <- cachedFile
}

func (sm *LocalStorageManager) persistCachedFileToDisk(cachedFile *LocalCachedFile, payload []byte) error {
	location := cachedFile.Location()
	err := ioutil.WriteFile(location, payload, 00777)
	return err
}

func (sm *LocalStorageManager) Delete(cachedFile CachedFile) error {
	err := os.Remove(cachedFile.Location())
	if err != nil {
		return err
	}
	err = sm.index.Clear(cachedFile.ContentHash())
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
