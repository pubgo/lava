package trace

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mattn/go-zglob/fastwalk"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/pubgo/lava/pkg/lavax"
)

func Cmd() *cobra.Command {
	var del bool
	var flags = pflag.NewFlagSet("", pflag.ContinueOnError)
	flags.BoolVar(&del, "del", false, "delete inject trace")

	var cmd = &cobra.Command{
		Use:     "trace",
		Short:   "func trace",
		Example: lavax.ExampleFmt(`lava trace path`),
		Run: func(cmd *cobra.Command, args []string) {
			defer xerror.RespExit()

			var path = "./"
			if len(args) > 0 {
				path = args[0]
			}

			path = xerror.PanicStr(filepath.Abs(path))
			xerror.Panic(fastwalk.FastWalk(path, func(path string, typ os.FileMode) error {
				defer xerror.RespExit()

				if typ.IsDir() {
					return nil
				}

				if filepath.Ext(path) != ".go" {
					return nil
				}

				newSrc, err := rewrite(path, del)
				if err != nil {
					panic(err)
				}

				if newSrc == nil {
					// add nothing to the source file. no change
					fmt.Printf("no trace added for %s\n", path)
					return nil
				}

				// write to the source file
				if err = ioutil.WriteFile(path, newSrc, typ.Perm()); err != nil {
					fmt.Printf("write %s error: %v\n", path, err)
					return err
				}

				return nil
			}))

			if del {
				fmt.Printf("del trace ok\n")
			} else {
				fmt.Printf("add trace ok\n")
			}

		},
	}
	cmd.Flags().AddFlagSet(flags)

	return cmd
}
