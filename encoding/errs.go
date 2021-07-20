package encoding

import (
	"github.com/pubgo/xerror"
)

var Err = xerror.New(Name)
var ErrNotFound = Err.New("key not found")
