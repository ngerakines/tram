package app

import (
	"fmt"
	"github.com/codegangsta/martini"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileCache interface {
	WarmAndQuery(url string, fileAliases []string) *CachedFile
	Query(tokens []string) *CachedFile
	Warm(url string, fileAliases []string)
}

type DiskFileCacheConfig struct {
	downloader RemoteFileFetcher
	basePath   string
	lruSize    uint64
}

var DefaultDiskFileCacheConfig = DiskFileCacheConfig{
	downloader: DedupeWrapDownloader(defaultRemoteFileFetcher),
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
	downloads         chan *CachedFile
	evictions         chan *Item
	downloadListeners *DownloadListeners
	downloadPool      *DownloadPool
	lru               *LRUCache

	// cachedFiles       map[string]*CachedFile
	cachedFileAliases map[string]string
}

func NewFileCacheWithConfig(config DiskFileCacheConfig) martini.Handler {
	fileCache := NewDiskFileCache(config)
	return NewFileCacheMiddleware(fileCache)
}

func NewFileCacheMiddleware(fileCache FileCache) martini.Handler {
	return func(_ http.ResponseWriter, _ *http.Request, c martini.Context, _ *log.Logger) {
		c.MapTo(fileCache, (*FileCache)(nil))
		c.Next()
	}
}

func NewDiskFileCache(config DiskFileCacheConfig) *DiskFileCache {
	diskFileCache := new(DiskFileCache)
	diskFileCache.config = config
	diskFileCache.query = make(chan QueryCachedFiles, 10)
	diskFileCache.warm = make(chan WarmCachedFiles, 10)
	diskFileCache.warmAndQuery = make(chan WarmAndQueryCachedFiles, 10)
	diskFileCache.downloads = make(chan *CachedFile, 25)
	diskFileCache.downloadListeners = NewDownloadListeners()
	diskFileCache.evictions = make(chan *Item, 25)
	diskFileCache.lru = NewLRUCache(config.lruSize)
	// diskFileCache.cachedFiles = make(map[string]*CachedFile)
	diskFileCache.cachedFileAliases = make(map[string]string)

	diskFileCache.lru.AddListener(diskFileCache.evictions)
	go diskFileCache.fileCache()

	return diskFileCache
}

func (dfc *DiskFileCache) Close() {
	close(dfc.query)
	close(dfc.warm)
	close(dfc.warmAndQuery)
}

func (dfc *DiskFileCache) WarmAndQuery(url string, fileAliases []string) *CachedFile {
	command := WarmAndQueryCachedFiles{url, fileAliases, make(chan *CachedFile)}
	defer close(command.Response)
	dfc.warmAndQuery <- command

	select {
	case result := <-command.Response:
		return result
	case <-time.After(3 * 1e9):
		return nil
	}
}

func (dfc *DiskFileCache) Query(tokens []string) *CachedFile {
	search := QueryCachedFiles{tokens, make(chan *CachedFile)}
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
	dfc.initCachedFiles()
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

func (dfc *DiskFileCache) initCachedFiles() {
	walkFn := func(path string, _ os.FileInfo, err error) error {
		stat, err := os.Stat(path)
		if err != nil {
			return err
		}

		if stat.IsDir() && path != dfc.config.basePath {
			return filepath.SkipDir
		}

		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".metadata") {
			content, err := ioutil.ReadFile(path)
			if err == nil {
				dir, fileName := filepath.Split(path)
				fileNameParts := strings.Split(fileName, ".")
				contentHash := fileNameParts[0]
				fileAliases := strings.Split(string(content), "\n")
				dfc.lru.addNew(contentHash, &CachedFile{contentHash, fileAliases[0], fileAliases, filepath.Join(dir, contentHash)})
				// dfc.cachedFiles[contentHash] = &CachedFile{contentHash, fileAliases[0], fileAliases, filepath.Join(dir, contentHash)}
				for _, alias := range fileAliases {
					dfc.cachedFileAliases[alias] = contentHash
				}
			}
		}
		return nil
	}
	err := filepath.Walk(dfc.config.basePath, walkFn)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (dfc *DiskFileCache) findCachedFile(tokens []string) *CachedFile {
	for _, token := range tokens {
		contentHash, hasContentHash := dfc.cachedFileAliases[token]
		if hasContentHash {
			// cachedFile, hasCachedFile := dfc.cachedFiles[contentHash]
			cachedFile, hasCachedFile := dfc.lru.Get(contentHash)
			if hasCachedFile {
				return cachedFile.(*CachedFile)
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
	go asyncDownload(dfc.config.downloader, dfc.config.basePath, url, urlAliases, dfc.downloads)
}

func (dfc *DiskFileCache) downloadAndNotify(url string, urlAliases []string, channel chan *CachedFile) {
	existingCachedFile := dfc.findCachedFile(append(urlAliases, url))
	if existingCachedFile != nil {
		channel <- existingCachedFile
		return
	}
	dfc.downloadListeners.Add(url, urlAliases, channel)
	go asyncDownload(dfc.config.downloader, dfc.config.basePath, url, urlAliases, dfc.downloads)
}

func (dfc *DiskFileCache) handleDownload(cachedFile *CachedFile) {
	dfc.lru.Set(cachedFile.ContentHash, cachedFile)
	// dfc.cachedFiles[cachedFile.ContentHash] = cachedFile
	for _, alias := range cachedFile.Aliases {
		dfc.cachedFileAliases[alias] = cachedFile.ContentHash
	}
	dfc.downloadListeners.Notify(cachedFile)
}

func (dfc *DiskFileCache) handleEviction(evicted *Item) {
	for _, alias := range evicted.Value.(*CachedFile).Aliases {
		delete(dfc.cachedFileAliases, alias)
	}
	evicted.Value.(*CachedFile).Delete()
}

func asyncDownload(downloader RemoteFileFetcher, basePath, url string, urlAliases []string, callback chan *CachedFile) {
	body, err := downloader(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	urlHash := hash([]byte(url))
	contentHash := hash(body)
	path := filepath.Join(basePath, contentHash)

	allAliases := make(map[string]bool)
	allAliases[url] = true
	allAliases[urlHash] = true
	allAliases[contentHash] = true
	for _, alias := range urlAliases {
		allAliases[alias] = true
	}

	cachedFile := &CachedFile{contentHash, url, mapKeys(allAliases), path}
	cachedFile.StoreAsset(body)
	cachedFile.StoreMetadata()

	callback <- cachedFile
}
