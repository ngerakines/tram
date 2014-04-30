package app

import (
	"github.com/codegangsta/martini"
	"log"
	"net/http"
)

func NewFileCacheWithConfig(config DiskFileCacheConfig) martini.Handler {
	fileCache := NewDiskFileCache(config)
	return NewFileCacheMiddleware(fileCache)
}

func NewFileCacheMiddleware(fileCache FileCache) martini.Handler {
	return func(_ http.ResponseWriter, _ *http.Request, c martini.Context, _ *log.Logger) {
		c.MapTo(fileCache, (*FileCache)(nil))
		c.Next()
	}
}
