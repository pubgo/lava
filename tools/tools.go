// +build tools

package tools

import (
	_ "gitea.com/xorm/reverse"
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/fatih/gomodifytags"
	_ "github.com/rakyll/statik"
	_ "golang.org/x/lint/golint"
	_ "golang.org/x/tools/cmd/goimports"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
