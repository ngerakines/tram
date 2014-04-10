
package app;

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"crypto/sha1"
	"os"
	"path/filepath"
	"strings"
)

var (
	aliases = make(map[string]string)
	cacheDirectory = ""
)

func init() {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	cacheDirectory = filepath.Join(pwd, "cache")
	os.MkdirAll(cacheDirectory, 00777)
}

func HandleIndex(res http.ResponseWriter, req *http.Request) (int, string) {
	fmt.Println("method", req.Method)
	values := getValues(req, []string{"url"})
	url := values["url"]
	if req.Method == "HEAD" {
		fmt.Println("Checking to see if url", url, "was downloaded.")
		hash := hashForAlias(url)
		if hash == "" {
			return 404, "not found"
		}
		return 200, hash
	} else if req.Method == "GET" {
		fmt.Println("Attempting to return content of downloaded url", url, ".")
		hash := hashForAlias(url)
		if hash == "" {
			return 404, "not found"
		}
		http.ServeFile(res, req, filepath.Join(cacheDirectory, hash))
		return 200, hash
	} else if req.Method == "POST" {
		fmt.Println("Attempting have url", url, "fetched and cached.")
		download(url, []string{})
	}
	return 200, "OK"
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

	fmt.Println("Storing url", url, "as id", urlHash, "with content addressed as", contentHash)

	ioutil.WriteFile(filepath.Join(cacheDirectory, contentHash), body, 00777)
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
		buffer.WriteString(strings.Join(urlAliases,"\n"))
	}
	ioutil.WriteFile(filepath.Join(cacheDirectory, contentHash + ".aliases"), buffer.Bytes(), 00777)
}

func hashForAlias(alias string) string {
	hash, ok := aliases[alias]; if ok {
		return hash
	}
	return ""
}

func getValues(req *http.Request, keys []string) map[string]string {
	values := make(map[string]string)
	fmt.Println(req.Method)
	if (req.Method == "GET" || req.Method == "HEAD" ) {
		queryValues := req.URL.Query()
		for _, key := range keys {
			values[key] = queryValues.Get(key)
		}
	} else {
		req.ParseForm()
		formValues := req.Form
		fmt.Println(formValues)
		for _, key := range keys {
			values[key] = formValues.Get(key)
		}
	}
	return values
}

func hash(bytes []byte) string {
	hasher := sha1.New()
	hasher.Write(bytes)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
