package config

import (
	"github.com/ngerakines/tram/util"
	"os"
	"path/filepath"
)

func NewDefaultAppConfig() string {
	return buildDefaultConfig(defaultBasePath)
}

func NewDefaultAppConfigWithBaseDirectory(root string) string {
	return buildDefaultConfig(func(section string) string {
		cacheDirectory := filepath.Join(root, ".cache", section)
		os.MkdirAll(cacheDirectory, 00777)
		return cacheDirectory
	})
}

func buildDefaultConfig(basePathFunc basePath) string {
	return `{
   "listen": ":7040",
   "lruSize": 120000,
   "index": {
     "engine": "local",
     "localBasePath": "` + basePathFunc("index") + `"
   },
   "storage": {
      "engine": "local",
      "basePath": "` + basePathFunc("storage") + `"
   }
}`
}

type basePath func(string) string

func defaultBasePath(section string) string {
	cacheDirectory := filepath.Join(util.Cwd(), ".cache", section)
	os.MkdirAll(cacheDirectory, 00777)
	return cacheDirectory
}
