package sh_wrap

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"

	"github.com/pubgo/xerror"
)

var buf = &sync.Pool{
	New: func() interface{} {
		return bytes.NewBufferString("")
	},
}

func Run(args ...string) error {
	return Shell(args...).Run()
}

func Shell(args ...string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd
}

func GoMod() *exec.Cmd {
	return Shell("go", "mod", "graph")
}

func GraphViz(in, out string) (err error) {
	defer xerror.RespErr(&err)

	b := buf.Get().(*bytes.Buffer)
	defer buf.Put(b)
	cmd := Shell("dot", "-Tsvg", in)
	cmd.Stdout = b
	xerror.Panic(cmd.Start())
	xerror.Panic(cmd.Wait())
	return ioutil.WriteFile(out, b.Bytes(), 0600)
}
