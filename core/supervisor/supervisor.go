package supervisor

import (
	"fmt"

	"github.com/thejerf/suture/v4"
)

type Service interface {
	Name() string
	Error() error
	fmt.Stringer
	suture.Service
}

type Supervisor = suture.Supervisor
type Spec = suture.Spec
