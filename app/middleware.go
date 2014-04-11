package app

import (
	"fmt"
	"github.com/codegangsta/martini"
	"log"
	"net/http"
	"time"
	"os"
	"path/filepath"
	"strings"
	"io/ioutil"
)

type FileCache interface {
	WarmAndQuery(url string, fileAliases []string) *CachedFile
	Query(tokens []string) *CachedFile
	Warm(url string, fileAliases []string)
}

func NewFileCacheWithPath(basePath string) martini.Handler {
	fileCache := NewDiskFileCache(basePath)
	return NewFileCacheMiddleware(fileCache)
}

func NewFileCacheMiddleware(fileCache FileCache) martini.Handler {
	return func(_ http.ResponseWriter, _ *http.Request, c martini.Context, _ *log.Logger) {
		c.MapTo(fileCache, (*FileCache)(nil))
		c.Next()
	}
}

type DiskFileCache struct {
	query             chan QueryCachedFiles
	warm              chan WarmCachedFiles
	warmAndQuery      chan WarmAndQueryCachedFiles
}

func NewDiskFileCache(basePath string) *DiskFileCache {
	diskFileCache := new(DiskFileCache)
	diskFileCache.query = make(chan QueryCachedFiles, 10)
	diskFileCache.warm = make(chan WarmCachedFiles, 10)
	diskFileCache.warmAndQuery = make(chan WarmAndQueryCachedFiles, 10)

	go fileCache(diskFileCache.query, diskFileCache.warm, diskFileCache.warmAndQuery, basePath)

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
	case <- time.After(3 * 1e9):
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
	case <- time.After(3 * 1e9):
		return nil
	}
}

func (dfc *DiskFileCache) Warm(url string, fileAliases []string) {
	command := WarmCachedFiles{url, fileAliases}
	dfc.warm <- command
}

func fileCache(query chan QueryCachedFiles, warm chan WarmCachedFiles, warmAndQuery chan WarmAndQueryCachedFiles, cacheDirectory string) {
	cachedFiles := make(map[string]*CachedFile)
	cachedFileAliases := make(map[string]string)
	initCachedFiles(cacheDirectory, cachedFiles, cachedFileAliases)
	for {
		select {
		case command, ok := <-warm:
		{
			if !ok {
				return
			}
			download(command.Url, command.Aliases, cacheDirectory, cachedFiles, cachedFileAliases)
		}
		case command, ok := <-warmAndQuery:
		{
			if !ok {
				return
			}
			command.Response <- download(command.Url, command.Aliases, cacheDirectory, cachedFiles, cachedFileAliases)
		}
		case command, ok := <-query:
		{
			if !ok {
				return
			}
			command.Response <- findCachedFile(command.Query, cacheDirectory, cachedFiles, cachedFileAliases)
		}
		}
	}
}

func initCachedFiles(cacheDirectory string, cachedFiles map[string]*CachedFile, aliases map[string]string) {
	walkFn := func(path string, _ os.FileInfo, err error) error {
		stat, err := os.Stat(path)
		if err != nil {
			return err
		}

		if stat.IsDir() && path != cacheDirectory {
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
				cachedFiles[contentHash] = &CachedFile{fileAliases[0], fileAliases, filepath.Join(dir, contentHash)}
				for _, alias := range fileAliases {
					aliases[alias] = contentHash
				}
			}
		}
		return nil
	}
	err := filepath.Walk(cacheDirectory, walkFn)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func findCachedFile(tokens []string, _ string, cachedFiles map[string]*CachedFile, aliases map[string]string) *CachedFile {
	for _, token := range tokens {
		contentHash, hasContentHash := aliases[token]
		if hasContentHash {
			cachedFile, hasCachedFile := cachedFiles[contentHash]
			if hasCachedFile {
				return cachedFile
			}
		}
	}
	return nil
}

func download(url string, urlAliases []string, cacheDirectory string, cachedFiles map[string]*CachedFile, aliases map[string]string) *CachedFile {
	existingCachedFile := findCachedFile(append(urlAliases, url), cacheDirectory, cachedFiles, aliases)
	if existingCachedFile != nil {
		return existingCachedFile
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	urlHash := hash([]byte(url))
	contentHash := hash(body)
	path := filepath.Join(cacheDirectory, contentHash)

	allAliases := make(map[string]bool)
	allAliases[url] = true
	allAliases[urlHash] = true
	allAliases[contentHash] = true
	for _, alias := range urlAliases {
		allAliases[alias] = true
	}

	cachedFile := &CachedFile{url, mapKeys(allAliases), path}
	cachedFile.StoreAsset(body)
	cachedFile.StoreMetadata()

	cachedFiles[contentHash] = cachedFile
	for _, alias := range cachedFile.Aliases {
		aliases[alias] = contentHash
	}

	return cachedFile
}
