package app

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"strings"
)

type QueryCachedFiles struct {
	Query    []string
	Response chan *CachedFile
}

type WarmCachedFiles struct {
	Url     string
	Aliases []string
}

type WarmAndQueryCachedFiles struct {
	Url      string
	Aliases  []string
	Response chan *CachedFile
}

type CachedFile struct {
	Url     string
	Aliases []string
	Path    string
}

func (cf *CachedFile) StoreAsset(body []byte) {
	err := ioutil.WriteFile(cf.Path, body, 00777)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (cf *CachedFile) StoreMetadata() {
	var buffer bytes.Buffer
	buffer.WriteString(cf.Url)
	if len(cf.Aliases) > 0 {
		buffer.WriteString("\n")
		buffer.WriteString(strings.Join(cf.Aliases, "\n"))
	}
	err := ioutil.WriteFile(cf.Path+".metadata", buffer.Bytes(), 00777)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func mapKeys(source map[string]bool) []string {
	values := make([]string, 0, 0)
	for key, _ := range source {
		values = append(values, key)
	}
	return values
}

func hash(bytes []byte) string {
	hasher := sha1.New()
	hasher.Write(bytes)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
