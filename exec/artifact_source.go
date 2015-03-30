package exec

import (
	"errors"
	"io"
	"os"

	"github.com/concourse/atc"
	"github.com/tedsuo/ifrit"
)

var ErrFileNotFound = errors.New("file not found")

//go:generate counterfeiter . Step

type Step interface {
	ifrit.Runner

	Release() error
	Result(interface{}) bool

	ArtifactSource
}

type SourceName string

//go:generate counterfeiter . SourceRepository

type SourceRepository interface {
	RegisterSource(SourceName, ArtifactSource)
	SourceFor(SourceName) (ArtifactSource, bool)
}

//go:generate counterfeiter . ArtifactSource

type ArtifactSource interface {
	StreamTo(ArtifactDestination) error
	StreamFile(path string) (io.ReadCloser, error)
}

//go:generate counterfeiter . ArtifactDestination

type ArtifactDestination interface {
	StreamIn(io.Reader) error
}

type Success bool

type ExitStatus int

type VersionInfo struct {
	Version  atc.Version
	Metadata []atc.MetadataField
}

type NoopStep struct{}

func (NoopStep) Run(<-chan os.Signal, chan<- struct{}) error {
	return nil
}

func (NoopStep) Release() error { return nil }

func (NoopStep) Result(interface{}) bool {
	return false
}
