package goimportdot

import (
	"fmt"
	"os"

	"github.com/pubgo/lug/cmds/lug/goimportdot/core"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

func GetCmd() *cobra.Command {
	var ignoreGit = true
	var ignoreTest = true
	var onlySelfPkg = true

	var packageName = ""
	var root = ""
	var filters = ""

	var level = -1
	var args = func(cmd *cobra.Command) *cobra.Command {
		flag := cmd.Flags()
		flag.BoolVar(&ignoreGit, "ignoregit", ignoreGit, "ignore files in git")
		flag.BoolVar(&ignoreTest, "ignoretest", ignoreTest, "ignore test files")
		flag.BoolVar(&onlySelfPkg, "only", onlySelfPkg, "only to draw the input package")
		flag.StringVar(&filters, "filter", "", "filter to (ignore/only include) package match wildcard,example: -filter=w:a*,*b;b:c means only include package start with a and ends with b, ignore package named c")
		flag.StringVar(&root, "root", root, "only draw package with the graph start from root")
		flag.IntVar(&level, "level", level, "show how many level , -1 for all")
		flag.StringVar(&packageName, "pkg", packageName, "the package to draw")
		return cmd
	}

	return args(&cobra.Command{
		Use:   "pkg",
		Short: "go 模块依赖分析",
		Run: func(_ *cobra.Command, args []string) {
			defer xerror.RespExit()

			if packageName == "" {
				fmt.Println("You must specify the packge name with -pkg ")
				return
			}

			fileFilter := []core.FileFilter{
				core.HasSuffix(false, ".go"),
			}
			if ignoreGit {
				fileFilter = append(fileFilter, core.NameContains(true, ".git"))
			}
			if ignoreTest {
				fileFilter = append(fileFilter, core.NameContains(true, "_test.go"))
			}

			pkgAndImports, err := core.GetImports(packageName, fileFilter...)
			xerror.Panic(err)

			var pkgFilters []core.PkgFilter
			if onlySelfPkg {
				pkgFilters = append(pkgFilters, core.PkgWildcardFilter(false, packageName+"*"))
			}
			if root != "" {
				pkgFilters = append(pkgFilters, core.RootFilter(root))
			}
			moreFilters, err := core.ParsePkgWildcardStr(filters)
			xerror.PanicF(err, "No right filter [%s], please check!", filters)

			pkgFilters = append(pkgFilters, moreFilters...)

			if level >= 0 {
				pkgFilters = append(pkgFilters, core.PkgLevelFilter(level))
			}

			for _, f := range pkgFilters {
				pkgAndImports = f(pkgAndImports)
			}
			xerror.Panic(core.WriteDot(pkgAndImports, os.Stdout))
		},
	})
}
