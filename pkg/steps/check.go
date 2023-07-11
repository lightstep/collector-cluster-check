package steps

import (
	"context"
)

type Check struct {
	name        string
	description string
	steps       []Step
	depMap      map[string]Dependency
}

func NewCheck(name string, description string, steps []Step) *Check {
	return &Check{
		name:        name,
		description: description,
		steps:       steps,
		depMap:      map[string]Dependency{},
	}
}

func (c *Check) Name() string {
	return c.name
}

func (c *Check) Description() string {
	return c.description
}

func (c *Check) Run(ctx context.Context, deps *Deps, conf *Config) ([]Results, []Results) {
	var acc []Results
	var depAcc []Results
	for _, step := range c.steps {
		depResults := c.initDeps(ctx, step.Dependencies(conf), deps, conf)
		depAcc = append(depAcc, depResults...)
		r := step.Run(ctx, deps)
		acc = append(acc, r)
		if r.ShouldStop() {
			break
		}
	}
	return depAcc, acc
}

func (c *Check) Dependencies(conf *Config) []Dependency {
	return nil
}

func (c *Check) initDeps(ctx context.Context, dependencies []Dependency, deps *Deps, conf *Config) []Results {
	var results []Results
	for _, dep := range dependencies {
		depResults := c.initDeps(ctx, dep.Dependencies(conf), deps, conf)
		results = append(results, depResults...)
		if _, ok := c.depMap[dep.Name()]; ok {
			continue
		}
		opt, r := dep.Run(ctx, deps)
		results = append(results, NewResults(dep, r))
		if !r.Successful() && !r.ShouldContinue() {
			return results
		}
		c.depMap[dep.Name()] = dep
		opt(deps)
	}
	return results
}
