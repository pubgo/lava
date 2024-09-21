//go:build tools
// +build tools

package tools

import (
	_ "github.com/ecordell/optgen"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/pubgo/protobuild/pkg/retag"
	_ "golang.org/x/tools/cmd/stringer"
	_ "golang.org/x/vuln/cmd/govulncheck"
)
