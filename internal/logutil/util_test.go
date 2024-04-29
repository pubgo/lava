package logutil

import (
	"testing"

	"github.com/pkg/errors"
)

func TestName(t *testing.T) {
	HandlerErr(nil)
	HandlerErr(errors.New("test error"))
}
