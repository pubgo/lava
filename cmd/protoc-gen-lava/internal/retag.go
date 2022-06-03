package internal

import (
	"fmt"
	"github.com/pubgo/x/pathutil"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/structtag"
	"github.com/pubgo/xerror"
	"google.golang.org/protobuf/compiler/protogen"
	gp "google.golang.org/protobuf/proto"

	"github.com/pubgo/lava/proto/lava"
)

func GenerateTag(gen *protogen.Plugin, file *protogen.File) {
	defer xerror.RecoverAndExit()

	if len(file.Messages) == 0 {
		return
	}

	var sep = string(os.PathSeparator)
	var rootDirList = strings.Split(strings.Trim(path, sep), sep)
	pbPath := filepath.Join(path, strings.Join(strings.Split(file.GeneratedFilenamePrefix, sep)[len(rootDirList)-1:], sep))
	pbPath = fmt.Sprintf("%s.pb.go", pbPath)
	xerror.Assert(pathutil.IsNotExist(pbPath), "path=>%s not found", pbPath)

	var tagMap = make(StructTags)

	for i := range file.Messages {
		var name, tt = tagMap.HandleMessage(file.Messages[i])
		if len(tt) == 0 {
			continue
		}

		tagMap[name] = tt
	}

	if len(tagMap) == 0 {
		return
	}

	log.Println("retag:", pbPath)

	fs := token.NewFileSet()
	fn, err := parser.ParseFile(fs, pbPath, nil, parser.ParseComments)
	xerror.Exit(err)

	xerror.Exit(Retag(fn, tagMap))

	var buf strings.Builder
	xerror.Exit(printer.Fprint(&buf, fs, fn))
	xerror.Exit(ioutil.WriteFile(pbPath, []byte(buf.String()), 0755))
}

type StructTags map[string]map[string]*structtag.Tags

func (t StructTags) HandleMessage(msg *protogen.Message) (string, map[string]*structtag.Tags) {
	var msgName = msg.GoIdent.GoName
	var tags = make(map[string]*structtag.Tags)

	for i := range msg.Fields {
		var name, tag = t.HandleField(msg.Fields[i])
		if name == "" {
			continue
		}
		tags[name] = tag
	}

	return msgName, tags
}

// HandleOneOf
// TODO 未实现
func (t StructTags) HandleOneOf(one *protogen.Oneof) *structtag.Tags {
	var opts = one.Desc.Options()
	if !gp.HasExtension(opts, lava.E_Tags) {
		return nil
	}

	var tags = gp.GetExtension(opts, lava.E_Tags).([]*lava.Tag)
	if len(tags) == 0 {
		return nil
	}

	var tt = new(structtag.Tags)
	for _, tag := range tags {
		xerror.Panic(tt.Set(&structtag.Tag{Key: tag.Key, Name: tag.Value}))
	}
	return tt
}

func (t StructTags) HandleField(field *protogen.Field) (string, *structtag.Tags) {
	var opts = field.Desc.Options()
	if !gp.HasExtension(opts, lava.E_Tags) {
		return "", nil
	}

	var tags = gp.GetExtension(opts, lava.E_Tags).([]*lava.Tag)
	if len(tags) == 0 {
		return "", nil
	}

	var tt = new(structtag.Tags)
	for _, tag := range tags {
		xerror.Panic(tt.Set(&structtag.Tag{Key: tag.Key, Name: tag.Value}))
	}
	return field.GoName, tt
}

type Visit func(node ast.Node) (w ast.Visitor)

func (v Visit) Visit(node ast.Node) (w ast.Visitor) { return v(node) }

// Retag updates the existing tags with the map passed and modifies existing tags if any of the keys are matched.
// First key to the tags argument is the name of the struct, the second key corresponds to field names.
func Retag(n ast.Node, tags StructTags) error {
	r := retag{}
	f := func(n ast.Node) ast.Visitor {
		if r.err != nil {
			return nil
		}

		if tp, ok := n.(*ast.TypeSpec); ok {
			r.tags = tags[tp.Name.String()]
			return r
		}

		return nil
	}

	ast.Walk(structVisitor{f}, n)

	return r.err
}

type structVisitor struct {
	visitor func(n ast.Node) ast.Visitor
}

func (v structVisitor) Visit(n ast.Node) ast.Visitor {
	if tp, ok := n.(*ast.TypeSpec); ok {
		if _, ok := tp.Type.(*ast.StructType); ok {
			ast.Walk(v.visitor(n), n)
			return nil // This will ensure this struct is no longer traversed
		}
	}
	return v
}

type retag struct {
	err  error
	tags map[string]*structtag.Tags
}

func (v retag) Visit(n ast.Node) ast.Visitor {
	if v.err != nil {
		return nil
	}

	if f, ok := n.(*ast.Field); ok {
		if len(f.Names) == 0 {
			return nil
		}
		newTags := v.tags[f.Names[0].String()]
		if newTags == nil {
			return nil
		}

		if f.Tag == nil {
			f.Tag = &ast.BasicLit{
				Kind: token.STRING,
			}
		}

		oldTags, err := structtag.Parse(strings.Trim(f.Tag.Value, "`"))
		if err != nil {
			v.err = err
			return nil
		}

		for _, t := range newTags.Tags() {
			oldTags.Set(t)
		}

		f.Tag.Value = "`" + oldTags.String() + "`"

		return nil
	}

	return v
}
