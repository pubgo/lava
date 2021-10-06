package protoutil

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"unicode"

	"github.com/flosch/pongo2/v4"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pubgo/xerror"
	options "google.golang.org/genproto/googleapis/api/annotations"
	gp "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Append(s *string, args ...string) {
	*s += "\n"
	*s += strings.Join(args, "\n")
	*s += "\n"
}

// baseName
// returns the last path element of the name, with the last dotted suffix removed.
func baseName(name string) string {
	// First, find the last element
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	// Now drop the suffix
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[0:i]
	}
	return name
}

// getGoPackage
// returns the file's go_package option.
// If it contains a semicolon, only the part before it is returned.
func getGoPackage(fd *descriptor.FileDescriptorProto) string {
	pkg := fd.GetOptions().GetGoPackage()
	if strings.Contains(pkg, ";") {
		parts := strings.Split(pkg, ";")
		if len(parts) > 2 {
			log.Fatalf("protoc-gen-nrpc: go_package '%s' contains more than 1 ';'", pkg)
		}
		pkg = parts[1]
	}

	return pkg
}

// goPackageOption
// interprets the file's go_package option.
// If there is no go_package, it returns ("", "", false).
// If there's a simple name, it returns ("", Pkg, true).
// If the option implies an import path, it returns (impPath, Pkg, true).
func goPackageOption(d *descriptor.FileDescriptorProto) (impPath, pkg string, ok bool) {
	pkg = getGoPackage(d)
	if pkg == "" {
		return
	}

	ok = true
	// The presence of a slash implies there's an import path.
	slash := strings.LastIndex(pkg, "/")
	if slash < 0 {
		return
	}

	impPath, pkg = pkg, pkg[slash+1:]
	// A semicolon-delimited suffix overrides the package name.
	sc := strings.IndexByte(impPath, ';')
	if sc < 0 {
		return
	}

	impPath, pkg = impPath[:sc], impPath[sc+1:]
	return
}

// goPackageName
// returns the Go package name to use in the
// generated Go file.  The result explicit reports whether the name
// came from an option go_package statement.  If explicit is false,
// the name was derived from the protocol buffer's package statement
// or the input file name.
func goPackageName(d *descriptor.FileDescriptorProto) (name string, explicit bool) {
	// Does the file have a "go_package" option?
	//if _, pkg, ok := goPackageOption(d); ok {
	//	return pkg, true
	//}

	// Does the file have a package clause?
	if pkg := d.GetPackage(); pkg != "" {
		return pkg, false
	}

	// Use the file base name.
	return baseName(d.GetName()), false
}

func getTypeName(pkg, mth string) string {
	mth = strings.TrimSpace(mth)
	mth = strings.Trim(mth, ".")
	mth = strings.TrimPrefix(mth, pkg)
	mth = strings.Trim(mth, ".")
	return mth
}

// Is c an ASCII lower-case letter?
func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// Is c an ASCII digit?
func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

// CamelCase
// returns the CamelCased name.
// If there is an interior underscore followed by a lower case letter,
// drop the underscore and convert the letter to upper case.
// There is a remote possibility of this rewrite causing a name collision,
// but it's so remote we're prepared to pretend it's nonexistent - since the
// C++ generator lowercases names, it's extremely unlikely to have two fields
// with different capitalizations.
// In short, _my_field_name_2 becomes XMyFieldName_2.
func CamelCase(s string) string {
	if s == "" {
		return ""
	}
	t := make([]byte, 0, 32)
	i := 0
	if s[0] == '_' {
		// Need a capital letter; drop the '_'.
		t = append(t, 'X')
		i++
	}
	// Invariant: if the next letter is lower case, it must be converted
	// to upper case.
	// That is, we process a word at a time, where words are marked by _ or
	// upper case letter. Digits are treated as words.
	for ; i < len(s); i++ {
		c := s[i]
		if c == '_' && i+1 < len(s) && isASCIILower(s[i+1]) {
			continue // Skip the underscore in s.
		}
		if isASCIIDigit(c) {
			t = append(t, c)
			continue
		}
		// Assume we have a letter now - if not, it's a bogus identifier.
		// The next word is a sequence of characters that must start upper case.
		if isASCIILower(c) {
			c ^= ' ' // Make it a capital letter.
		}
		t = append(t, c) // Guaranteed not lower case.
		// Accept lower case sequence that follows.
		for i+1 < len(s) && isASCIILower(s[i+1]) {
			i++
			t = append(t, s[i])
		}
	}
	return string(t)
}

// DefaultAPIOptions
// This generates an HttpRule that matches the gRPC mapping to HTTP/2 described in
// https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md#requests
// i.e.:
//   * method is POST
//   * path is "<pkg name>/<service name>/<method name>"
//   * body should contain the serialized request message
func DefaultAPIOptions(pkg string, srv string, mth string) *options.HttpRule {
	return &options.HttpRule{
		Pattern: &options.HttpRule_Post{
			Post: "/" + camel2Case(fmt.Sprintf("%s/%s/%s", camel2Case(pkg), camel2Case(srv), camel2Case(mth))),
		},
		Body: "*",
	}
}

func ExtractAPIOptions(mth protoreflect.MethodDescriptor) (*options.HttpRule, error) {
	if mth == nil {
		return nil, nil
	}

	if !gp.HasExtension(mth.Options(), options.E_Http) {
		return nil, nil
	}

	ext := gp.GetExtension(mth.Options(), options.E_Http)
	opts, ok := ext.(*options.HttpRule)
	if !ok {
		return nil, xerror.Fmt("extension is %T; want an HttpRule", ext)
	}

	return opts, nil
}

func ExtractHttpMethod(opts *options.HttpRule) (method string, path string) {
	var (
		httpMethod   string
		pathTemplate string
	)

	switch {
	case opts.GetGet() != "":
		httpMethod = "GET"
		pathTemplate = opts.GetGet()

	case opts.GetPut() != "":
		httpMethod = "PUT"
		pathTemplate = opts.GetPut()

	case opts.GetPost() != "":
		httpMethod = "POST"
		pathTemplate = opts.GetPost()

	case opts.GetDelete() != "":
		httpMethod = "DELETE"
		pathTemplate = opts.GetDelete()

	case opts.GetPatch() != "":
		httpMethod = "PATCH"
		pathTemplate = opts.GetPatch()

	case opts.GetCustom() != nil:
		custom := opts.GetCustom()
		httpMethod = custom.Kind
		pathTemplate = custom.Path

	default:
		return "", ""
	}

	return httpMethod, pathTemplate
}

func UnExport(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// camel2Case
// 驼峰式写法转为下划线写法
func camel2Case(name string) string {
	name = trim(name)
	buf := new(bytes.Buffer)
	for i, r := range name {
		if !unicode.IsUpper(r) {
			buf.WriteRune(r)
			continue
		}

		if i != 0 {
			buf.WriteRune('-')
		}
		buf.WriteRune(unicode.ToLower(r))
	}
	return strings.NewReplacer(".", "-", "_", "-", "--", "-").Replace(buf.String())
}

func trim(s string) string {
	return strings.Trim(strings.TrimSpace(s), ".-_/")
}

type Context = pongo2.Context

func Template(tpl string, m pongo2.Context) string {
	m["unExport"] = UnExport

	temp, err := pongo2.FromString(tpl)
	xerror.PanicF(err, tpl)

	w := bytes.NewBuffer(nil)
	xerror.PanicF(temp.ExecuteWriter(m, w), tpl)
	return w.String()
}

func goZeroValue(f *descriptor.FieldDescriptorProto) string {
	const nilString = "nil"
	if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		return nilString
	}
	switch *f.Type {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return "0.0"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return "0.0"
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		return "0"
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		return "0"
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		return "0"
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		return "0"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return "false"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return "\"\""
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		return nilString
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return "0"
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		return nilString
	default:
		return nilString
	}
}

func goPkg(f *descriptor.FileDescriptorProto) string {
	return f.Options.GetGoPackage()
}

func goPkgLastElement(f *descriptor.FileDescriptorProto) string {
	pkg := goPkg(f)
	pkgSplitted := strings.Split(pkg, "/")
	return pkgSplitted[len(pkgSplitted)-1]
}

func httpBody(m *descriptor.MethodDescriptorProto) string {
	ext, err := proto.GetExtension(m.Options, options.E_Http)
	if err != nil {
		return err.Error()
	}
	opts, ok := ext.(*options.HttpRule)
	if !ok {
		return fmt.Sprintf("extension is %T; want an HttpRule", ext)
	}
	return opts.Body
}

func httpVerb(m *descriptor.MethodDescriptorProto) string {
	ext, err := proto.GetExtension(m.Options, options.E_Http)
	if err != nil {
		return err.Error()
	}
	opts, ok := ext.(*options.HttpRule)
	if !ok {
		return fmt.Sprintf("extension is %T; want an HttpRule", ext)
	}

	switch t := opts.Pattern.(type) {
	default:
		return ""
	case *options.HttpRule_Get:
		return "GET"
	case *options.HttpRule_Post:
		return "POST"
	case *options.HttpRule_Put:
		return "PUT"
	case *options.HttpRule_Delete:
		return "DELETE"
	case *options.HttpRule_Patch:
		return "PATCH"
	case *options.HttpRule_Custom:
		return t.Custom.Kind
	}
}

func httpPathsAdditionalBindings(m *descriptor.MethodDescriptorProto) []string {
	ext, err := proto.GetExtension(m.Options, options.E_Http)
	if err != nil {
		panic(err.Error())
	}
	opts, ok := ext.(*options.HttpRule)
	if !ok {
		panic(fmt.Sprintf("extension is %T; want an HttpRule", ext))
	}

	var httpPaths []string
	var optsAdditionalBindings = opts.GetAdditionalBindings()
	for _, optAdditionalBindings := range optsAdditionalBindings {
		switch t := optAdditionalBindings.Pattern.(type) {
		case *options.HttpRule_Get:
			httpPaths = append(httpPaths, t.Get)
		case *options.HttpRule_Post:
			httpPaths = append(httpPaths, t.Post)
		case *options.HttpRule_Put:
			httpPaths = append(httpPaths, t.Put)
		case *options.HttpRule_Delete:
			httpPaths = append(httpPaths, t.Delete)
		case *options.HttpRule_Patch:
			httpPaths = append(httpPaths, t.Patch)
		case *options.HttpRule_Custom:
			httpPaths = append(httpPaths, t.Custom.Path)
		default:
			// nothing
		}
	}

	return httpPaths
}

func ParseRequest(r io.Reader) (*plugin.CodeGeneratorRequest, error) {
	input, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read code generator request: %v", err)
	}
	req := new(plugin.CodeGeneratorRequest)
	if err = proto.Unmarshal(input, req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal code generator request: %v", err)
	}
	return req, nil
}

func ParseParameter(args string) {
	if args == "" {
		return
	}

	for _, arg := range strings.Split(args, ",") {
		spec := strings.SplitN(arg, "=", 2)
		if len(spec) == 1 {
			xerror.PanicF(flag.CommandLine.Set(spec[0], ""), "Cannot set flag %s", args)
			continue
		}

		key, value := spec[0], spec[1]
		if strings.HasPrefix(key, "M") {
			continue
		}

		xerror.PanicF(flag.CommandLine.Set(key, value), "Cannot set flag %s", arg)
	}
}

func SourceCode(buf *bytes.Buffer) (string, error) {
	code, err := format.Source(buf.Bytes())
	return string(code), err
}

