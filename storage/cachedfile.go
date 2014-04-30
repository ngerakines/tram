package storage

import (
	"crypto/sha1"
	"fmt"
	"github.com/ngerakines/tram/util"
)

var (
	CachedFile_Local  = "LocalCachedFile"
	CachedFile_Remote = "RemoteCachedFile"
)

type CachedFile interface {
	ContentHash() string
	LocationType() string
	Location() string
	Urls() []string
	Aliases() []string
	Size() int
	Delete() error
	Serialize() ([]byte, error)
}

type LocalCachedFile struct {
	contentHash string
	path        string
	urls        []string
	aliases     []string
}

type S3CachedFile struct {
	contentHash string
	remoteUrl   string
	bucket      string
	urls        []string
	aliases     []string
}

type StorageManager interface {
	Store(payload []byte, sourceUrl string, contentHash string, aliases []string, callback chan CachedFile)
	Load(callback chan CachedFile)
}

type S3StorageManager struct {
	buckets []string
}

type LocalStorageManager struct {
	basePath string
}

type StorageError struct {
	message string
}

func (err StorageError) Error() string {
	return err.message
}

func Download(downloader util.RemoteFileFetcher, storageManager StorageManager, url string, aliases []string, callback chan CachedFile) {
	body, err := downloader(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	contentHash := hash(body)
	urlHash := hash([]byte(url))

	allAliases := make(map[string]bool)
	allAliases[url] = true
	allAliases[urlHash] = true
	allAliases[contentHash] = true
	for _, alias := range aliases {
		allAliases[alias] = true
	}

	storageManager.Store(body, url, contentHash, util.MapKeys(allAliases), callback)
}

func hash(bytes []byte) string {
	hasher := sha1.New()
	hasher.Write(bytes)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
