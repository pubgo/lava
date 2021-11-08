package trace

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mattn/go-zglob/fastwalk"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/pkg/clix"
)

func Cmd() *cli.Command {
	var del bool
	var cmd = &cli.Command{
		Name:        "trace",
		Usage:       "func trace",
		Description: clix.ExampleFmt(`lava trace path`),
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "del",
				Value:       false,
				Usage:       "delete inject trace",
				Destination: &del,
			},
		},
		Action: func(ctx *cli.Context) error {
			defer xerror.RespExit()

			var path = "./"
			if ctx.NArg() > 0 {
				path = ctx.Args().First()
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

			return nil

		},
	}
	return cmd
}
