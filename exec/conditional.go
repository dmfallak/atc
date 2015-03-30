package exec

import (
	"io"
	"os"

	"github.com/concourse/atc"
)

type Conditional struct {
	Conditions  atc.Conditions
	StepFactory StepFactory

	prev Step
	repo SourceRepository

	result Step
}

func (c Conditional) Using(prev Step, repo SourceRepository) Step {
	c.prev = prev
	c.repo = repo
	return &c
}

func (c *Conditional) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	var succeeded Success

	conditionMatched := false
	if c.prev.Result(&succeeded) {
		conditionMatched = c.Conditions.SatisfiedBy(bool(succeeded))
	} else {
		// if previous step cannot indicate success, presume that it succeeded
		// for this step to be running in the first place.
		conditionMatched = c.Conditions.SatisfiedBy(true)
	}

	if conditionMatched {
		c.result = c.StepFactory.Using(c.prev, c.repo)
	} else {
		c.result = &NoopStep{}
	}

	return c.result.Run(signals, ready)
}

func (c *Conditional) StreamTo(dst ArtifactDestination) error {
	return c.outputSource.StreamTo(dst)
}

func (c *Conditional) StreamFile(path string) (io.ReadCloser, error) {
	return c.outputSource.StreamFile(path)
}

func (c *Conditional) Release() error {
	return c.result.Release()
}

func (c *Conditional) Result(x interface{}) bool {
	return c.result.Result(x)
}
