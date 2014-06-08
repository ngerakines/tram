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

type DiskFileCache struct {
	appConfig config.AppConfig

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

func NewDiskFileCache(appConfig config.AppConfig, index Index, storageManager StorageManager, downloader util.RemoteFileFetcher) *DiskFileCache {
	diskFileCache := new(DiskFileCache)
	diskFileCache.appConfig = appConfig
	diskFileCache.index = index
	diskFileCache.storageManager = storageManager
	diskFileCache.downloader = downloader

	diskFileCache.warmAndQuery = make(chan warmAndQueryCachedFiles, 1024)
	diskFileCache.downloads = make(chan CachedFile, 25)
	diskFileCache.downloadListeners = NewDownloadListeners()
	diskFileCache.evictions = make(chan *Item, 25)
	diskFileCache.lru = NewLRUCache(appConfig.LruSize())

	diskFileCache.lru.AddListener(diskFileCache.evictions)
	go diskFileCache.fileCache()

	return diskFileCache
}

func (dfc *DiskFileCache) Close() {
	close(dfc.warmAndQuery)
}

func (dfc *DiskFileCache) WarmAndQuery(url string, aliases []string) CachedFile {
	command := warmAndQueryCachedFiles{url, aliases, make(chan CachedFile)}
	defer close(command.Response)
	dfc.warmAndQuery <- command

	select {
	case result := <-command.Response:
		return result
	case <-time.After(30 * time.Second):
		return nil
	}
}

func (dfc *DiskFileCache) fileCache() {
	for {
		select {
		case command, ok := <-dfc.warmAndQuery:
			{
				if !ok {
					return
				}
				dfc.downloadAndNotify(command.Url, command.Aliases, command.Response)
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

func (dfc *DiskFileCache) findCachedFile(terms []string) CachedFile {
	contentHash, err := dfc.index.Find(terms)
	if err == nil {
		cachedFile, hasCachedFile := dfc.lru.Get(contentHash)
		if hasCachedFile {
			return cachedFile.(CachedFile)
		}
	}
	// NKG: Returning nil like this kind of bugs me.
	return nil
}

func (dfc *DiskFileCache) downloadAndNotify(url string, urlAliases []string, channel chan CachedFile) {
	existingCachedFile := dfc.findCachedFile(append(urlAliases, url))
	if existingCachedFile != nil {
		dfc.index.Merge(existingCachedFile.ContentHash(), urlAliases, []string{url}, existingCachedFile.Size())
		channel <- existingCachedFile
		return
	}
	dfc.downloadListeners.Add(url, urlAliases, channel)
	go Download(dfc.downloader, dfc.storageManager, url, urlAliases, dfc.downloads)
}

func (dfc *DiskFileCache) handleDownload(cachedFile CachedFile) {
	dfc.lru.Set(cachedFile.ContentHash(), cachedFile)
	dfc.index.Update(cachedFile.ContentHash(), cachedFile.Aliases(), cachedFile.Urls(), cachedFile.Size())
	dfc.downloadListeners.Notify(cachedFile)
}

func (dfc *DiskFileCache) handleEviction(evicted *Item) {
	cachedFile := evicted.Value.(CachedFile)
	dfc.storageManager.Delete(cachedFile)
	dfc.index.Clear(cachedFile.ContentHash())
}
