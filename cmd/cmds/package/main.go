package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"golang.org/x/tools/go/packages"
)

func main() {
	flag.Parse()

	// Many tools pass their command-line arguments (after any flags)
	// uninterpreted to packages.Load so that it can interpret them
	// according to the conventions of the underlying build system.
	cfg := &packages.Config{Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedImports}
	pkgs, err := packages.Load(cfg, flag.Args()...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		//os.Exit(1)
	}
	if packages.PrintErrors(pkgs) > 0 {
		//os.Exit(1)
	}

	// Print the names of the source files
	// for each package listed on the command line.
	var size int64
	for _, pkg := range pkgs {
		for _, file := range pkg.GoFiles {
			s, err := os.Stat(file)
			if err != nil {
				log.Println(err)
				continue
			}
			size += s.Size()
		}
	}
	fmt.Printf("size of %v is %v b\n", pkgs[0].ID, size)

	size = 0
	for _, pkg := range allPkgs(pkgs) {
		for _, file := range pkg.GoFiles {
			s, err := os.Stat(file)
			if err != nil {
				log.Println(err)
				continue
			}
			size += s.Size()
		}
	}
	fmt.Printf("size of %v and deps is %v b\n", pkgs[0].ID, size)
}

func allPkgs(lpkgs []*packages.Package) []*packages.Package {
	var all []*packages.Package // postorder
	seen := make(map[*packages.Package]bool)
	var visit func(*packages.Package)
	visit = func(lpkg *packages.Package) {
		if !seen[lpkg] {
			seen[lpkg] = true

			// visit imports
			var importPaths []string
			for path := range lpkg.Imports {
				importPaths = append(importPaths, path)
			}
			sort.Strings(importPaths) // for determinism
			for _, path := range importPaths {
				visit(lpkg.Imports[path])
			}

			all = append(all, lpkg)
		}
	}
	for _, lpkg := range lpkgs {
		visit(lpkg)
	}
	return all
}
