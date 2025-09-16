package stdlog

import (
	stdLog "log"
	"testing"

	"github.com/pubgo/funk/v2/log"
)

func TestName(t *testing.T) {
	SetLogger(log.GetLogger("test"))
	stdLog.Print("hello")
}
