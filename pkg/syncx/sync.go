package syncx

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Group = errgroup.Group

func ErrGroup(contexts ...context.Context) (*Group, context.Context) {
	var ctx = context.Background()
	if len(contexts) > 0 {
		ctx = contexts[0]
	}
	return errgroup.WithContext(ctx)
}
