package protoc

import (
	"bufio"
	"fmt"
	"github.com/pubgo/x/q"
	"github.com/pubgo/x/strutil"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/emicklei/proto"
	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/clix"
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/pkg/modutil"
	"github.com/pubgo/lava/pkg/shutil"
)

func Cmd() *cobra.Command {
	var protoRoot []string
	var protoCfg = "protobuf.yaml"

	return clix.Command(func(cmd *cobra.Command, flags *pflag.FlagSet) {
		flags.StringVar(&protoCfg, "config", protoCfg, "protobuf config")
		cmd.Use = "protoc"
		cmd.Short = "protobuf generation, configuration and management"
		cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
			defer xerror.RespExit()

			xerror.Panic(pathutil.IsNotExistMkDir(protoPath))

			content := xerror.PanicBytes(ioutil.ReadFile(protoCfg))
			xerror.Panic(yaml.Unmarshal(content, &cfg))

			protoRoot = append(protoRoot, cfg.Root...)
			if !strutil.Contain(protoRoot, "proto") {
				protoRoot = append(protoRoot, "proto")
			}

			// protobuf文件检查
			for _, dep := range cfg.Depends {
				xerror.Assert(dep.Name == "" || dep.Url == "", "name和url都不能为空")
			}
		}

		cmd.AddCommand(&cobra.Command{
			Use:   "bindata",
			Short: "gen swagger",
			Run: func(cmd *cobra.Command, args []string) {
				// 把生成的openapi嵌入到go代码
				var shell = `go-bindata -fs -pkg docs -o docs/docs.go -prefix docs/ -ignore=docs\\.go docs/...`
				xerror.Panic(shutil.Shell(shell).Run())

				// swagger加载和注册
				var code = lavax.CodeFormat(
					"package docs",
					`import "github.com/pubgo/lava/plugins/swagger"`,
					fmt.Sprintf("// build time: %s", time.Now().Format(consts.DefaultTimeFormat)),
					`func init() {swagger.Init(AssetNames, MustAsset)}`,
				)

				const path = "docs/swagger.go"
				_ = os.RemoveAll(path)
				xerror.Panic(ioutil.WriteFile(path, []byte(code), 0755))
			}})

		cmd.AddCommand(&cobra.Command{
			Use:   "tidy",
			Short: "检查缺失protobuf依赖并把版本信息写入protobuf.yaml",
			Run: func(cmd *cobra.Command, args []string) {
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
			},
		})
		cmd.AddCommand(&cobra.Command{
			Use:   "gen",
			Short: "编译protobuf文件",
			Run: func(cmd *cobra.Command, args []string) {
				defer xerror.RespExit()

				var protoList sync.Map

				for i := range protoRoot {
					if pathutil.IsNotExist(protoRoot[i]) {
						zap.S().Warnf("file %s not flund", protoRoot[i])
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

					var data = fmt.Sprintf("protoc -I %s -I %s", protoPath, env.Pwd)
					for name, out := range cfg.Plugins {
						if len(out) > 0 {
							data += fmt.Sprintf(" --%s_out=%s", name, strings.Join(out, ","))
						}
					}
					data = data + " " + filepath.Join(in, "*.proto")

					fmt.Println(data + "\n")
					xerror.Panic(shutil.Shell(data).Run(), data)
					return true
				})
			},
		})
		cmd.AddCommand(&cobra.Command{
			Use:   "vendor",
			Short: "把项目protobuf依赖同步到.lava/proto中",
			Run: func(cmd *cobra.Command, args []string) {
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
		})
		cmd.AddCommand(&cobra.Command{
			Use:   "check",
			Short: "protobuf文件检查",
			Run: func(cmd *cobra.Command, args []string) {
				defer xerror.RespExit()

				var protoList sync.Map
				for i := range protoRoot {
					if pathutil.IsNotExist(protoRoot[i]) {
						zap.S().Warnf("file %s not flund", protoRoot[i])
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

				var handler = func(protoFile string) {
					reader, err := os.Open(protoFile)
					xerror.Panic(err, protoFile)
					defer reader.Close()

					parser := proto.NewParser(reader)
					definition, err := parser.Parse()
					xerror.Panic(err, protoFile)

					proto.Walk(definition, proto.WithRPC(func(rpc *proto.RPC) {
						if rpc.StreamsRequest || rpc.StreamsReturns {
							return
						}

						var hasHttp bool
						for _, e := range rpc.Elements {
							var opt, ok = e.(*proto.Option)
							if !ok {
								continue
							}

							if strings.Contains(opt.Name, "google.api.http") {
								hasHttp = true
							}
						}

						if !hasHttp {
							q.Q(rpc)
							panic(fmt.Errorf("method=>%s path=>%s 请设置gateway url", rpc.Name, protoFile))
						}
					}))
				}

				protoList.Range(func(key, _ interface{}) bool {
					defer xerror.RespExit(key)

					handler(key.(string))
					return true
				})
			},
		})
	})
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
