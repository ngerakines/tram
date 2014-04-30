package app

import (
	"github.com/ngerakines/tram/storage"
	"github.com/ngerakines/tram/util"
	"os"
	"path/filepath"
	"time"
)

type QueryCachedFiles struct {
	Query    []string
	Response chan storage.CachedFile
}

type WarmCachedFiles struct {
	Url     string
	Aliases []string
}

type WarmAndQueryCachedFiles struct {
	Url      string
	Aliases  []string
	Response chan storage.CachedFile
}

type FileCache interface {
	WarmAndQuery(url string, fileAliases []string) storage.CachedFile
	Query(tokens []string) storage.CachedFile
	Warm(url string, fileAliases []string)
}

type DiskFileCacheConfig struct {
	downloader util.RemoteFileFetcher
	basePath   string
	lruSize    uint64
}

var DefaultDiskFileCacheConfig = DiskFileCacheConfig{
	downloader: util.DedupeWrapDownloader(util.DefaultRemoteFileFetcher),
	lruSize:    33554432, // 32 megabytes
	// NKG: I know.
	basePath: func() string {
		pwd, err := os.Getwd()
		if err != nil {
			panic(err.Error())
		}
		cacheDirectory := filepath.Join(pwd, ".cache")
		os.MkdirAll(cacheDirectory, 00777)
		return cacheDirectory
	}(),
}

type DiskFileCache struct {
	config            DiskFileCacheConfig
	query             chan QueryCachedFiles
	warm              chan WarmCachedFiles
	warmAndQuery      chan WarmAndQueryCachedFiles
	downloads         chan storage.CachedFile
	evictions         chan *Item
	downloadListeners *DownloadListeners
	downloadPool      *util.DownloadPool
	lru               *LRUCache
	storageManager    storage.StorageManager

	cachedFileAliases map[string]string
}

func NewDiskFileCache(config DiskFileCacheConfig) *DiskFileCache {
	diskFileCache := new(DiskFileCache)
	diskFileCache.config = config
	diskFileCache.query = make(chan QueryCachedFiles, 10)
	diskFileCache.warm = make(chan WarmCachedFiles, 10)
	diskFileCache.warmAndQuery = make(chan WarmAndQueryCachedFiles, 10)
	diskFileCache.downloads = make(chan storage.CachedFile, 25)
	diskFileCache.downloadListeners = NewDownloadListeners()
	diskFileCache.evictions = make(chan *Item, 25)
	diskFileCache.lru = NewLRUCache(config.lruSize)
	diskFileCache.cachedFileAliases = make(map[string]string)
	diskFileCache.storageManager = storage.NewLocalStorageManager(config.basePath)

	diskFileCache.lru.AddListener(diskFileCache.evictions)
	go diskFileCache.fileCache()

	return diskFileCache
}

func (dfc *DiskFileCache) Close() {
	close(dfc.query)
	close(dfc.warm)
	close(dfc.warmAndQuery)
}

func (dfc *DiskFileCache) WarmAndQuery(url string, fileAliases []string) storage.CachedFile {
	command := WarmAndQueryCachedFiles{url, fileAliases, make(chan storage.CachedFile)}
	defer close(command.Response)
	dfc.warmAndQuery <- command

	select {
	case result := <-command.Response:
		return result
	case <-time.After(3 * 1e9):
		return nil
	}
}

func (dfc *DiskFileCache) Query(tokens []string) storage.CachedFile {
	search := QueryCachedFiles{tokens, make(chan storage.CachedFile)}
	defer close(search.Response)
	dfc.query <- search

	select {
	case result := <-search.Response:
		return result
	case <-time.After(3 * 1e9):
		return nil
	}
}

func (dfc *DiskFileCache) Warm(url string, fileAliases []string) {
	command := WarmCachedFiles{url, fileAliases}
	dfc.warm <- command
}

func (dfc *DiskFileCache) fileCache() {
	dfc.storageManager.Load(dfc.downloads)
	for {
		select {
		case command, ok := <-dfc.warm:
			{
				if !ok {
					return
				}
				dfc.download(command.Url, command.Aliases)
			}
		case command, ok := <-dfc.warmAndQuery:
			{
				if !ok {
					return
				}
				dfc.downloadAndNotify(command.Url, command.Aliases, command.Response)
			}
		case command, ok := <-dfc.query:
			{
				if !ok {
					return
				}
				command.Response <- dfc.findCachedFile(command.Query)
			}
		case cachedFile, ok := <-dfc.downloads:
			{
				if !ok {
					return
				}
				dfc.handleDownload(cachedFile)
			}
		case evicted, ok := <-dfc.evictions:
			{
				if !ok {
					return
				}
				dfc.handleEviction(evicted)
			}
		}
	}
}

func (dfc *DiskFileCache) findCachedFile(tokens []string) storage.CachedFile {
	for _, token := range tokens {
		contentHash, hasContentHash := dfc.cachedFileAliases[token]
		if hasContentHash {
			cachedFile, hasCachedFile := dfc.lru.Get(contentHash)
			if hasCachedFile {
				return cachedFile.(storage.CachedFile)
			}
		}
	}
	return nil
}

func (dfc *DiskFileCache) download(url string, urlAliases []string) {
	existingCachedFile := dfc.findCachedFile(append(urlAliases, url))
	if existingCachedFile != nil {
		return
	}
	go storage.Download(dfc.config.downloader, dfc.storageManager, url, urlAliases, dfc.downloads)
}

func (dfc *DiskFileCache) downloadAndNotify(url string, urlAliases []string, channel chan storage.CachedFile) {
	existingCachedFile := dfc.findCachedFile(append(urlAliases, url))
	if existingCachedFile != nil {
		channel <- existingCachedFile
		return
	}
	dfc.downloadListeners.Add(url, urlAliases, channel)
	go storage.Download(dfc.config.downloader, dfc.storageManager, url, urlAliases, dfc.downloads)
}

func (dfc *DiskFileCache) handleDownload(cachedFile storage.CachedFile) {
	dfc.lru.Set(cachedFile.ContentHash(), cachedFile)
	for _, alias := range cachedFile.Aliases() {
		dfc.cachedFileAliases[alias] = cachedFile.ContentHash()
	}
	dfc.downloadListeners.Notify(cachedFile)
}

func (dfc *DiskFileCache) handleEviction(evicted *Item) {
	for _, alias := range evicted.Value.(storage.CachedFile).Aliases() {
		delete(dfc.cachedFileAliases, alias)
	}
	evicted.Value.(storage.CachedFile).Delete()
}
