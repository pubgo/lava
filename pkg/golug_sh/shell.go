package golug_sh

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"

	"github.com/pubgo/xerror"
)

var bufPool = &sync.Pool{New: func() interface{} { return bytes.NewBufferString("") }}

func getBuffer() *bytes.Buffer    { return bufPool.Get().(*bytes.Buffer) }
func putBuffer(buf *bytes.Buffer) { buf.Reset(); bufPool.Put(buf) }

func Run(args ...string) (string, error) {
	b := getBuffer()
	defer putBuffer(b)

	cmd := Shell(args...)
	cmd.Stdout = b
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return b.String(), nil
}

func Shell(args ...string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd
}

func GoMod() (string, error) {
	return Run("go", "mod", "graph")
}

func GoList() (string, error) {
	return Run("go", "list", "./...")
}

func GraphViz(in, out string) (err error) {
	ret, err := Run("dot", "-Tsvg", in)
	if err != nil {
		return xerror.WrapF(err, "in:%s, out:%s", in, out)
	}

	return ioutil.WriteFile(out, []byte(ret), 0600)
}
