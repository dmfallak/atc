package exec

import "github.com/concourse/atc"

type BuildConfigSource interface {
	FetchConfig(ArtifactSource) (atc.BuildConfig, error)
}

type DirectConfigSource struct {
	Config atc.BuildConfig
}

func (source DirectConfigSource) FetchConfig(ArtifactSource) (atc.BuildConfig, error) {
	return source.Config, nil
}

type FileConfigSource struct {
	Path string
}

func (source FileConfigSource) FetchConfig(ArtifactSource) (atc.BuildConfig, error) {

	return atc.BuildConfig{}, nil
}
