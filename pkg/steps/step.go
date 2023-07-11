package steps

import (
	"context"
)

type Results struct {
	results []Result
	d       Describable
}

func (r Results) ShouldStop() bool {
	for _, result := range r.results {
		if !result.Successful() && !result.ShouldContinue() {
			return true
		}
	}
	return false
}

func (r Results) StepName() string {
	return r.d.Name()
}

func (r Results) Steps() []Result {
	return r.results
}

func NewResults(s Describable, r ...Result) Results {
	return Results{d: s, results: r}
}

type Result struct {
	// successful if the check completed without error
	successful bool

	// err is any error that the step encountered
	err error

	// shouldContinue is whether the check should continue
	shouldContinue bool

	// message is an optional string that informs something about this check
	message string
}

func (r Result) Successful() bool {
	return r.successful
}

func (r Result) Err() error {
	return r.err
}

func (r Result) ShouldContinue() bool {
	return r.shouldContinue
}

func (r Result) Message() string {
	return r.message
}

func NewSuccessfulResult(message string) Result {
	return Result{
		successful:     true,
		shouldContinue: true,
		message:        message,
	}
}

func NewFailureResult(err error) Result {
	return Result{
		successful:     false,
		err:            err,
		shouldContinue: false,
	}
}

func NewAcceptableFailureResult(err error) Result {
	return Result{
		successful:     false,
		err:            err,
		shouldContinue: true,
	}
}

func NewAcceptableFailureResultWithHelp(err error, help string) Result {
	return Result{
		successful:     false,
		err:            err,
		message:        help,
		shouldContinue: true,
	}
}

func NewFailureResultWithHelp(err error, help string) Result {
	return Result{
		successful:     false,
		err:            err,
		message:        help,
		shouldContinue: false,
	}
}

type Describable interface {
	// Name is a single word identifier
	Name() string

	// Description is the optional explanation for what the step does
	Description() string
}

type Dependency interface {
	// Describable requires that every Dependency has a name and description
	Describable
	// Run can return an option that should be applied to the config object
	// Result represents whether the run was successful
	Run(ctx context.Context, deps *Deps) (Option, Result)

	// Dependencies is a list of dependencies that must be run prior to this one
	Dependencies(conf *Config) []Dependency

	// Shutdown allows the dependency to gracefully exit
	Shutdown(ctx context.Context) error
}

type Step interface {
	// Describable requires that every Step has a name and description
	Describable
	// Run can return an option that should be applied to the config object
	// Result represents whether the run was successful
	Run(ctx context.Context, deps *Deps) Results

	// Dependencies is a list of steps that must be run prior to this step
	Dependencies(conf *Config) []Dependency
}
