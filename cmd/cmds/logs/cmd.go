package logs

import (
	"fmt"
	"io"
	"os"

	"github.com/blevesearch/bleve/v2"
	"github.com/mattn/go-zglob/fastwalk"
	"github.com/nxadm/tail"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"
)

var Cmd = &cli.Command{
	Name:  "logs",
	Usage: "logs query",
	Action: func(ctx *cli.Context) error {
		defer xerror.RespExit()

		config := tail.Config{Follow: true}
		n := int64(0)
		if ctx.NArg() < 1 {
			fmt.Println("need one or more files as arguments")
			os.Exit(1)
		}

		mapping := bleve.NewIndexMapping()
		index, err := bleve.New("logs/bleve", mapping)
		if err == bleve.ErrorIndexPathExists {
			index, err = bleve.Open("logs/bleve")
			xerror.Panic(err)
		}

		//bleve.NewQueryStringQuery()

		//pquery := bleve.NewTermQuery(strings.Join(args[1:], " "))
		//pquery.SetField()
		////(query, limit, skip, explain)
		//bleve.NewSearchRequestOptions()

		go searchInit(index, "logs", ":8095")

		if n != 0 {
			config.Location = &tail.SeekInfo{Offset: -n, Whence: io.SeekEnd}
		}

		done := make(chan bool)
		xerror.Panic(fastwalk.FastWalk(ctx.Args().First(), func(path string, typ os.FileMode) error {
			if typ.IsDir() {
				return nil
			}

			go tailFile(index, path, config, done)

			return nil
		}))
		select {}
	},
}
