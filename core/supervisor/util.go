package supervisor

import (
	"context"
	"log/slog"
	"time"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/thejerf/suture/v4"
)

const ServiceTimeout = 10 * time.Second

type FatalErr struct {
	Err    error
	Status ExitStatus
}

// AsFatalErr wraps the given error creating a FatalErr. If the given error
// already is of type FatalErr, it is not wrapped again.
func AsFatalErr(err error, status ExitStatus) (gErr *FatalErr) {
	if errors.As(err, &gErr) {
		return
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

type doneService func()

func (fn doneService) Serve(ctx context.Context) error {
	<-ctx.Done()
	fn()
	return nil
}

func SpecWithDebugLogger() suture.Spec {
	return spec(func(e suture.Event) { log.Debug().Msg(e.String()) })
}

func SpecWithInfoLogger() suture.Spec {
	return spec(infoEventHook())
}

func spec(eventHook suture.EventHook) suture.Spec {
	return suture.Spec{
		EventHook:                eventHook,
		Timeout:                  ServiceTimeout,
		PassThroughPanics:        true,
		DontPropagateTermination: false,
	}
}

// infoEventHook prints service failures and failures to stop services at level
// info. All other events and identical, consecutive failures are logged at
// debug only.
func infoEventHook() suture.EventHook {
	var prevTerminate suture.EventServiceTerminate
	return func(ei suture.Event) {
		m := ei.Map()
		l := slog.Default().With("supervisor", m["supervisor_name"], "service", m["service_name"])
		switch e := ei.(type) {
		case suture.EventStopTimeout:
			l.Warn("Service failed to terminate in a timely manner")
		case suture.EventServicePanic:
			l.Error("Caught a service panic, which shouldn't happen")
			l.Warn(e.String()) //nolint:sloglint
		case suture.EventServiceTerminate:
			if e.ServiceName == prevTerminate.ServiceName && e.Err == prevTerminate.Err {
				l.Debug("Service failed repeatedly", e.Err)
			} else {
				l.Warn("Service failed", e.Err)
			}
			prevTerminate = e
			l.Debug(e.String()) // Contains some backoff statistics
		case suture.EventBackoff:
			l.Debug("Exiting the backoff state")
		case suture.EventResume:
			l.Debug("Too many service failures - entering the backoff state")
		default:
			l.Warn("Unknown suture supervisor event", slog.Any("type", e.Type()))
			l.Warn(e.String()) //nolint:sloglint
		}
	}
}

func CallWithContext(ctx context.Context, fn func() error) error {
	var err error
	done := make(chan struct{})
	go func() {
		defer close(done)
		err = fn()
	}()

	select {
	case <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
