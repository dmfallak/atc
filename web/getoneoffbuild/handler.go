package getoneoffbuild

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/concourse/atc/builds"
	"github.com/concourse/atc/config"
	"github.com/concourse/atc/db"
	"github.com/pivotal-golang/lager"
)

type handler struct {
	logger lager.Logger

	jobs     config.Jobs
	db       db.DB
	template *template.Template
}

func NewHandler(logger lager.Logger, db db.DB, template *template.Template) http.Handler {
	return &handler{
		logger: logger,

		db:       db,
		template: template,
	}
}

type TemplateData struct {
	Builds []builds.Build

	Build   builds.Build
	Inputs  []db.BuildInput
	Outputs []db.BuildOutput

	Abortable bool
}

func (handler *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	buildIDStr := r.FormValue(":build_id")
	if len(buildIDStr) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	buildID, err := strconv.Atoi(buildIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log := handler.logger.Session("get-one-off-build", lager.Data{
		"build": buildID,
	})

	build, err := handler.db.GetBuild(buildID)
	if err != nil {
		log.Error("get-build-failed", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	inputs, outputs, err := handler.db.GetBuildResources(build.ID)
	if err != nil {
		log.Error("failed-to-get-build-resources", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var abortable bool
	switch build.Status {
	case builds.StatusPending, builds.StatusStarted:
		abortable = true
	default:
		abortable = false
	}

	templateData := TemplateData{
		Build:     build,
		Inputs:    inputs,
		Outputs:   outputs,
		Abortable: abortable,
	}

	err = handler.template.Execute(w, templateData)
	if err != nil {
		log.Fatal("failed-to-execute-template", err, lager.Data{
			"template-data": templateData,
		})
	}
}
