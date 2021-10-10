package protoc

import (
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pubgo/lug/pkg/env"
)

var (
	protoPath = filepath.Join(filepath.Join(env.Pwd, ".lug"), "proto")
	modPath   = filepath.Join(os.Getenv("GOPATH"), "/pkg/mod")
)

var colorMajorVersion = color.New(color.FgHiYellow)
var colorSuccess = color.New(color.FgHiGreen)
var colorInfo = color.New(color.FgHiGreen)
var colorError = color.New(color.FgHiRed)

var cfg Cfg

type Cfg struct {
	Depends []depend            `yaml:"deps,omitempty"`
	Input   []string            `yaml:"input,omitempty"`
	Plugins map[string][]string `yaml:"plugins,omitempty"`
}

type depend struct {
	Name    string `yaml:"name,omitempty"`
	Url     string `yaml:"url,omitempty"`
	Path    string `yaml:"path,omitempty"`
	Version string `yaml:"version,omitempty"`
}

type plugin struct {
	Name string   `yaml:"name,omitempty"`
	Out  []string `yaml:"out,omitempty"`
}
