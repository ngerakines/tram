package main

import (
	"fmt"
	"github.com/docopt/docopt.go"
	"github.com/ngerakines/tram/app"
	"github.com/ngerakines/tram/config"
	"log"
	"os"
	"os/signal"
)

func main() {
	usage := `Tram

Usage: tram [--help --version --config=<file>]
       tram daemon [--help --version --config <file>]

Options:
  --help           Show this screen.
  --version        Show version.
  --config=<file>  The configuration file to use.`

	arguments, _ := docopt.Parse(usage, nil, true, "1.0.0", false)

	command := newDaemonCommand(arguments)
	command.Execute()
}

type daemonCommand struct {
	config string
}

func newDaemonCommand(arguments map[string]interface{}) *daemonCommand {
	command := new(daemonCommand)
	command.config = getCliString(arguments, "--config")
	return command
}

func (command *daemonCommand) String() string {
	return fmt.Sprintf("daemonCommand<config=%s>", command.config)
}

func (command *daemonCommand) Execute() {
	appConfig, err := config.LoadAppConfig(command.config)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	tramApp, err := app.NewApp(appConfig)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	k := make(chan os.Signal, 1)
	signal.Notify(k, os.Interrupt, os.Kill)
	go func() {
		<-k
		tramApp.Stop()
	}()

	tramApp.Start()
}

func getCliString(arguments map[string]interface{}, key string) string {
	configPath, hasConfigPath := arguments[key]
	if hasConfigPath {
		value, ok := configPath.(string)
		if ok {
			return value
		}
	}
	return ""
}

func getCliStringArray(arguments map[string]interface{}, key string) []string {
	configPath, hasConfigPath := arguments[key]
	if hasConfigPath {
		value, ok := configPath.([]string)
		if ok {
			return value
		}
	}
	return []string{}
}

func getCliBool(arguments map[string]interface{}, key string) bool {
	configPath, hasConfigPath := arguments[key]
	if hasConfigPath {
		value, ok := configPath.(bool)
		if ok {
			return value
		}
	}
	return false
}

func getCliInt(arguments map[string]interface{}, key string) int {
	configPath, hasConfigPath := arguments[key]
	if hasConfigPath {
		value, ok := configPath.(int)
		if ok {
			return value
		}
	}
	return 0
}
