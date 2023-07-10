package steps

import (
	"context"
)

type Shutdown func(ctx context.Context) error

type Result struct {
	// successful if the check completed without error
	successful bool

	// err is any error that the step encountered
	err error

	// shouldContinue is whether the check should continue
	shouldContinue bool

	// message is an optional string that informs something about this check
	message string

	// shutdown is an optional function that is called at the end of a check's execution
	shutdown Shutdown
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

func (r Result) Shutdown() Shutdown {
	return r.shutdown
}

func NewSuccessfulResult(message string) Result {
	return Result{
		successful:     true,
		shouldContinue: true,
		message:        message,
	}
}

func NewSuccessfulResultWithShutdown(message string, shutdown Shutdown) Result {
	return Result{
		successful:     true,
		shouldContinue: true,
		message:        message,
		shutdown:       shutdown,
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

func NewFailureResultWithHelp(err error, help string) Result {
	return Result{
		successful:     false,
		err:            err,
		message:        help,
		shouldContinue: false,
	}
}

type Step interface {
	// Name is a single word identifier
	Name() string

	// Description is the optional explanation for what the step does
	Description() string

	// Run can return an option that should be applied to the config object
	// Result represents whether the run was successful
	Run(ctx context.Context, deps *Deps) (Option, Result)

	// Dependencies is a list of steps that must be run prior to this step
	Dependencies(conf *Config) []Step
}
