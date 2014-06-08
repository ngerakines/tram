package app

import (
	"github.com/ngerakines/tram/util"
	"log"
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
	Serialize() ([]byte, error)
}

type StorageManager interface {
	Store(payload []byte, sourceUrl string, contentHash string, aliases []string, callback chan CachedFile)
	Delete(cachedFile CachedFile) error
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
		log.Println(err.Error())
		return
	}

	contentHash := util.Hash(body)

	storageManager.Store(body, url, contentHash, aliases, callback)
}
