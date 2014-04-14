package main

import (
	"github.com/codegangsta/cli"
	"github.com/codegangsta/martini"
	"github.com/ngerakines/tram/app"
	"os"
)

func main() {
	tram := cli.NewApp()
	tram.Name = "tram"
	tram.Usage = "Fetch and cache remote content."
	tram.Flags = []cli.Flag{
		cli.StringFlag{"config", "tram.json", "The configuration file to use."},
	}
	tram.Version = "1.0.0"
	tram.Author = "Nick Gerakines"
	tram.Email = "nick@gerakines.net"
	tram.Action = func(c *cli.Context) {
		m := martini.Classic()
		m.Use(app.NewFileCacheWithConfig(app.DefaultDiskFileCacheConfig))

		r := martini.NewRouter()
		r.Any("/", app.HandleIndex)

		m.Action(r.Handle)
		m.Run()
	}

	tram.Run(os.Args)
}
