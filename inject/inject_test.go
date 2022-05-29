package inject

import (
	"fmt"
	"io"
	"testing"

	"github.com/pubgo/xerror"
	"go.uber.org/fx"
)

type njnj struct {
	CC []io.Closer `group:"a"`
}

func TestName(t *testing.T) {
	var ss njnj
	xerror.Exit(fx.New(
		fx.Provide(fx.Annotated{
			Group:  "a",
			Target: func() io.Closer { return nil },
		}),
		fx.Invoke(func(in struct {
			fx.In
			CC []io.Closer `group:"a"`
		}) {
			fmt.Println(in.CC)
		}),
		fx.Populate(&ss),
	).Err())
	fmt.Println(ss.CC)
}
