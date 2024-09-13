package cloudjobs

import (
	"context"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestCalcTimeout(t *testing.T) {
	var ctx = context.Background()
	deadline, ok := ctx.Deadline()
	assert.Equal(t, ok, false)

	cc := lo.T2(context.WithTimeout(ctx, time.Second*5))
	deadline, ok = cc.A.Deadline()
	assert.Equal(t, ok, true)
	assert.Equal(t, time.Until(deadline) > time.Second*4, true)
}
