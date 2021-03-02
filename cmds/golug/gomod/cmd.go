package gomod

import (
	"bytes"
	"io/ioutil"
	"strings"

	"github.com/pubgo/x/shutil"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

func GetCmd() *cobra.Command {
	var keyword string
	var fillColor string
	var args = func(cmd *cobra.Command) *cobra.Command {
		flag := cmd.Flags()
		flag.StringVar(&keyword, "k", "", "specific keyword to filter lib")
		flag.StringVar(&fillColor, "c", "yellow", "specific mod node fill color")
		return cmd
	}

	return args(&cobra.Command{
		Use:   "mod",
		Short: "go mod graph 分析",
		Run: func(_ *cobra.Command, args []string) {
			defer xerror.RespExit()

			graph := NewModGraph(strings.NewReader(xerror.PanicStr(shutil.GoMod())))
			graph.Keyword = keyword
			graph.FillColor = fillColor
			graph.Parse()

			var w = bytes.NewBuffer(nil)
			xerror.Panic(graph.Render(w))
			xerror.Panic(ioutil.WriteFile("mod.dot", w.Bytes(), 0600))
			xerror.Panic(shutil.GraphViz("mod.dot", "mod.svg"))
		},
	})
}
