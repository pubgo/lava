package encoding

import (
	"errors"
	"fmt"
	"testing"

	"github.com/pubgo/xerror"
)

func TestName(t *testing.T) {
	fmt.Println(errors.Is(xerror.Wrap(ErrNotFound), Err))
}
