package core

import (
	"bytes"
	"fmt"
	"io"
	rand "math/rand"
	"os"
	"path/filepath"
	"strings"
)

func GetImports(pkg string, filters ...FileFilter) (pkgimports map[string]StrSet, err error) {
	fullpath := ""

	goPath := os.Getenv("GOPATH")
	gopaths := strings.Split(goPath, ":")
	for _, gp := range gopaths {
		fp := filepath.Join(gp, "src", pkg)
		if _, err := os.Stat(fp); err == nil {
			fullpath = fp
		}
	}
	if fullpath == "" {
		err = fmt.Errorf("Can not find package [%s] in GOPATH [%s]", pkg, goPath)
		return
	}
	pkgimports = make(map[string]StrSet)
	filepath.Walk(fullpath, func(fp string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		for _, filter := range filters {
			if !filter.IsBlack {
				continue
			}
			if filter.Func(fp, info, err) {
				return nil
			}
		}
		for _, filter := range filters {
			if filter.IsBlack {
				continue
			}
			if !filter.Func(fp, info, err) {
				return nil
			}
		}
		pkg := PkgOfFile(fp)
		if _, ok := pkgimports[pkg]; !ok {
			pkgimports[pkg] = NewStrSet()
		}
		ss, err := ParseGoImport(fp)
		if err != nil {
			// TODO: better err
			panic(err)
		}
		pkgimports[pkg].Merge(ss)
		return nil
	})
	return
}

var colors = []string{
	"blue",
	"orange",
	"violet",
	"red",
	"coral",
	"black",
	"cyan",
	"orchid",
	"olive",
	"maroon",
	"lightsalmon",
	"indigo",
	"chocolate",
	"aquamarine",
}

func WriteDot(pkgimports map[string]StrSet, writer io.Writer) (err error) {
	nodes := NewStrSet()
	edges := [][2]string{}
	for pkg, imps := range pkgimports {
		nodes.Put(pkg)
		for imp := range imps {
			nodes.Put(imp)
			edges = append(edges, [2]string{pkg, imp})
		}
	}
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString("digraph G {\n")
	buf.WriteString("rankdir=LR;\n")
	buf.WriteString("node [color=lightblue2, style=filled];\n")

	var data = make(map[string][]string)
	for _, edge := range edges {
		data[edge[0]] = append(data[edge[0]], edge[1])
	}

	i := 0
	for k, v := range data {
		cl := colors[rand.Intn(i+1)%len(colors)]
		for _, v1 := range v {
			buf.WriteString(fmt.Sprintf(`"%s"->"%s" [style=bold,color=%s,label = "(%d)" ];`, k, v1, cl, i))
			buf.WriteByte('\n')
		}
		i++
	}

	//for pkg, _ := range nodes {
	//	buf.WriteString(fmt.Sprintf(`"%s";`, pkg))
	//	buf.WriteByte('\n')
	//}

	buf.WriteString("}\n")
	_, err = writer.Write(buf.Bytes())
	return
}
