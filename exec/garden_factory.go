package exec

import (
	"github.com/concourse/atc"
	"github.com/concourse/atc/exec/resource"
)

type gardenFactory struct {
	resourceTracker resource.Tracker
}

func NewGardenFactory(resourceTracker resource.Tracker) Factory {
	return &gardenFactory{
		resourceTracker: resourceTracker,
	}
}

func (factory *gardenFactory) Get(config atc.ResourceConfig, params atc.Params, version atc.Version) Step {
	return resourceStep{
		Tracker: factory.resourceTracker,
		Type:    resource.ResourceType(config.Type),

		Action: func(r resource.Resource, s ArtifactSource) (resource.VersionedSource, error) {
			return r.Get(config.Source, params, version), nil
		},
	}
}

func (factory *gardenFactory) Put(config atc.ResourceConfig, params atc.Params) Step {
	return resourceStep{
		Tracker: factory.resourceTracker,
		Type:    resource.ResourceType(config.Type),

		Action: func(r resource.Resource, s ArtifactSource) (resource.VersionedSource, error) {
			stream, err := s.StreamOut(".")
			if err != nil {
				return nil, err
			}

			return r.Put(config.Source, params, stream), nil
		},
	}
}

func (factory *gardenFactory) Execute(BuildConfigSource) Step {
	return nil
}
