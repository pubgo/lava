package resource

import "errors"

var (
	ErrKindNull = errors.New("resource: kind and name should not be null")
)
