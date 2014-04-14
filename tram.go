package main

import (
	"github.com/codegangsta/martini"
	"github.com/ngerakines/tram/app"
)

func main() {
	m := martini.Classic()
	m.Use(app.NewFileCacheWithPath())

	r := martini.NewRouter()
	r.Any("/", app.HandleIndex)

	m.Action(r.Handle)
	m.Run()
}
