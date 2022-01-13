package basicauth

import (
	"errors"

	"github.com/pubgo/lava/types"
)

var cfg Cfg

type Cfg struct {
	Realm        string
	Authenticate func(user, pwd string) error
	Authorize    func(user string, req types.Request) error
}

//"basicAuth"

//errors
var (
	ErrInvalidBase64 = errors.New("invalid base64")
	ErrNoHeader      = errors.New("not authorized")
	ErrInvalidAuth   = errors.New("invalid authentication")
)
