package protoc

import (
	"os"
	"path/filepath"

	"github.com/pubgo/lava/pkg/modutil"
	"github.com/pubgo/lava/runtime"
)

var (
	protoPath = filepath.Join(runtime.Pwd, ".lava", "proto")
	modPath   = filepath.Join(os.Getenv("GOPATH"), "/pkg/mod")
)

func init() {
	goModPath := filepath.Dir(modutil.GoModPath())
	if goModPath == "" {
		panic("没找到项目go.mod文件")
	}
	protoPath = filepath.Join(goModPath, ".lava", "proto")
}

var cfg Cfg

type Cfg struct {
	Version string   `yaml:"version,omitempty"`
	Root    []string `yaml:"root,omitempty"`
	Depends []depend `yaml:"deps,omitempty"`
	Input   []string `yaml:"input,omitempty"`
	Plugins []plugin `yaml:"plugins,omitempty"`
}

type plugin struct {
	Name string      `yaml:"name,omitempty"`
	Out  string      `yaml:"out,omitempty"`
	Opt  interface{} `yaml:"opt,omitempty"`
}

type depend struct {
	Name    string `yaml:"name,omitempty"`
	Url     string `yaml:"url,omitempty"`
	Path    string `yaml:"path,omitempty"`
	Version string `yaml:"version,omitempty"`
}
