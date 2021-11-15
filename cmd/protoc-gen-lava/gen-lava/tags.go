package gen_lava

import (
	"fmt"
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
	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	"google.golang.org/protobuf/compiler/protogen"
	gp "google.golang.org/protobuf/proto"

	"github.com/pubgo/lava/proto/lava"
)

func GenerateTag(gen *protogen.Plugin, file *protogen.File) {
	defer xerror.RespExit(file.GeneratedFilenamePrefix)

	if len(file.Messages) == 0 {
		return
	}

	var sep = string(os.PathSeparator)
	var rootDirList = strings.Split(strings.Trim(path, sep), sep)

	var path = fmt.Sprintf("%s.pb.go", file.GeneratedFilenamePrefix)
	path = filepath.Join(path, strings.Join(strings.Split(path, sep)[len(rootDirList)-1:], sep))
	if pathutil.IsNotExist(path) {
		return
	}

	var tags = make(StructTags)

	for i := range file.Messages {
		var name, tt = tags.HandleMessage(file.Messages[i])
		if len(tt) == 0 {
			continue
		}

		tags[name] = tt
	}

	if len(tags) == 0 {
		return
	}

	log.Println("retag:", path)

	fs := token.NewFileSet()
	fn, err := parser.ParseFile(fs, path, nil, parser.ParseComments)
	xerror.Exit(err)

	xerror.Exit(Retag(fn, tags))

	var buf strings.Builder
	xerror.Exit(printer.Fprint(&buf, fs, fn))
	xerror.Exit(ioutil.WriteFile(path, []byte(buf.String()), 0755))
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

	//msg.Messages
	//msg.Oneofs
}

// HandleOneOf
// TODO 未实现
func (t StructTags) HandleOneOf(one *protogen.Oneof) map[string]*structtag.Tags {
	return make(map[string]*structtag.Tags)
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

type tags struct {
	tags StructTags
}

func (t *tags) FileNode(n ast.Node) {
	ast.Walk(Visit(func(node ast.Node) (w ast.Visitor) {
		tp, ok := n.(*ast.TypeSpec)
		if !ok {
			return nil
		}

		switch tp.Type.(type) {
		case *ast.StructType:
			ast.Walk(t.StructNode(n), n)
			return nil
		}

		return nil
	}), n)
}

func (t *tags) TypeNode(n ast.Node) {

}

func (t *tags) StructNode(n ast.Node) Visit {
	return nil
}

func (t *tags) FieldNode(n ast.Node) {

}

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
