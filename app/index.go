package app

import (
	"encoding/json"
	"errors"
	// _ "github.com/gwenn/gosqlite"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Index interface {
	Find(terms []string) (string, error)
	Update(cachedFile CachedFile) error
	Merge(cachedFile CachedFile, aliases, urls []string) error
	Clear(id string) error
}

type localIndex struct {
	path string

	aliases map[string]string
}

func newLocalIndex(path string) Index {
	index := new(localIndex)
	index.path = path

	err := os.MkdirAll(path, 0777)
	if err != nil {
		panic(err)
	}

	index.aliases = make(map[string]string)
	index.init()
	return index
}

func (index *localIndex) init() {
	walkFn := func(path string, _ os.FileInfo, err error) error {
		stat, err := os.Stat(path)
		if err != nil {
			return err
		}

		if stat.IsDir() && path != index.path {
			return filepath.SkipDir
		}

		if err != nil {
			return err
		}
		_, file := filepath.Split(path)
		data, err := index.load(file)
		if err == nil {
			for _, alias := range data.Aliases() {
				index.aliases[alias] = data.ContentHash()
			}
			for _, url := range data.Urls() {
				index.aliases[url] = data.ContentHash()
			}
		}
		return nil
	}
	err := filepath.Walk(index.path, walkFn)
	if err != nil {
		log.Println(err)
	}
}

func (index *localIndex) Update(cachedFile CachedFile) error {
	err := index.write(cachedFile)
	if err != nil {
		return err
	}
	for _, alias := range cachedFile.Aliases() {
		index.aliases[alias] = cachedFile.ContentHash()
	}
	for _, url := range cachedFile.Urls() {
		index.aliases[url] = cachedFile.ContentHash()
	}
	return nil
}

func (index *localIndex) Merge(cachedFile CachedFile, aliases, urls []string) error {
	allAliases := make([]string, 0, 0)
	for _, alias := range cachedFile.Aliases() {
		allAliases = append(allAliases, alias)
	}
	for _, alias := range aliases {
		allAliases = append(allAliases, alias)
	}

	allUrls := make([]string, 0, 0)
	for _, url := range cachedFile.Urls() {
		allUrls = append(allUrls, url)
	}
	for _, url := range urls {
		allUrls = append(allUrls, url)
	}

	newCachedFile := new(simpleCachedFile)
	newCachedFile.InternalContentHash = cachedFile.ContentHash()
	newCachedFile.InternalUrls = allUrls
	newCachedFile.InternalAliases = allAliases
	newCachedFile.InternalSize = cachedFile.Size()
	newCachedFile.InternalAttributes = cachedFile.Attributes()

	err := index.write(newCachedFile)
	if err != nil {
		return err
	}
	for _, alias := range allAliases {
		index.aliases[alias] = cachedFile.ContentHash()
	}
	for _, url := range allUrls {
		index.aliases[url] = cachedFile.ContentHash()
	}
	return nil
}

func (index *localIndex) Clear(contentHash string) error {
	cachedFile, err := index.load(contentHash)
	if err != nil {
		return err
	}

	for _, alias := range cachedFile.Aliases() {
		delete(index.aliases, alias)
	}
	for _, url := range cachedFile.Urls() {
		delete(index.aliases, url)
	}

	location := index.indexPath(contentHash)

	err = os.RemoveAll(location)
	return err
}

func (index *localIndex) Find(terms []string) (string, error) {
	for _, term := range terms {
		contentHash, hasContentHash := index.aliases[term]
		if hasContentHash {
			log.Println("Found content", contentHash, "for terms", terms)
			return contentHash, nil
		}
	}
	log.Println("No content has found for terms", terms)
	return "", errors.New("No content hash found for term")
}

func (index *localIndex) write(cachedFile CachedFile) error {
	location := index.indexPath(cachedFile.ContentHash())

	data, err := json.Marshal(cachedFile)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(location, data, 00777)
	return err
}

func (index *localIndex) load(contentHash string) (CachedFile, error) {
	location := index.indexPath(contentHash)

	content, err := ioutil.ReadFile(location)
	if err != nil {
		return nil, err
	}

	var cachedFile simpleCachedFile
	err = json.Unmarshal(content, &cachedFile)
	if err != nil {
		return nil, err
	}
	return &cachedFile, nil
}

func (index *localIndex) indexPath(id string) string {
	return filepath.Join(index.path, id)
}
