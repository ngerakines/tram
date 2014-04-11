package app

import (
	"fmt"
	"net/http"
)

func HandleIndex(res http.ResponseWriter, req *http.Request, fileCache FileCache) {
	if req.Method == "HEAD" {
		handleHead(res, req, fileCache)
	} else if req.Method == "GET" {
		handleGet(res, req, fileCache)
	} else if req.Method == "POST" {
		handlePost(res, req, fileCache)
	} else {
		res.WriteHeader(404)
	}
}

func handleHead(res http.ResponseWriter, req *http.Request, fileCache FileCache) {
	values := getValues(req, []string{"query"})
	query, hasQuery := values["query"]
	if hasQuery && len(query) > 0 {
		cachedFile := fileCache.Query(query)
		if cachedFile != nil {
			res.Header().Set("Content-Length", "0")
			res.WriteHeader(200)
			return
		}
	}
	res.Header().Set("Content-Length", "0")
	res.WriteHeader(404)
	return
}

func handleGet(res http.ResponseWriter, req *http.Request, fileCache FileCache) {
	values := getValues(req, []string{"query", "url", "alias"})
	query, hasQuery := values["query"]
	fmt.Println("hasQuery", hasQuery)
	if hasQuery && len(query) > 0 {
		cachedFile := fileCache.Query(query)
		if cachedFile != nil {
			http.ServeFile(res, req, cachedFile.Path)
			return
		}
	}
	url, hasUrl := values["url"]
	alias, hasAlias := values["alias"]
	if hasUrl && len(url) > 0 {
		fmt.Println("Getting url", url)
		cachedFile := fileCache.WarmAndQuery(url[0], aliasOrEmpty(hasAlias, alias))
		if cachedFile != nil {
			http.ServeFile(res, req, cachedFile.Path)
			return
		}
	}
	res.Header().Set("Content-Length", "0")
	res.WriteHeader(404)
	return
}

func handlePost(res http.ResponseWriter, req *http.Request, fileCache FileCache) {
	values := getValues(req, []string{"url", "alias"})
	url, hasUrl := values["url"]
	alias, hasAlias := values["alias"]
	if hasUrl && hasAlias && len(url) > 0 {
		fileCache.Warm(url[0], alias)
		res.Header().Set("Content-Length", "0")
		res.WriteHeader(202)
		return
	}
	res.Header().Set("Content-Length", "0")
	res.WriteHeader(500)
	return
}

func aliasOrEmpty(_ bool, alias []string) []string {
	if alias != nil {
		if len(alias) > 0 {
			return alias
		}
	}
	return []string{}
}

func getValues(req *http.Request, keys []string) map[string][]string {
	values := make(map[string][]string)
	if req.Method == "GET" || req.Method == "HEAD" {
		queryValues := req.URL.Query()
		for _, key := range keys {
			value, hasValue := queryValues[key]
			if hasValue {
				values[key] = value
			}
		}
	} else {
		req.ParseForm()
		formValues := req.Form
		for _, key := range keys {
			value, hasValue := formValues[key]
			if hasValue {
				values[key] = value
			}
		}
	}
	return values
}
