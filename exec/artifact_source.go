package exec

import (
	"io"
	"os"

	garden "github.com/cloudfoundry-incubator/garden/api"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

//go:generate counterfeiter . ArtifactSource
type ArtifactSource interface {
	ifrit.Runner

	StreamOut(string) (io.Reader, error)

	Release() error
}

// type VersionedArtifactSource interface {
// 	ArtifactSource
//
// 	Version() atc.Version
// }

type containerArtifactSource struct {
	Container garden.Container
	RootPath  string
}

type aggregateArtifactSource map[string]ArtifactSource

func (source aggregateArtifactSource) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	members := make(grouper.Members, 0, len(source))

	for name, runner := range source {
		members = append(members, grouper.Member{
			Name:   name,
			Runner: runner,
		})
	}

	return grouper.NewParallel(os.Interrupt, members).Run(signals, ready)
}

func (source aggregateArtifactSource) StreamOut(string) (io.Reader, error) {
	return nil, nil
}

func (source aggregateArtifactSource) Release() error {
	// release all sources
	return nil
}
