package main

import (
	"github.com/codegangsta/martini"
	"github.com/ngerakines/tram/app"
	"net/http"
)

func main() {
	m := martini.Classic()
	m.Use(app.NewFileCacheWithConfig(app.DefaultDiskFileCacheConfig))

	r := martini.NewRouter()
	r.Any("/", app.HandleIndex)

	m.Action(r.Handle)
	http.ListenAndServe(":7040", m)
}
