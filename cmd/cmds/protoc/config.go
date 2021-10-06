package protoc

import (
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
)

var (
	protoPath = filepath.Join(os.Getenv("GOPATH"), "proto")
	modPath   = filepath.Join(os.Getenv("GOPATH"), "/pkg/mod")
)

var colorMajorVersion = color.New(color.FgHiYellow)
var colorSuccess = color.New(color.FgHiGreen)
var colorInfo = color.New(color.FgHiGreen)
var colorError = color.New(color.FgHiRed)

func init() {
	xerror.Panic(pathutil.IsNotExistMkDir(protoPath))
}

var cfg Cfg

type Cfg struct {
	Depends []depend            `json:"deps,omitempty"`
	Input   []string            `json:"input,omitempty"`
	Plugins map[string][]string `json:"plugins,omitempty"`
}

type depend struct {
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
	Path string `json:"path,omitempty"`
}

type plugin struct {
	Name string   `json:"name,omitempty"`
	Out  []string `json:"out,omitempty"`
}
