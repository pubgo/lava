package shutil

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
)

func Run(args ...string) (r result.Result[string]) {
	defer recovery.Result(&r)

	b := bytes.NewBufferString("")

	cmd := Shell(args...)
	cmd.Stdout = b

	assert.Must(cmd.Run(), strings.Join(args, " "))
	return r.WithVal(strings.TrimSpace(b.String()))
}

func GoModGraph() result.Result[string] {
	return Run("go", "mod", "graph")
}

func GoList() result.Result[string] {
	return Run("go", "list", "./...")
}

func GraphViz(in, out string) (err error) {
	ret := Run("dot", "-Tsvg", in)
	if ret.IsErr() {
		return ret.Err()
	}

	return ioutil.WriteFile(out, []byte(ret.Unwrap()), 0o600)
}

func Shell(args ...string) *exec.Cmd {
	shell := strings.Join(args, " ")
	cmd := exec.Command("/bin/sh", "-c", shell)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd
}
