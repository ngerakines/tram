package app

import (
	"bytes"
	"encoding/json"
	"github.com/bmizerany/pat"
	"github.com/ngerakines/tram/config"
	"github.com/rcrowley/go-metrics"
	"net/http"
	"strconv"
)

type adminBlueprint struct {
	base      string
	registry  metrics.Registry
	appConfig config.AppConfig
}

type errorViewError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type errorsView struct {
	Errors []errorViewError `json:"errors"`
}

// NewAdminBlueprint creates a new adminBlueprint object.
func newAdminBlueprint(registry metrics.Registry, appConfig config.AppConfig) *adminBlueprint {
	blueprint := new(adminBlueprint)
	blueprint.base = "/admin"
	blueprint.registry = registry
	blueprint.appConfig = appConfig
	return blueprint
}

func (blueprint *adminBlueprint) AddRoutes(p *pat.PatternServeMux) {
	p.Get(blueprint.base+"/config", http.HandlerFunc(blueprint.configHandler))
	p.Get(blueprint.base+"/errors", http.HandlerFunc(blueprint.errorsHandler))
	p.Get(blueprint.base+"/metrics", http.HandlerFunc(blueprint.metricsHandler))
}

func (blueprint *adminBlueprint) configHandler(res http.ResponseWriter, req *http.Request) {
	content := blueprint.appConfig.Source()
	res.Header().Set("Content-Length", strconv.Itoa(len(content)))
	res.Write([]byte(content))
}

func (blueprint *adminBlueprint) metricsHandler(res http.ResponseWriter, req *http.Request) {
	content := &bytes.Buffer{}
	enc := json.NewEncoder(content)
	enc.Encode(blueprint.registry)
	res.Header().Set("Content-Length", strconv.Itoa(content.Len()))
	res.Write(content.Bytes())
}

func (blueprint *adminBlueprint) errorsHandler(res http.ResponseWriter, req *http.Request) {
	view := new(errorsView)
	view.Errors = make([]errorViewError, 0, 0)
	for _, err := range AllErrors {
		view.Errors = append(view.Errors, errorViewError{err.Error(), err.Description()})
	}
	body, err := json.Marshal(view)
	if err != nil {
		res.WriteHeader(500)
		return
	}

	res.Header().Set("Content-Length", strconv.Itoa(len(body)))
	res.Write(body)
}
