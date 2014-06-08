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

type indexData struct {
	ContentHash string
	Aliases     []string
	Urls        []string
	size        int
}

type Index interface {
	Find(terms []string) (string, error)
	Update(contentHash string, aliases, urls []string, size int) error
	Merge(contentHash string, aliases, urls []string, size int) error
	Clear(id string) error
}

type localIndex struct {
	path string

	aliases map[string]string
}

func NewLocalIndex(path string) Index {
	index := new(localIndex)
	index.path = path
	index.aliases = make(map[string]string)
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
			for _, alias := range data.Aliases {
				index.aliases[alias] = data.ContentHash
			}
			for _, url := range data.Urls {
				index.aliases[url] = data.ContentHash
			}
		}
		return nil
	}
	err := filepath.Walk(index.path, walkFn)
	if err != nil {
		log.Println(err)
	}
}

func (index *localIndex) Update(contentHash string, aliases, urls []string, size int) error {
	err := index.write(contentHash, aliases, urls)
	if err != nil {
		return err
	}
	for _, alias := range aliases {
		index.aliases[alias] = contentHash
	}
	for _, url := range urls {
		index.aliases[url] = contentHash
	}
	return nil
}

func (index *localIndex) Merge(contentHash string, aliases, urls []string, size int) error {

	existingData, err := index.load(contentHash)
	if err != nil {
		return index.Update(contentHash, aliases, urls, size)
	}
	if existingData.size != size {
		log.Println("Merge called and there is a size mismatch.", contentHash)
	}

	allAliases := make([]string, 0, 0)
	for _, alias := range existingData.Aliases {
		allAliases = append(allAliases, alias)
	}
	for _, alias := range aliases {
		allAliases = append(allAliases, alias)
	}

	allUrls := make([]string, 0, 0)
	for _, url := range existingData.Urls {
		allUrls = append(allUrls, url)
	}
	for _, url := range urls {
		allUrls = append(allUrls, url)
	}

	err = index.write(contentHash, allAliases, allUrls)
	if err != nil {
		return err
	}
	for _, alias := range aliases {
		index.aliases[alias] = contentHash
	}
	for _, url := range urls {
		index.aliases[url] = contentHash
	}
	return nil
}

func (index *localIndex) Clear(contentHash string) error {
	data, err := index.load(contentHash)
	if err != nil {
		return err
	}

	for _, alias := range data.Aliases {
		delete(index.aliases, alias)
	}
	for _, url := range data.Urls {
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
			return contentHash, nil
		}
	}
	return "", errors.New("No content hash found for term")
}

func (index *localIndex) write(contentHash string, aliases, urls []string) error {
	location := index.indexPath(contentHash)

	data, err := json.Marshal(&indexData{ContentHash: contentHash, Aliases: aliases, Urls: urls})
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(location, data, 00777)
	return err
}

func (index *localIndex) load(contentHash string) (*indexData, error) {
	location := index.indexPath(contentHash)

	content, err := ioutil.ReadFile(location)
	if err != nil {
		return nil, err
	}

	var data indexData
	err = json.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (index *localIndex) indexPath(id string) string {
	return filepath.Join(index.path, id)
}
