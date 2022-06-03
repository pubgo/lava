package errors

import (
	"fmt"
)

type Err struct {
	Err    error
	Msg    string
	Detail string
}

func (e Err) Unwrap() error { return e.Err }

func (e Err) Error() string {
	return fmt.Sprintf("%s, err=%v detail=%s", e.Msg, e.Err, e.Detail)
}
