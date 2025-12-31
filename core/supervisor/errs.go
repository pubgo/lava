package supervisor

import (
	"github.com/pubgo/funk/v2/errors"
	"github.com/thejerf/suture/v4"
)

type FatalErr struct {
	Err    error
	Status ExitStatus
}

// AsFatalErr wraps the given error creating a FatalErr. If the given error
// already is of type FatalErr, it is not wrapped again.
func AsFatalErr(err error, status ExitStatus) (gErr *FatalErr) {
	if errors.As(err, &gErr) {
		return gErr
	}

	return &FatalErr{
		Err:    err,
		Status: status,
	}
}

func IsFatal(err error) bool {
	return errors.As(err, &FatalErr{})
}

func (e *FatalErr) Error() string {
	return e.Err.Error()
}

func (e *FatalErr) Unwrap() error {
	return e.Err
}

func (*FatalErr) Is(target error) bool {
	return target == suture.ErrTerminateSupervisorTree
}

// NoRestartErr wraps the given error err (which may be nil) to make sure that
// `errors.Is(err, suture.ErrDoNotRestart) == true`.
func NoRestartErr(err error) error {
	if err == nil {
		return suture.ErrDoNotRestart
	}
	return &noRestartErr{err}
}

type noRestartErr struct {
	err error
}

func (e *noRestartErr) Error() string {
	return e.err.Error()
}

func (e *noRestartErr) Unwrap() error {
	return e.err
}

func (*noRestartErr) Is(target error) bool {
	return target == suture.ErrDoNotRestart
}

type ExitStatus int

const (
	ExitSuccess            ExitStatus = 0
	ExitError              ExitStatus = 1
	ExitNoUpgradeAvailable ExitStatus = 2
	ExitRestart            ExitStatus = 3
	ExitUpgrade            ExitStatus = 4
)

func (s ExitStatus) AsInt() int {
	return int(s)
}
