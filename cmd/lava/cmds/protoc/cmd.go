package protoc

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/emicklei/proto"
	"github.com/pubgo/x/iox"
	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/pkg/modutil"
	"github.com/pubgo/lava/pkg/protoutil"
	"github.com/pubgo/lava/pkg/shutil"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/types"
)

var protoRoot []string
var protoCfg = "protobuf.yaml"

func Cmd() *cli.Command {
	return &cli.Command{
		Name:  "protoc",
		Usage: "protobuf generation, configuration and management",
		Flags: types.Flags{
			&cli.StringFlag{
				Name:        "protobuf",
				Usage:       "protobuf config path",
				Value:       protoCfg,
				Destination: &protoCfg,
			},
		},
		Before: func(ctx *cli.Context) error {
			defer xerror.RespExit()

			xerror.Panic(pathutil.IsNotExistMkDir(protoPath))

			content := xerror.PanicBytes(ioutil.ReadFile(protoCfg))
			xerror.Panic(yaml.Unmarshal(content, &cfg))

			protoRoot = append(protoRoot, cfg.Root...)

			// protobuf文件检查
			for _, dep := range cfg.Depends {
				xerror.Assert(dep.Name == "" || dep.Url == "", "name和url都不能为空")
			}
			return nil
		},
		Subcommands: cli.Commands{
			&cli.Command{
				Name:  "bindata",
				Usage: "gen swagger",
				Action: func(ctx *cli.Context) error {
					// 把生成的openapi嵌入到go代码
					var shell = `go-bindata -fs -pkg docs -o docs/docs.go -prefix docs/ -ignore=docs\\.go docs/...`
					xerror.Panic(shutil.Shell(shell).Run())

					// swagger加载和注册
					var code = lavax.CodeFormat(
						"package docs",
						`import "github.com/pubgo/lava/plugins/swagger"`,
						fmt.Sprintf("// build time: %s", time.Now().Format(consts.DefaultTimeFormat)),
						`func init() {swagger.initCfg(AssetNames, MustAsset)}`,
					)

					const path = "docs/swagger.go"
					_ = os.RemoveAll(path)
					xerror.Panic(ioutil.WriteFile(path, []byte(code), 0755))
					return nil
				},
			},
			&cli.Command{
				Name:  "tidy",
				Usage: "检查缺失protobuf依赖并把版本信息写入protobuf.yaml",
				Action: func(ctx *cli.Context) error {
					defer xerror.RespExit()

					// 解析go.mod并获取所有pkg版本
					var versions = modutil.LoadVersions()
					for i, dep := range cfg.Depends {
						var url = dep.Url

						// url是本地目录, 不做检查
						if pathutil.IsDir(url) {
							continue
						}

						var version = versions[url]
						if version == "" {
							version = dep.Version
						}

						// go.mod中version不存在, 并且protobuf.yaml也没有指定
						if version == "" {
							// go pkg缓存
							var localPkg, err = ioutil.ReadDir(filepath.Dir(filepath.Join(modPath, url)))
							xerror.Panic(err)

							var _, name = filepath.Split(url)
							for j := range localPkg {
								if !localPkg[j].IsDir() {
									continue
								}

								if strings.HasPrefix(localPkg[j].Name(), name+"@") {
									version = strings.TrimPrefix(localPkg[j].Name(), name+"@")
									break
								}
							}
						}

						if version == "" {
							xerror.Panic(shutil.Shell("go", "get", "-d", url+"/...").Run())

							// 再次解析go.mod然后获取版本信息
							versions = modutil.LoadVersions()
							version = versions[url]

							xerror.Assert(version == "", "%s version为空", url)
						}

						cfg.Depends[i].Version = version
					}
					xerror.Panic(ioutil.WriteFile(protoCfg, xerror.PanicBytes(yaml.Marshal(cfg)), 0755))
					return nil
				},
			},
			&cli.Command{
				Name:  "gen",
				Usage: "编译protobuf文件",
				Action: func(ctx *cli.Context) error {
					defer xerror.RespExit()

					var protoList sync.Map

					for i := range protoRoot {
						if pathutil.IsNotExist(protoRoot[i]) {
							log.Printf("file %s not flund", protoRoot[i])
							continue
						}

						xerror.Panic(filepath.Walk(protoRoot[i], func(path string, info fs.FileInfo, err error) error {
							if err != nil {
								return err
							}

							if info.IsDir() {
								return nil
							}

							if !strings.HasSuffix(info.Name(), ".proto") {
								return nil
							}

							protoList.Store(filepath.Dir(path), struct{}{})
							return nil
						}))
					}

					protoList.Range(func(key, _ interface{}) bool {
						var in = key.(string)

						var data = ""
						var base = fmt.Sprintf("protoc -I %s -I %s", protoPath, runtime.Pwd)
						var lavaOut = ""
						var lavaOpt = ""
						for i := range cfg.Plugins {
							var plg = cfg.Plugins[i]

							var name = plg.Name

							var out = func() string {
								// https://github.com/pseudomuto/protoc-gen-doc
								// 目录特殊处理
								if name == "doc" {
									var out = filepath.Join(plg.Out, in)
									xerror.Panic(pathutil.IsNotExistMkDir(out))
									return out
								}

								if plg.Out != "" {
									return plg.Out
								}

								return "."
							}()

							var opts = func(dt interface{}) []string {
								switch _dt := dt.(type) {
								case string:
									if _dt != "" {
										return []string{_dt}
									}
								case []string:
									return _dt
								case []interface{}:
									var dtList []string
									for i := range _dt {
										dtList = append(dtList, _dt[i].(string))
									}
									return dtList
								}
								return nil
							}(plg.Opt)

							if name == "lava" {
								lavaOut = fmt.Sprintf(" --%s_out=%s", name, out)
								lavaOpt = fmt.Sprintf(" --%s_opt=%s", name, strings.Join(opts, ","))
								continue
							}

							data += fmt.Sprintf(" --%s_out=%s", name, out)

							if len(opts) > 0 {
								data += fmt.Sprintf(" --%s_opt=%s", name, strings.Join(opts, ","))
							}
						}
						data = base + data + " " + filepath.Join(in, "*.proto")
						fmt.Println(data + "\n")
						xerror.Panic(shutil.Shell(data).Run(), data)
						data = base + lavaOut + lavaOpt + " " + filepath.Join(in, "*.proto")
						fmt.Println(data + "\n")
						xerror.Panic(shutil.Shell(data).Run(), data)
						return true
					})
					return nil
				},
			},
			&cli.Command{
				Name:  "vendor",
				Usage: "把项目protobuf依赖同步到.lava/proto中",
				Action: func(ctx *cli.Context) error {
					defer xerror.RespExit()

					// 删除老的protobuf文件
					_ = os.RemoveAll(protoPath)

					for _, dep := range cfg.Depends {
						if dep.Name == "" || dep.Url == "" {
							continue
						}

						var url = dep.Url
						var version = dep.Version

						// 加载版本
						if version != "" {
							url = fmt.Sprintf("%s@%s", url, version)
						}

						// 加载路径
						url = filepath.Join(url, dep.Path)

						if !lavax.DirExists(url) {
							url = filepath.Join(modPath, url)
						}

						zap.S().Debug(url)

						url = xerror.PanicStr(filepath.Abs(url))
						var newUrl = filepath.Join(protoPath, dep.Name)
						xerror.Panic(filepath.Walk(url, func(path string, info fs.FileInfo, err error) (gErr error) {
							if err != nil {
								return err
							}

							defer xerror.RespErr(&gErr)

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
					return nil
				},
			},
			&cli.Command{
				Name:  "check",
				Usage: "protobuf文件检查",
				Action: func(ctx *cli.Context) error {
					defer xerror.RespExit()

					var protoList sync.Map
					for i := range protoRoot {
						if pathutil.IsNotExist(protoRoot[i]) {
							log.Printf("proto root (%s) not flund\n", protoRoot[i])
							continue
						}

						xerror.Panic(filepath.Walk(protoRoot[i], func(path string, info fs.FileInfo, err error) error {
							if err != nil {
								return err
							}

							if info.IsDir() {
								return nil
							}

							if !strings.HasSuffix(info.Name(), ".proto") {
								return nil
							}

							protoList.Store(path, struct{}{})
							return nil
						}))
					}

					// 处理检测gateway url
					var handler = func(protoFile string) {
						var data, err = iox.ReadText(protoFile)
						xerror.Panic(err)

						parser := proto.NewParser(strings.NewReader(data))
						definition, err := parser.Parse()
						xerror.Panic(err, protoFile)

						// package name
						var pkg string
						proto.Walk(definition, proto.WithPackage(func(p *proto.Package) {
							var replacer = strings.NewReplacer(".", "/", "-", "/")
							pkg = replacer.Replace(p.Name)
						}))

						var rpcList []*proto.RPC
						proto.Walk(definition, proto.WithService(func(srv *proto.Service) {
							for _, e := range srv.Elements {
								var rpc, ok = e.(*proto.RPC)
								if !ok {
									continue
								}

								rpcList = append(rpcList, rpc)
							}
						}))

						var dataLine = strings.Split(data, "\n")
						for i := range rpcList {
							rpc := rpcList[i]
							insert := fmt.Sprintf(`
rpc %s (%s) returns (%s) {
  option (google.api.http) = {
    post: "%s"
    body: "*"
  };`, rpc.Name, rpc.RequestType, rpc.ReturnsType, "/"+protoutil.Camel2Case(fmt.Sprintf("%s/%s/%s", protoutil.Camel2Case(pkg), protoutil.Camel2Case(rpc.Parent.(*proto.Service).Name), protoutil.Camel2Case(rpc.Name))))

							var hasHttp bool
							for i := range rpc.Options {
								if rpc.Options[i].Name == "(google.api.http)" {
									hasHttp = true
								}
							}

							// 如果option为0, 那么可以整体替换, 通过正则表达式
							if len(rpc.Options) == 0 || !hasHttp {
								_ = insert
								var rpcData = strings.Trim(dataLine[rpc.Position.Line-1], ";")
								// 以}结尾
								if rpcData[len(rpcData)-1] == '}' {
									dataLine[rpc.Position.Line-1] = insert + "\n}\n"
								} else {
									dataLine[rpc.Position.Line-1] = insert
								}
							}
						}

						data = strings.Join(dataLine, "\n")
						xerror.Panic(ioutil.WriteFile(protoFile, []byte(data), 0755))
					}
					protoList.Range(func(key, _ interface{}) bool {
						defer xerror.RespExit(key)
						handler(key.(string))
						return true
					})
					return nil
				},
			},
		},
	}
}

func copyFile(dstFilePath string, srcFilePath string) (written int64, err error) {
	srcFile, err := os.Open(srcFilePath)
	xerror.Panic(err, "打开源文件错误，错误信息")

	defer srcFile.Close()
	reader := bufio.NewReader(srcFile)

	dstFile, err := os.OpenFile(dstFilePath, os.O_WRONLY|os.O_CREATE, 0777)
	xerror.Panic(err, "打开目标文件错误，错误信息")

	writer := bufio.NewWriter(dstFile)
	defer dstFile.Close()
	return io.Copy(writer, reader)
}
