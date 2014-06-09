package app

import (
	"github.com/bmizerany/pat"
	"github.com/codegangsta/negroni"
	"github.com/etix/stoppableListener"
	"github.com/ngerakines/tram/config"
	"github.com/ngerakines/tram/util"
	"github.com/rcrowley/go-metrics"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type Blueprint interface {
	AddRoutes(p *pat.PatternServeMux)
}

type AppContext struct {
	registry       metrics.Registry
	appConfig      config.AppConfig
	index          Index
	storageManager StorageManager
	fileCache      FileCache
	apiBlueprint   Blueprint
	adminBlueprint Blueprint
	negroni        *negroni.Negroni
	listener       *stoppableListener.StoppableListener
}

func NewApp(appConfig config.AppConfig) (*AppContext, error) {
	log.Println("Creating application with config", appConfig)
	app := new(AppContext)
	app.registry = metrics.NewRegistry()

	metrics.RegisterRuntimeMemStats(app.registry)
	go metrics.CaptureRuntimeMemStats(app.registry, 60e9)

	app.appConfig = appConfig

	var err error
	err = app.initCache()
	if err != nil {
		return nil, err
	}
	err = app.initApis()
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (app *AppContext) Start() {
	httpListener, err := net.Listen("tcp", app.appConfig.Listen())
	if err != nil {
		panic(err)
	}
	app.listener = stoppableListener.Handle(httpListener)

	http.Serve(app.listener, app.negroni)

	if app.listener.Stopped {
		var alive int

		/* Wait at most 5 seconds for the clients to disconnect */
		for i := 0; i < 5; i++ {
			/* Get the number of clients still connected */
			alive = app.listener.ConnCount.Get()
			if alive == 0 {
				break
			}
			log.Printf("%d client(s) still connected…\n", alive)
			time.Sleep(1 * time.Second)
		}

		alive = app.listener.ConnCount.Get()
		if alive > 0 {
			log.Fatalf("Server stopped after 5 seconds with %d client(s) still connected.", alive)
		} else {
			log.Println("Server stopped gracefully.")
			os.Exit(0)
		}
	} else if err != nil {
		log.Fatal(err)
	}
}

func (app *AppContext) Stop() {
	app.listener.Stop <- true
}

func (app *AppContext) initCache() error {
	app.index = newLocalIndex(app.appConfig.Index().LocalBasePath())

	switch app.appConfig.Storage().Engine() {
	case "local":
		{
			app.storageManager = newLocalStorageManager(app.appConfig.Storage().BasePath())
		}
	case "s3":
		{
			s3Client := app.buildStorageS3Client()
			buckets, err := app.appConfig.Storage().S3Buckets()
			if err != nil {
				panic(err)
			}
			app.storageManager = NewS3StorageManager(buckets, s3Client)
		}
	}

	app.fileCache = newDiskFileCache(app.appConfig, app.index, app.storageManager, util.DedupeWrapDownloader(util.DefaultRemoteFileFetcher))
	return nil
}

func (app *AppContext) initApis() error {
	p := pat.New()

	app.apiBlueprint = newApiBlueprint(app.fileCache, app.storageManager)
	app.apiBlueprint.AddRoutes(p)

	app.adminBlueprint = newAdminBlueprint(app.registry, app.appConfig)
	app.adminBlueprint.AddRoutes(p)

	app.negroni = negroni.Classic()
	app.negroni.UseHandler(p)

	return nil
}

func (app *AppContext) buildStorageS3Client() S3Client {
	if app.appConfig.Storage().Engine() != "s3" {
		return nil
	}
	awsKey, err := app.appConfig.Storage().S3Key()
	if err != nil {
		panic(err)
	}
	awsSecret, err := app.appConfig.Storage().S3Secret()
	if err != nil {
		panic(err)
	}
	awsHost, err := app.appConfig.Storage().S3Host()
	if err != nil {
		panic(err)
	}
	log.Println("Creating s3 client with host", awsHost, "key", awsKey, "and secret", awsSecret)
	return NewAmazonS3Client(NewBasicS3Config(awsKey, awsSecret, awsHost))
}
