package resource

import "github.com/pubgo/xerror"

var (
	Err         = xerror.New("resource")
	ErrKindNull = Err.New("kind and name should not be null")
)
