package resource

import (
	"io"
	"os"

	"github.com/concourse/atc"
	"github.com/tedsuo/ifrit"
)

// Request payload from resource to /opt/resource/out script
type outRequest struct {
	Source atc.Source `json:"source"`
	Params atc.Params `json:"params,omitempty"`
}

func (resource *resource) Put(source atc.Source, params atc.Params, sourceStream io.Reader) VersionedSource {
	vs := &versionedSource{
		container: resource.container,
	}

	vs.Runner = ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
		err := resource.container.StreamIn(ResourcesDir, sourceStream)
		if err != nil {
			return err
		}

		return resource.runScript(
			"/opt/resource/out",
			[]string{ResourcesDir},
			outRequest{
				Params: params,
				Source: source,
			},
			&vs.versionResult,
		).Run(signals, ready)
	})

	return vs
}
