package app

import (
	"errors"
	"github.com/bmizerany/pat"
	"log"
	"net/http"
)

type Blueprint interface {
	AddRoutes(p *pat.PatternServeMux)
}

type apiBlueprint struct {
	base      string
	fileCache FileCache
}

func newApiBlueprint(fileCache FileCache) Blueprint {
	blueprint := new(apiBlueprint)
	blueprint.base = "/"
	blueprint.fileCache = fileCache
	return blueprint
}

func (blueprint *apiBlueprint) AddRoutes(p *pat.PatternServeMux) {
	p.Get(blueprint.base, http.HandlerFunc(blueprint.handleGet))
}

func (blueprint *apiBlueprint) handleGet(res http.ResponseWriter, req *http.Request) {
	values := blueprint.getValues(req, []string{"url", "alias"})
	url, err := blueprint.collectUrl(values)
	aliases := blueprint.collectAliases(values)
	if err == nil {
		cachedFile := blueprint.fileCache.WarmAndQuery(url, aliases)
		if cachedFile != nil {
			http.ServeFile(res, req, cachedFile.Location())
			return
		}
	}
	log.Println(err)
	res.Header().Set("Content-Length", "0")
	res.WriteHeader(404)
	return
}

func (blueprint *apiBlueprint) collectAliases(args map[string][]string) []string {
	values, hasValues := args["aliases"]
	if hasValues && values != nil && len(values) > 0 {
		return values
	}
	return []string{}
}

func (blueprint *apiBlueprint) collectUrl(args map[string][]string) (string, error) {
	urls, hasUrls := args["url"]
	if hasUrls && urls != nil && len(urls) > 0 {
		return urls[0], nil
	}
	return "", errors.New("Missing url.")
}

func (blueprint *apiBlueprint) getValues(req *http.Request, keys []string) map[string][]string {
	values := make(map[string][]string)
	if req.Method == "GET" || req.Method == "HEAD" {
		queryValues := req.URL.Query()
		for _, key := range keys {
			value, hasValue := queryValues[key]
			if hasValue {
				values[key] = value
			}
		}
	}
	return values
}
