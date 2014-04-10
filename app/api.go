package app

import (
	"net/http"
)

func HandleIndex(res http.ResponseWriter, req *http.Request) {
	values := getValues(req, []string{"url"})
	url := values["url"]
	if req.Method == "HEAD" {
		cachedFile := Query(url)
		if cachedFile != nil {
			res.Header().Set("Content-Length", "0")
			res.WriteHeader(200)
			return
		}
		res.Header().Set("Content-Length", "0")
		res.WriteHeader(404)
		return
	} else if req.Method == "GET" {
		cachedFile := Query(url)
		if cachedFile != nil {
			http.ServeFile(res, req, cachedFile.Path)
			return
		}
		res.WriteHeader(404)
		return
	} else if req.Method == "POST" {
		PublishEvent("download", map[string]string{"url": url, "aliases": ""})
		res.Header().Set("Content-Length", "0")
		res.WriteHeader(202)
		return
	}
	res.Header().Set("Content-Length", "0")
	res.WriteHeader(404)
	return
}

func getValues(req *http.Request, keys []string) map[string]string {
	values := make(map[string]string)
	if req.Method == "GET" || req.Method == "HEAD" {
		queryValues := req.URL.Query()
		for _, key := range keys {
			values[key] = queryValues.Get(key)
		}
	} else {
		req.ParseForm()
		formValues := req.Form
		for _, key := range keys {
			values[key] = formValues.Get(key)
		}
	}
	return values
}
