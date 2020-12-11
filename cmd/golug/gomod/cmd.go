package gomod

import (
	"bytes"
	"io/ioutil"

	"github.com/pubgo/golug/internal/sh_wrap"
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

			cmd := sh_wrap.GoMod()
			buffer := &bytes.Buffer{}
			cmd.Stdout = buffer
			xerror.Panic(cmd.Run())

			graph := NewModGraph(buffer)
			graph.Keyword = keyword
			graph.FillColor = fillColor
			graph.Parse()

			var w = bytes.NewBuffer(nil)
			xerror.Panic(graph.Render(w))
			xerror.Panic(ioutil.WriteFile("mod.dot", w.Bytes(), 0600))
			xerror.Panic(sh_wrap.GraphViz("mod.dot", "mod.svg"))
		},
	})
}
