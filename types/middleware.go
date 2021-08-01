package types

import "context"

type MiddleNext func(ctx context.Context, req Request, resp func(rsp Response) error) error
type Middleware func(next MiddleNext) MiddleNext
