package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pubgo/lava/v2/cmds/lavacurlcmd"
)

func main() {
	if err := lavacurlcmd.New().Run(context.Background()); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
