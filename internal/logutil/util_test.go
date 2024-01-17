package logutil

import (
	"github.com/pkg/errors"
	"testing"
)

func TestName(t *testing.T) {
	HandlerErr(nil)
	HandlerErr(errors.New("test error"))
}
