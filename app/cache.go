package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	notify         = make(chan Event, 10)
	query          = make(chan QueryCachedFiles, 10)
	cachedFiles    = make(map[string]*CachedFile)
	aliases        = make(map[string]string)
	cacheDirectory = ""
)

type Event struct {
	Name       string
	Attributes map[string]string
}

type QueryCachedFiles struct {
	Query    string
	Response chan *CachedFile
}

type CachedFile struct {
	Url     string
	Aliases []string
	Path    string
}

func (event Event) String() string {
	pairs := make([]string, 0, 0)
	for key, value := range event.Attributes {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
	}
	return fmt.Sprintf("%s %s", event.Name, strings.Join(pairs, " "))
}

func PublishEvent(name string, attributes map[string]string) {
	notify <- Event{name, attributes}
}

func Query(token string) *CachedFile {
	search := QueryCachedFiles{token, make(chan *CachedFile)}
	defer close(search.Response)
	query <- search

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()

	select {
	case result := <-search.Response:
		return result
	case <-timeout:
		return nil
	}
}

func init() {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	cacheDirectory = filepath.Join(pwd, ".cache")
	os.MkdirAll(cacheDirectory, 00777)

	go fileCache()
	PublishEvent("init", make(map[string]string))
}

func fileCache() {

	for {
		select {
		case event := <-notify:
			{
				switch event.Name {
				case "init":
					{
						fmt.Println("Running init")
					}
				case "download":
					{
						url, hasUrl := event.Attributes["url"]
						urlAliases, hasUrlAliases := event.Attributes["aliases"]
						if hasUrl && hasUrlAliases {
							cachedFile := findCachedFile(url)
							if cachedFile == nil {
								download(url, strings.Split(urlAliases, ","))
							}
						}
					}
				}
			}
		case search := <-query:
			{
				search.Response <- findCachedFile(search.Query)
			}
		}
	}
}

func findCachedFile(search string) *CachedFile {
	contentHash, hasContentHash := aliases[search]
	if hasContentHash {
		cachedFile, hasCachedFile := cachedFiles[contentHash]
		if hasCachedFile {
			return cachedFile
		}
	}
	return nil
}

func download(url string, urlAliases []string) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	urlHash := hash([]byte(url))
	contentHash := hash(body)
	path := filepath.Join(cacheDirectory, contentHash)

	fmt.Println("Storing url", url, "as id", urlHash, "with content addressed as", contentHash)

	ioutil.WriteFile(path, body, 00777)
	cachedFiles[contentHash] = &CachedFile{url, urlAliases, path}

	aliases[url] = contentHash
	aliases[urlHash] = contentHash
	for _, alias := range urlAliases {
		aliases[alias] = contentHash
	}

	var buffer bytes.Buffer
	buffer.WriteString(url)
	buffer.WriteString("\n" + contentHash)
	if len(urlAliases) > 0 {
		buffer.WriteString("\n")
		buffer.WriteString(strings.Join(urlAliases, "\n"))
	}
	ioutil.WriteFile(filepath.Join(cacheDirectory, contentHash+".aliases"), buffer.Bytes(), 00777)
}

func hash(bytes []byte) string {
	hasher := sha1.New()
	hasher.Write(bytes)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
