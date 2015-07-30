package getresource

import (
	"errors"
	"html/template"
	"log"
	"net/http"

	"github.com/concourse/atc"
	"github.com/concourse/atc/api/present"
	"github.com/concourse/atc/auth"
	"github.com/concourse/atc/db"
	"github.com/concourse/atc/web/group"
	"github.com/pivotal-golang/lager"
)

type server struct {
	logger lager.Logger

	validator auth.Validator

	template *template.Template
}

func NewServer(logger lager.Logger, template *template.Template, validator auth.Validator) *server {
	return &server{
		logger: logger,

		validator: validator,

		template: template,
	}
}

type TemplateData struct {
	Resource atc.Resource
	History  []*db.VersionHistory

	FailingToCheck bool
	CheckError     error

	GroupStates  []group.State
	PipelineName string

	HasPagination bool
	HasOlder      bool
	HasNewer      bool
	OlderStartID  int
	NewerStartID  int
}

//go:generate counterfeiter . ResourcesDB

type ResourcesDB interface {
	GetPipelineName() string
	GetConfig() (atc.Config, db.ConfigVersion, error)
	GetResource(string) (db.SavedResource, error)
	GetResourceHistoryCursor(string, int, bool, int) ([]*db.VersionHistory, bool, error)
	GetResourceHistoryMaxID(string) (int, error)
}

var ErrResourceConfigNotFound = errors.New("could not find resource")

func FetchTemplateData(resourceDB ResourcesDB, authenticated bool, resourceName string, id int, searchUpwards bool) (TemplateData, error) {
	config, _, err := resourceDB.GetConfig()
	if err != nil {
		return TemplateData{}, err
	}

	configResource, found := config.Resources.Lookup(resourceName)
	if !found {
		return TemplateData{}, ErrResourceConfigNotFound
	}

	maxID, err := resourceDB.GetResourceHistoryMaxID(configResource.Name)
	if err != nil {
		return TemplateData{}, err
	}

	startingID := maxID

	if id < maxID && id != 0 {
		startingID = id
	}

	history, hasNext, err := resourceDB.GetResourceHistoryCursor(configResource.Name, startingID, searchUpwards, 100)
	if err != nil {
		return TemplateData{}, err
	}

	dbResource, err := resourceDB.GetResource(configResource.Name)
	if err != nil {
		return TemplateData{}, err
	}

	maxIDFromResults := maxID
	var olderStartID int
	var newerStartID int

	if len(history) > 0 {
		maxIDFromResults = history[0].VersionedResource.ID
		minIDFromResults := history[len(history)-1].VersionedResource.ID
		olderStartID = minIDFromResults - 1
		newerStartID = maxIDFromResults + 1
	}

	hasNewer := maxID > maxIDFromResults
	hasOlder := searchUpwards || hasNext

	hasPagination := hasOlder || hasNewer

	resource := present.Resource(configResource, config.Groups, dbResource, authenticated)

	templateData := TemplateData{
		Resource:      resource,
		History:       history,
		HasPagination: hasPagination,
		HasOlder:      hasOlder,
		HasNewer:      hasNewer,
		OlderStartID:  olderStartID,
		NewerStartID:  newerStartID,
		PipelineName:  resourceDB.GetPipelineName(),
		GroupStates: group.States(config.Groups, func(g atc.GroupConfig) bool {
			for _, groupResource := range g.Resources {
				if groupResource == configResource.Name {
					return true
				}
			}

			return false
		}),
	}

	return templateData, nil
}

func (server *server) GetResource(pipelineDB db.PipelineDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resourceName := r.FormValue(":resource")
		authenticated := server.validator.IsAuthenticated(r)
		templateData, err := FetchTemplateData(pipelineDB, authenticated, resourceName, 0, false)

		switch err {
		case ErrResourceConfigNotFound:
			server.logger.Error("could-not-find-resource-in-config", ErrResourceConfigNotFound, lager.Data{
				"resource": resourceName,
			})
			w.WriteHeader(http.StatusNotFound)
			return
		case nil:
			break
		default:
			server.logger.Error("failed-to-build-template-data", err, lager.Data{
				"resource": resourceName,
			})
			http.Error(w, "failed to fetch resources", http.StatusInternalServerError)
			return
		}

		err = server.template.Execute(w, templateData)
		if err != nil {
			log.Fatal("failed-to-task-template", err, lager.Data{
				"template-data": templateData,
			})
		}
	})
}
