package logs

import (
	"fmt"
	"io"
	"os"

	"github.com/nxadm/tail"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/pubgo/lava/pkg/clix"
)

var Cmd = clix.Command(func(cmd *cobra.Command, flags *pflag.FlagSet) {
	config := tail.Config{Follow: true}
	n := int64(0)
	maxlinesize := int(0)

	flags = cmd.Flags()
	flags.Int64Var(&n, "n", 0, "tail from the last Nth location")
	flags.IntVar(&maxlinesize, "max", 0, "max line size")
	flags.BoolVar(&config.Follow, "f", false, "wait for additional data to be appended to the file")
	flags.BoolVar(&config.ReOpen, "F", false, "follow, and track file rename/rotation")
	flags.BoolVar(&config.Poll, "p", false, "use polling, instead of inotify")

	if config.ReOpen {
		config.Follow = true
	}
	config.MaxLineSize = maxlinesize

	cmd.Use = "logs"
	cmd.Short = "logs query"
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("need one or more files as arguments")
			os.Exit(1)
		}

		if n != 0 {
			config.Location = &tail.SeekInfo{Offset: -n, Whence: io.SeekEnd}
		}

		done := make(chan bool)
		for _, filename := range args {
			go tailFile(filename, config, done)
		}

		for range args {
			<-done
		}
	}
})
