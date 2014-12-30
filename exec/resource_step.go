package exec

import (
	"io"
	"os"

	"github.com/concourse/atc/exec/resource"
)

type resourceStep struct {
	Tracker resource.Tracker
	Type    resource.ResourceType

	Action func(resource.Resource, ArtifactSource) (resource.VersionedSource, error)

	ArtifactSource ArtifactSource

	Resource        resource.Resource
	VersionedSource resource.VersionedSource
}

func (step resourceStep) Using(source ArtifactSource) ArtifactSource {
	step.ArtifactSource = source
	return &step
}

func (ras *resourceStep) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	resource, err := ras.Tracker.Init(ras.Type)
	if err != nil {
		return err
	}

	ras.Resource = resource

	ras.VersionedSource, err = ras.Action(resource, ras.ArtifactSource)
	if err != nil {
		return err
	}

	return ras.VersionedSource.Run(signals, ready)
}

func (ras *resourceStep) Release() error {
	return ras.Resource.Release()
}

func (ras *resourceStep) StreamOut(src string) (io.Reader, error) {
	return ras.VersionedSource.StreamOut(src)
}
