package config

import (
	"github.com/ngerakines/tram/util"
	"os"
	"path/filepath"
)

func NewDefaultAppConfig() (AppConfig, error) {
	return buildDefaultConfig(defaultBasePath)
}

func NewDefaultAppConfigWithBaseDirectory(root string) (AppConfig, error) {
	return buildDefaultConfig(func() string {
		cacheDirectory := filepath.Join(root, ".cache")
		os.MkdirAll(cacheDirectory, 00777)
		return cacheDirectory
	})
}

func buildDefaultConfig(basePathFunc basePath) (AppConfig, error) {
	config := `{
   "listen": ":7040",
   "lruSize": 120000,
   "storage": {
      "engine": "local",
      "basePath": "` + basePathFunc() + `"
   }
}`
	return NewUserAppConfig([]byte(config))
}

type basePath func() string

func defaultBasePath() string {
	cacheDirectory := filepath.Join(util.Cwd(), ".cache")
	os.MkdirAll(cacheDirectory, 00777)
	return cacheDirectory
}
