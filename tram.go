package main

import (
	"github.com/codegangsta/martini"
	"github.com/docopt/docopt-go"
	"github.com/ngerakines/tram/app"
	"github.com/ngerakines/tram/config"
	"log"
	"net/http"
)

func main() {
	usage := `Tram

Usage: tram [--help]
            [--version]
            [--config <file>]

Options:
  -h --help                 Show this screen.
  --version                 Show version.
  -c <FILE> --config <FILE> The configuration file to use. Unless a config
                            file is specified, the following paths will be
                            loaded:
                                ./tram.conf
                                ~/.tram.conf
                                /etc/tram.conf`

	arguments, _ := docopt.Parse(usage, nil, true, "v0.1.0", false)
	configPath := getConfig(arguments)
	appConfig, err := config.LoadAppConfig(configPath)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	log.Println(appConfig)

	m := martini.Classic()
	m.Use(app.NewFileCacheWithConfig(app.DefaultDiskFileCacheConfig))

	r := martini.NewRouter()
	r.Any("/", app.HandleIndex)

	m.Action(r.Handle)
	http.ListenAndServe(":7040", m)
}

func getConfig(arguments map[string]interface{}) string {
	configPath, hasConfigPath := arguments["--config"]
	if hasConfigPath {
		value, ok := configPath.(string)
		if ok {
			return value
		}
	}
	return ""
}
