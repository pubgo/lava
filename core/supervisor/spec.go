package supervisor

import (
	"log/slog"
	"time"

	"github.com/pubgo/funk/log"
	"github.com/thejerf/suture/v4"
)

const ServiceTimeout = 10 * time.Second

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
