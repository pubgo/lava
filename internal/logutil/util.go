package logutil

import (
	"strings"

	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/try"
)

func HandlerErr(err error) {
	if err == nil || generic.IsNil(err) {
		return
	}

	log.Err(err).CallerSkipFrame(1).Msg(err.Error())
}

func HandleClose(log log.Logger, fn func() error) {
	if fn == nil || log == nil {
		log.Error().Msgf("log and fn are all required")
		return
	}

	err := fn()
	if generic.IsNil(err) {
		return
	}

	log.Err(err).Msg("failed to handle close")
}

func LogOrErr(log log.Logger, msg string, fn func() error) {
	msg = strings.TrimSpace(msg)
	log = log.WithCallerSkip(1)

	err := try.Try(fn)
	if generic.IsNil(err) {
		log.Info().Msg(msg)
	} else {
		log.Err(err).Msg(msg)
	}
}

func OkOrFailed(log log.Logger, msg string, fn func() error) {
	log = log.WithCallerSkip(1)
	log.Info().Msg(msg)

	err := try.Try(fn)
	if generic.IsNil(err) {
		log.Info().Msg(msg + " ok")
	} else {
		log.Err(err).Msg(msg + " failed")
	}
}

func ErrRecord(logger log.Logger, err error, fn func(evt *log.Event) string) {
	if generic.IsNil(err) {
		return
	}

	evt := log.NewEvent()
	msg := fn(evt)
	logger.WithCallerSkip(1).Err(err).Func(log.WithEvent(evt)).Msg(msg)
}
