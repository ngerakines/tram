package app

import (
	"crypto/sha1"
	"fmt"
	"net/http"
)

func HandleIndex(res http.ResponseWriter, req *http.Request) (int, string) {
	fmt.Println("method", req.Method)
	values := getValues(req, []string{"url"})
	url := values["url"]
	if req.Method == "HEAD" {
		fmt.Println("Checking to see if url", url, "was downloaded.")
		cachedFile := Query(url)
		if (cachedFile != nil) {
			return 200, cachedFile.Path
		}
		return 404, "not found"
	} else if req.Method == "GET" {
		fmt.Println("Attempting to return content of downloaded url", url, ".")
		cachedFile := Query(url)
		if (cachedFile != nil) {
			http.ServeFile(res, req, cachedFile.Path)
			return 200, "OK"
		}
		return 404, "not found"
	} else if req.Method == "POST" {
		fmt.Println("Attempting have url", url, "fetched and cached.")
		PublishEvent("download", map[string]string{"url": url, "aliases": ""})
	}
	return 200, "OK"
}

func hashForAlias(alias string) string {
	hash, ok := aliases[alias]
	if ok {
		return hash
	}
	return ""
}

func getValues(req *http.Request, keys []string) map[string]string {
	values := make(map[string]string)
	fmt.Println(req.Method)
	if req.Method == "GET" || req.Method == "HEAD" {
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
