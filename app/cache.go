package app

import (
	"github.com/ngerakines/tram/config"
	"github.com/ngerakines/tram/util"
	"time"
)

type warmAndQueryCachedFiles struct {
	Url      string
	Aliases  []string
	Response chan CachedFile
}

type FileCache interface {
	WarmAndQuery(url string, aliases []string) CachedFile
}

type diskFileCache struct {
	appConfig *config.AppConfig

	warmAndQuery chan warmAndQueryCachedFiles
	downloads    chan CachedFile
	evictions    chan *Item

	downloader        util.RemoteFileFetcher
	downloadListeners *DownloadListeners
	downloadPool      *util.DownloadPool

	lru *LRUCache

	index          Index
	storageManager StorageManager
}

func newDiskFileCache(appConfig *config.AppConfig, index Index, storageManager StorageManager, downloader util.RemoteFileFetcher) FileCache {
	fileCache := new(diskFileCache)
	fileCache.appConfig = appConfig
	fileCache.index = index
	fileCache.storageManager = storageManager
	fileCache.downloader = downloader

	fileCache.warmAndQuery = make(chan warmAndQueryCachedFiles, 1024)
	fileCache.downloads = make(chan CachedFile, 25)
	fileCache.downloadListeners = NewDownloadListeners()
	fileCache.evictions = make(chan *Item, 25)
	fileCache.lru = NewLRUCache(appConfig.LruSize)

	fileCache.lru.AddListener(fileCache.evictions)
	go fileCache.run()

	return fileCache
}

func (fileCache *diskFileCache) Close() {
	close(fileCache.warmAndQuery)
}

func (fileCache *diskFileCache) WarmAndQuery(url string, aliases []string) CachedFile {
	command := warmAndQueryCachedFiles{url, aliases, make(chan CachedFile)}
	defer close(command.Response)
	fileCache.warmAndQuery <- command

	select {
	case result := <-command.Response:
		return result
	case <-time.After(30 * time.Second):
		return nil
	}
}

func (fileCache *diskFileCache) run() {
	for {
		select {
		case command, ok := <-fileCache.warmAndQuery:
			{
				if !ok {
					return
				}
				fileCache.downloadAndNotify(command.Url, command.Aliases, command.Response)
			}
		case cachedFile, ok := <-fileCache.downloads:
			{
				if !ok {
					return
				}
				fileCache.handleDownload(cachedFile)
			}
		case evicted, ok := <-fileCache.evictions:
			{
				if !ok {
					return
				}
				fileCache.handleEviction(evicted)
			}
		}
	}
}

func (fileCache *diskFileCache) findCachedFile(terms []string) CachedFile {
	contentHash, err := fileCache.index.Find(terms)
	if err == nil {
		cachedFile, hasCachedFile := fileCache.lru.Get(contentHash)
		if hasCachedFile {
			return cachedFile.(CachedFile)
		}
	}
	// NKG: Returning nil like this kind of bugs me.
	return nil
}

func (fileCache *diskFileCache) downloadAndNotify(url string, urlAliases []string, channel chan CachedFile) {
	existingCachedFile := fileCache.findCachedFile(append(urlAliases, url))
	if existingCachedFile != nil {
		fileCache.index.Merge(existingCachedFile, urlAliases, []string{url})
		channel <- existingCachedFile
		return
	}
	fileCache.downloadListeners.Add(url, urlAliases, channel)
	go Download(fileCache.downloader, fileCache.storageManager, url, urlAliases, fileCache.downloads)
}

func (fileCache *diskFileCache) handleDownload(cachedFile CachedFile) {
	fileCache.lru.Set(cachedFile.ContentHash(), cachedFile)
	fileCache.index.Update(cachedFile)
	fileCache.downloadListeners.Notify(cachedFile)
}

func (fileCache *diskFileCache) handleEviction(evicted *Item) {
	cachedFile := evicted.Value.(CachedFile)
	fileCache.storageManager.Delete(cachedFile)
	fileCache.index.Clear(cachedFile.ContentHash())
}
