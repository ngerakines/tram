package main

import (
	"github.com/codegangsta/martini"
	"github.com/ngerakines/tram/app"
	"os"
	"path/filepath"
)

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	cacheDirectory := filepath.Join(pwd, ".cache2")
	os.MkdirAll(cacheDirectory, 00777)

	m := martini.Classic()
	m.Use(app.NewFileCacheWithPath(cacheDirectory))

	r := martini.NewRouter()
	r.Any("/", app.HandleIndex)

	m.Action(r.Handle)
	m.Run()
}
