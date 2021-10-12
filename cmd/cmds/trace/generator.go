package trace

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	"github.com/pubgo/xerror"
	"golang.org/x/tools/go/ast/astutil"
)

func hasFuncDec(f *ast.File) bool {
	if len(f.Decls) == 0 {
		return false
	}

	for _, dec := range f.Decls {
		_, ok := dec.(*ast.FuncDecl)
		if ok {
			return true
		}
	}

	return false
}

func rewrite(filename string, del bool) ([]byte, error) {
	fileSet := token.NewFileSet()
	oldAST, err := parser.ParseFile(fileSet, filename, nil, parser.ParseComments)
	xerror.Panic(err, "error parsing %s: %w", filename, err)

	if !hasFuncDec(oldAST) {
		return nil, nil
	}

	// import declaration
	var pkgImport func(*token.FileSet, *ast.File, string) bool
	if del {
		pkgImport = astutil.DeleteImport
	} else {
		pkgImport = astutil.AddImport
	}
	pkgImport(fileSet, oldAST, "github.com/pubgo/lava/pkg/functrace")

	// inject code into each function declaration
	addDeferTraceIntoFuncDec(oldAST, del)

	buf := &bytes.Buffer{}
	xerror.Panic(format.Node(buf, fileSet, oldAST), "formatting new code")
	return xerror.PanicBytes(format.Source(buf.Bytes())), nil
}

func addDeferTraceIntoFuncDec(f *ast.File, del bool) {
	for _, dec := range f.Decls {
		fd, ok := dec.(*ast.FuncDecl)
		if ok {
			// inject code to fd
			addDeferStmt(fd, del)
		}
	}
}

func addDeferStmt(fd *ast.FuncDecl, del bool) (added bool) {
	if fd.Body == nil {
		return false
	}

	stmtList := fd.Body.List
	for _, stmt := range stmtList {
		ds, ok := stmt.(*ast.DeferStmt)
		if !ok {
			continue
		}

		// it is a defer stmt
		ce, ok := ds.Call.Fun.(*ast.CallExpr)
		if !ok {
			continue
		}

		se, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		x, ok := se.X.(*ast.Ident)
		if !ok {
			continue
		}

		if x.Name == "functrace" && se.Sel.Name == "Trace" {
			if del {
				fmt.Println(astutil.NodeDescription(fd.Body.List[1]))
				fd.Body.List = fd.Body.List[1:]
			}

			// already exist , return
			return false
		}
	}

	// add one
	ds := &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "functrace"},
					Sel: &ast.Ident{Name: "Trace"},
				},
			},
		},
	}

	newList := make([]ast.Stmt, len(stmtList)+1)
	copy(newList[1:], stmtList)
	newList[0] = ds
	fd.Body.List = newList
	return true
}
