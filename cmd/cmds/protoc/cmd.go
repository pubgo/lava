package protoc

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emicklei/proto"
	"github.com/mattn/go-zglob/fastwalk"
	"github.com/pkg/browser"
	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/pkg/gutil"
	"github.com/pubgo/lug/pkg/shutil"
)

func Cmd() *cobra.Command {
	var protoRoot = "proto"

	var cmd = &cobra.Command{
		Use:   "protoc",
		Short: "protoc generate",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			content := xerror.PanicBytes(ioutil.ReadFile("protobuf.yaml"))
			xerror.Panic(yaml.Unmarshal(content, &cfg))
		},
	}
	cmd.Flags().StringVar(&protoRoot, "root", protoRoot, "protobuf directory")

	cmd.AddCommand(
		&cobra.Command{
			Use: "gen",
			Run: func(cmd *cobra.Command, args []string) {
				var protoList = make(map[string]struct{})
				xerror.Panic(fastwalk.FastWalk(protoRoot, func(path string, typ os.FileMode) error {
					if typ.IsDir() {
						return nil
					}

					protoList[filepath.Dir(path)] = struct{}{}
					return nil
				}))

				for in := range protoList {
					var data = fmt.Sprintf("protoc -I . -I %s", protoPath)
					for name, out := range cfg.Plugins {
						if len(out) > 0 {
							data += fmt.Sprintf(" --%s_out=%s", name, strings.Join(out, ","))
						}
					}
					data = data + " " + filepath.Join(in, "*.proto")

					xerror.Panic(shutil.Bash(data).Run(), data)
					fmt.Print("\n\n")
				}

				// 把生成的openapi嵌入到go代码
				var shell = `go-bindata -fs -pkg docs -o docs/docs.go -prefix docs/ -ignore=docs\\.go docs/...`
				xerror.Panic(shutil.Bash(shell).Run())

				// swagger加载和注册
				var code = gutil.CodeFormat(
					"package docs",
					`import "github.com/pubgo/lug/plugins/swagger"`,
					fmt.Sprintf("// build time: %s", time.Now().Format(consts.DefaultTimeFormat)),
					`func init() {swagger.Init(AssetNames, MustAsset)}`,
				)

				const path = "docs/swagger.go"
				_ = os.RemoveAll(path)
				xerror.Panic(ioutil.WriteFile(path, []byte(code), 0755))
			},
		},

		&cobra.Command{
			Use: "ls",
			Run: func(cmd *cobra.Command, args []string) {
				var infoList, err = ioutil.ReadDir(protoPath)
				xerror.Panic(err)

				for _, info := range infoList {
					colorInfo.Println(info.Name())
				}
			},
		},

		&cobra.Command{
			Use: "vendor-rm",
			Run: func(cmd *cobra.Command, args []string) {
				_ = os.RemoveAll(protoPath)
			},
		},

		&cobra.Command{
			Use: "open",
			Run: func(cmd *cobra.Command, args []string) {
				xerror.Panic(browser.OpenFile(protoPath))
			},
		},

		&cobra.Command{
			Use: "api",
			Run: func(cmd *cobra.Command, args []string) {
				var (
					Ctx       = context.Background()
					Task      = make(chan struct{}, 1)
					Stop      = make(chan struct{}, 1)
					Package   = os.Getenv("PROTO_DIR")
					Imports   = make(map[string][]*proto.Import)
					ProtoFile string
				)

				const (
					Annotations = "google/api/annotations.proto"
				)

				var handleImport = func(i *proto.Import) {
					Imports[i.Filename] = append(Imports[i.Filename], i)
				}

				var InsertImport = func(s *proto.Service) error {
					fileBytes, err := ioutil.ReadFile(ProtoFile)
					if err != nil {
						return fmt.Errorf("ioutil.ReadFile %s err: %s ", ProtoFile, err)
					}

					offset := 0
					if s.Comment != nil {
						offset = s.Comment.Position.Offset - s.Comment.Position.Column
					} else {
						offset = s.Position.Offset - s.Position.Column
					}

					insert := fmt.Sprintf("import \"%s\"; \n", Annotations)
					fileBytes = InsertByteSlice(fileBytes, []byte(insert), offset)
					return ioutil.WriteFile(ProtoFile, fileBytes, 0777)
				}

				var CheckImport = func(s *proto.Service) {
					if _, ok := Imports[Annotations]; !ok {
						Task <- struct{}{}
						if err := InsertImport(s); err != nil {
							Stop <- struct{}{}
							fmt.Printf("InsertImport error: %v", err.Error())
						}
						fmt.Println("Import annotations.")
					}
				}

				var InsertOption = func(r *proto.RPC) error {
					fileBytes, err := ioutil.ReadFile(ProtoFile)
					if err != nil {
						return fmt.Errorf("ioutil.ReadFile %s err: %v ", ProtoFile, err)
					}

					if _, ok := r.Parent.(*proto.Service); !ok {
						return fmt.Errorf("ioutil.ReadFile %s err: %v ", ProtoFile, err)
					}

					insert := fmt.Sprintf(`rpc %s (%s) returns (%s) {
        option (google.api.http) = {
          post: "/%s/%s/%s"
          body: "*"
        };
    }`, r.Name, r.RequestType, r.ReturnsType, Package, snakeString(r.Parent.(*proto.Service).Name), snakeString(r.Name))

					// rpc 结束方式很多种
					// {}
					// ;
					// {};
					// {} ;
					// {
					//  } ;
					// {
					//  }
					end := r.Position.Offset
					for {
						if end >= len(fileBytes) {
							Stop <- struct{}{}
							return fmt.Errorf(" Invalid rpc format")
						}

						if fileBytes[end] == '}' {
							next := end
							for {
								next++
								if len(fileBytes) <= next {
									break
								} else if fileBytes[next] == '\n' {
									break
								} else if fileBytes[next] == ';' {
									end = next
									break
								}
							}
							end++
							break
						}

						if fileBytes[end] == ';' {
							end++
							break
						}
						end++
					}

					fileBytes = ReplaceByteSlice(fileBytes, []byte(insert), r.Position.Offset, end)

					return ioutil.WriteFile(ProtoFile, fileBytes, 0777)
				}

				var handleService = func(s *proto.Service) {

					for _, element := range s.Elements {

						select {
						case <-Task:
							Task <- struct{}{}
							return
						case <-Stop:
							Stop <- struct{}{}
						default:
						}

						CheckImport(s)

						select {
						case <-Task:
							Task <- struct{}{}
							return
						case <-Stop:
							Stop <- struct{}{}
							return
						default:
						}

						if rpc, ok := element.(*proto.RPC); ok {
							if len(rpc.Options) == 0 {
								Task <- struct{}{}
								if err := InsertOption(rpc); err != nil {
									Stop <- struct{}{}
								}
								fmt.Printf("Rpc %s Insert option.\n", rpc.Name)
								return
							}
						}
					}
				}
				var Walk = func(cancelFunc context.CancelFunc, protoFile string) {
					reader, err := os.Open(protoFile)
					if err != nil {
						fmt.Printf("os.Open error: %s \n", err.Error())
						cancelFunc()
						return
					}
					defer reader.Close()

					parser := proto.NewParser(reader)
					definition, err := parser.Parse()
					if err != nil {
						fmt.Printf("proto.NewParser error: %s \n", err.Error())
						cancelFunc()
						return
					}

					proto.Walk(
						definition,
						proto.WithImport(handleImport),
						proto.WithService(handleService),
					)

					select {
					case <-Stop:
						cancelFunc()
					case <-Task:
					default:
						cancelFunc()
						fmt.Println("Done.")
					}
				}

				if len(os.Args) < 2 {
					fmt.Println("Invalid proto file")
					os.Exit(1)
				}
				ProtoFile = os.Args[1]
				fmt.Printf("Start checking proto file :%s \n", ProtoFile)

				if Package != "" {
					Package = strings.ToLower(strings.Split(Package, "-")[0])
				} else {
					Package = "micro"
				}

				ctx, cancel := context.WithCancel(Ctx)
				for {
					select {
					case <-ctx.Done():
						return
					default:
						Walk(cancel, ProtoFile)
					}
				}
			},
		},
		&cobra.Command{
			Use: "vendor",
			Run: func(cmd *cobra.Command, args []string) {
				for _, dep := range cfg.Depends {
					var url = dep.Url

					if pathutil.Exist(filepath.Join(modPath, url)) {
						url = filepath.Join(modPath, url)
					}

					if !pathutil.Exist(url) {
						continue
					}

					url = xerror.PanicStr(filepath.Abs(url))

					zap.L().Debug("proto url", zap.String("url", url))
					var newUrl = filepath.Join(protoPath, dep.Path)
					xerror.Panic(filepath.Walk(url, func(path string, info fs.FileInfo, err error) error {
						if info.IsDir() {
							return nil
						}

						if !strings.HasSuffix(info.Name(), ".proto") {
							return nil
						}

						var newPath = filepath.Join(newUrl, strings.TrimPrefix(path, url))
						xerror.Panic(pathutil.IsNotExistMkDir(filepath.Dir(newPath)))
						xerror.PanicErr(copyFile(newPath, path))

						return nil
					}))
				}
			},
		},
	)
	return cmd
}

func copyFile(dstFilePath string, srcFilePath string) (written int64, err error) {
	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		fmt.Printf("打开源文件错误，错误信息=%v\n", err)
	}

	defer srcFile.Close()
	reader := bufio.NewReader(srcFile)

	dstFile, err := os.OpenFile(dstFilePath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		fmt.Printf("打开目标文件错误，错误信息=%v\n", err)
		return
	}

	writer := bufio.NewWriter(dstFile)
	defer dstFile.Close()
	return io.Copy(writer, reader)
}
