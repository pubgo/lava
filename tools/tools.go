// +build tools

package tools

import (
	_ "gitea.com/xorm/reverse"
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/fatih/gomodifytags"
	_ "github.com/golang/mock/mockgen"
	_ "github.com/google/wire/cmd/wire"
	_ "github.com/izumin5210/gex/cmd/gex"
	_ "github.com/rakyll/statik"
	_ "golang.org/x/lint/golint"
	_ "golang.org/x/tools/cmd/goimports"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
