package exec

type Step interface {
	Using(ArtifactSource) ArtifactSource
}

func Compose(Step, Step) Step {
	return nil
}

type Aggregate map[string]Step

func (a Aggregate) Using(source ArtifactSource) ArtifactSource {
	sources := aggregateArtifactSource{}

	for name, step := range a {
		sources[name] = step.Using(source)
	}

	return sources
}
