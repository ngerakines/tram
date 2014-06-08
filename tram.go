package main

import (
	"github.com/ngerakines/tram/app"
	"github.com/ngerakines/tram/config"
	"log"
)

func main() {

	appConfig, err := config.LoadAppConfig("")
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	previewApp, err := app.NewApp(appConfig)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	previewApp.Start()

}
