package template

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/open2b/scriggo"
	"github.com/open2b/scriggo/builtin"
	"github.com/open2b/scriggo/native"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func main() {
	start := time.Now()

	dstDir, err := os.MkdirTemp(".", "public-temp-*")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = os.RemoveAll(dstDir)
		if err != nil {
			log.Fatal(err)
		}
	}()

	md := goldmark.New(
		goldmark.WithRendererOptions(html.WithUnsafe()),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithExtensions(extension.GFM))

	buildOptions := &scriggo.BuildOptions{
		Globals: make(native.Declarations, len(globals)+1),
		MarkdownConverter: func(src []byte, out io.Writer) error {
			return md.Convert(src, out)
		},
	}
	for n, v := range globals {
		buildOptions.Globals[n] = v
	}

	srcFS := os.DirFS("site")
	err = fs.WalkDir(srcFS, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if name[0] == '.' {
			return nil
		}

		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dstDir, name), 0o700)
		}

		ext := filepath.Ext(name)
		switch ext {
		case ".html", ".gohtml":
			var dir string
			if p := strings.Index(name, "/"); p > 0 {
				dir = name[0:p]
			}
			switch dir {
			case "imports", "layouts", "partials":
				return nil
			}
			fallthrough
		case ".md":
			fpath := strings.TrimSuffix(name, ext)
			buildOptions.Globals["filepath"] = fpath
			template, err := scriggo.BuildTemplate(srcFS, name, buildOptions)
			if err != nil {
				return err
			}
			name := filepath.Join(dstDir, fpath) + ".html"
			fi, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0o600)
			if err != nil {
				return err
			}
			err = template.Run(fi, nil, nil)
			if err == nil {
				err = fi.Close()
			}
		default:
			src, err := srcFS.Open(name)
			if err != nil {
				return err
			}

			name := filepath.Join(dstDir, name)
			err = os.MkdirAll(filepath.Dir(name), 0o700)
			if err != nil {
				return err
			}

			dst, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0o600)
			if err != nil {
				return err
			}

			_, err = io.Copy(dst, src)
			if err == nil {
				_ = src.Close()
				err = dst.Close()
			}
		}
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	err = os.RemoveAll("public")
	if err != nil {
		log.Fatal(err)
	}

	err = os.Rename(dstDir, "public")
	if err != nil {
		log.Fatal(err)
	}

	buildTime := time.Since(start)
	_, _ = fmt.Fprintf(os.Stderr, "Public site generated in %s\n", buildTime)
	return
}

var globals = native.Declarations{
	// crypto
	"hmacSHA1":   builtin.HmacSHA1,
	"hmacSHA256": builtin.HmacSHA256,
	"sha1":       builtin.Sha1,
	"sha256":     builtin.Sha256,

	// encoding
	"base64":            builtin.Base64,
	"hex":               builtin.Hex,
	"marshalJSON":       builtin.MarshalJSON,
	"marshalJSONIndent": builtin.MarshalJSONIndent,
	"md5":               builtin.Md5,
	"unmarshalJSON":     builtin.UnmarshalJSON,

	// html
	"htmlEscape": builtin.HtmlEscape,

	// math
	"abs": builtin.Abs,
	"max": builtin.Max,
	"min": builtin.Min,
	"pow": builtin.Pow,

	// net
	"File":        reflect.TypeOf((*builtin.File)(nil)).Elem(),
	"FormData":    reflect.TypeOf(builtin.FormData{}),
	"form":        (*builtin.FormData)(nil),
	"queryEscape": builtin.QueryEscape,

	// regexp
	"Regexp": reflect.TypeOf(builtin.Regexp{}),
	"regexp": builtin.RegExp,

	// sort
	"reverse": builtin.Reverse,
	"sort":    builtin.Sort,

	// strconv
	"formatFloat": builtin.FormatFloat,
	"formatInt":   builtin.FormatInt,
	"parseFloat":  builtin.ParseFloat,
	"parseInt":    builtin.ParseInt,

	// strings
	"abbreviate":    builtin.Abbreviate,
	"capitalize":    builtin.Capitalize,
	"capitalizeAll": builtin.CapitalizeAll,
	"hasPrefix":     builtin.HasPrefix,
	"hasSuffix":     builtin.HasSuffix,
	"index":         builtin.Index,
	"indexAny":      builtin.IndexAny,
	"join":          builtin.Join,
	"lastIndex":     builtin.LastIndex,
	"replace":       builtin.Replace,
	"replaceAll":    builtin.ReplaceAll,
	"runeCount":     builtin.RuneCount,
	"split":         builtin.Split,
	"splitAfter":    builtin.SplitAfter,
	"splitAfterN":   builtin.SplitAfterN,
	"splitN":        builtin.SplitN,
	"sprint":        builtin.Sprint,
	"sprintf":       builtin.Sprintf,
	"toKebab":       builtin.ToKebab,
	"toLower":       builtin.ToLower,
	"toUpper":       builtin.ToUpper,
	"trim":          builtin.Trim,
	"trimLeft":      builtin.TrimLeft,
	"trimPrefix":    builtin.TrimPrefix,
	"trimRight":     builtin.TrimRight,
	"trimSuffix":    builtin.TrimSuffix,

	// time
	"Duration":      reflect.TypeOf(builtin.Duration(0)),
	"Hour":          time.Hour,
	"Microsecond":   time.Microsecond,
	"Millisecond":   time.Millisecond,
	"Minute":        time.Minute,
	"Nanosecond":    time.Nanosecond,
	"Second":        time.Second,
	"Time":          reflect.TypeOf(builtin.Time{}),
	"date":          builtin.Date,
	"now":           builtin.Now,
	"parseDuration": builtin.ParseDuration,
	"parseTime":     builtin.ParseTime,
	"unixTime":      builtin.UnixTime,
}
