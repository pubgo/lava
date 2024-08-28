package cloudjobs

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalcTimeout(t *testing.T) {
	var ctx = context.Background()
	deadline, ok := ctx.Deadline()
	assert.Equal(t, ok, false)

	cc, _ := context.WithTimeout(ctx, time.Second*5)
	deadline, ok = cc.Deadline()
	assert.Equal(t, ok, true)
	assert.Equal(t, time.Until(deadline) > time.Second*4, true)
}
