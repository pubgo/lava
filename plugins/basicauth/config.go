package basicauth

import (
	"errors"
	"github.com/pubgo/lava/service/service_type"
)

var cfg Cfg

type Cfg struct {
	Realm        string
	Authenticate func(user, pwd string) error
	Authorize    func(user string, req service_type.Request) error
}

//"basicAuth"

//errors
var (
	ErrInvalidBase64 = errors.New("invalid base64")
	ErrNoHeader      = errors.New("not authorized")
	ErrInvalidAuth   = errors.New("invalid authentication")
)
