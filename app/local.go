package app

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func newLocalStorageManager(basePath string) StorageManager {
	return &LocalStorageManager{basePath}
}

func (storageManager *LocalStorageManager) Store(contentHash string, payload []byte, urls, aliases []string, callback chan CachedFile) {
	path := filepath.Join(storageManager.basePath, contentHash)

	cachedFile := storageManager.newCachedFile(contentHash, urls, aliases, len(payload), path)
	err := ioutil.WriteFile(path, payload, 00777)
	if err != nil {
		log.Println(err)
		return
	}

	callback <- cachedFile
}

func (storageManager *LocalStorageManager) Delete(cachedFile CachedFile) error {
	path := filepath.Join(storageManager.basePath, cachedFile.ContentHash())
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return err
}

func (storageManager *LocalStorageManager) Serve(cachedFile CachedFile, res http.ResponseWriter, req *http.Request) error {
	path, hasPath := cachedFile.Attributes()["path"]
	if !hasPath {
		log.Println("Could not serve file because path attribute not set", cachedFile)
		return errors.New("Invalid cached file.")
	}
	http.ServeFile(res, req, path)
	return nil
}

func (storageManager *LocalStorageManager) newCachedFile(contentHash string, urls, aliases []string, size int, path string) CachedFile {
	attributes := make(map[string]string)
	attributes["path"] = path
	cachedFile := new(simpleCachedFile)
	cachedFile.InternalContentHash = contentHash
	cachedFile.InternalUrls = urls
	cachedFile.InternalAliases = aliases
	cachedFile.InternalSize = size
	cachedFile.InternalAttributes = attributes
	return cachedFile
}
