package ctxutil_test

import (
	"context"
	"testing"

	"github.com/aginetwork7/portal-server/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
)

func TestClone(t *testing.T) {
	cc, cancel := context.WithCancel(context.TODO())
	oldCtx := context.WithValue(cc, "hello", "hello")

	newCtx, _ := ctxutil.Clone(oldCtx)
	cancel()

	assert.Equal(t, oldCtx.Err(), context.Canceled)
	assert.Equal(t, newCtx.Value("hello"), "hello")
	assert.Equal(t, newCtx.Err(), nil)
}
