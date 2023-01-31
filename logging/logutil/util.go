package logutil

import (
	"strings"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/try"
)

func LogOrErr(log log.Logger, msg string, fn func() error) {
	msg = strings.TrimSpace(msg)
	log = log.WithCallerSkip(1)

	var err = try.Try(fn)
	if errors.IsNil(err) {
		log.Info().Msg(msg)
	} else {
		log.Err(err).Msg(msg)
	}
}

func OkOrFailed(log log.Logger, msg string, fn func() error) {
	log = log.WithCallerSkip(1)
	log.Info().Msg(msg)

	var err = try.Try(fn)
	if errors.IsNil(err) {
		log.Info().Msg(msg + " ok")
	} else {
		log.Err(err).Msg(msg + " failed")
	}
}

func ErrRecord(log log.Logger, err error) bool {
	if errors.IsNil(err) {
		return false
	}

	log.WithCallerSkip(1).Err(err).Msg(err.Error())
	return true
}
