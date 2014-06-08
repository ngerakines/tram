package app

import (
	"github.com/ngerakines/tram/util"
	"log"
	"net/http"
)

type CachedFile interface {
	ContentHash() string
	Urls() []string
	Aliases() []string
	Size() int
	Attributes() map[string]string
}

type StorageManager interface {
	Store(contentHash string, payload []byte, urls, aliases []string, callback chan CachedFile)
	Delete(cachedFile CachedFile) error
	Serve(cachedFile CachedFile, res http.ResponseWriter, req *http.Request) error
}

type simpleCachedFile struct {
	InternalContentHash string            `json:"ContentHash"`
	InternalUrls        []string          `json:"Urls"`
	InternalAliases     []string          `json:"Aliases"`
	InternalSize        int               `json:"Size"`
	InternalAttributes  map[string]string `json:"Attributes"`
}

func Download(downloader util.RemoteFileFetcher, storageManager StorageManager, url string, aliases []string, callback chan CachedFile) {
	body, err := downloader(url)
	if err != nil {
		log.Println(err.Error())
		return
	}

	contentHash := util.Hash(body)

	storageManager.Store(contentHash, body, []string{url}, aliases, callback)
}

func (cachedFile *simpleCachedFile) ContentHash() string {
	return cachedFile.InternalContentHash
}

func (cachedFile *simpleCachedFile) Urls() []string {
	return cachedFile.InternalUrls
}

func (cachedFile *simpleCachedFile) Aliases() []string {
	return cachedFile.InternalAliases
}

func (cachedFile *simpleCachedFile) Size() int {
	return cachedFile.InternalSize
}

func (cachedFile *simpleCachedFile) Attributes() map[string]string {
	return cachedFile.InternalAttributes
}
