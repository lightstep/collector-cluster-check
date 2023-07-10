package steps

import (
	"context"
	"fmt"
)

var _ Step = Check{}

type Check struct {
	name        string
	description string
	steps       []Step
	conf        *Config
}

func NewCheck(name string, description string, steps []Step, conf *Config) *Check {
	return &Check{
		name:        name,
		description: description,
		steps:       steps,
		conf:        conf,
	}
}

func (c Check) Name() string {
	return c.name
}

func (c Check) Description() string {
	return c.description
}

func (c Check) Run(ctx context.Context, deps *Deps) (Option, Result) {
	results := map[string]Result{}
	for _, step := range c.steps {
		depResults := c.initDeps(ctx, step, deps, c.conf)
		for key, res := range depResults {
			results[key] = res
		}
		opt, r := step.Run(ctx, deps)
		results[step.Name()] = r
		if !r.Successful() && !r.ShouldContinue() {
			break
		}
		opt(deps)
	}
	return Empty, NewSuccessfulResult(fmt.Sprint(results))
}

func (c Check) Dependencies(conf *Config) []Step {
	return nil
}

func (c Check) initDeps(ctx context.Context, s Step, deps *Deps, conf *Config) map[string]Result {
	results := map[string]Result{}
	for _, dep := range s.Dependencies(conf) {
		depResults := c.initDeps(ctx, dep, deps, conf)
		for key, res := range depResults {
			results[key] = res
		}
		opt, r := dep.Run(ctx, deps)
		results[dep.Name()] = r
		if !r.Successful() && !r.ShouldContinue() {
			return results
		}
		opt(deps)
	}
	return results
}
