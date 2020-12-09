package rsync

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash/crc64"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	badger "github.com/dgraph-io/badger/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/tikdog/tikdog_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xprocess"
	"github.com/spf13/cobra"
	"github.com/twmb/murmur3"
	"go.uber.org/atomic"
)

type SyncFile struct {
	Crc64ecma uint64
	Name      string
	Path      string
	Changed   bool
	Synced    bool
	Size      int64
	Mode      os.FileMode
	ModTime   int64
	IsDir     bool
}

func getBytes(data interface{}) []byte {
	dt, _ := jsoniter.Marshal(data)
	return dt
}

func Hash(data []byte) (hash string) {
	var h = murmur3.New64()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func getCrc64Sum(path string) uint64 {
	var n = time.Now()
	defer func() {
		fmt.Println(path, time.Since(n))
	}()

	dt, err := ioutil.ReadFile(path)
	xerror.Panic(err)

	c := crc64.New(crc64.MakeTable(crc64.ECMA))
	xerror.PanicErr(c.Write(dt))
	return c.Sum64()
}

func printStack() {
	fmt.Println(xprocess.Stack())
}

// 本地文件加载
// 本地存储中，如果已经同步了，那么就不用同步了
//

var syncPrefix = "sync_files"
var ext = "drawio"

var delPrefix = "trash"
var backupPrefix = "backup"

func Md5(path string) string {
	dt, err := ioutil.ReadFile(path)
	xerror.Panic(err)

	c := md5.New()
	xerror.PanicErr(c.Write(dt))
	return base64.StdEncoding.EncodeToString(c.Sum(nil))
}

func checkAndBackup(dir string, kk *oss.Bucket) {
	var handle = func(path string) {
		fmt.Println(path, "backup")

		var g = xprocess.NewGroup()
		defer g.Wait()
		xerror.Exit(filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				if info.Name()[0] == '.' {
					return filepath.SkipDir
				}
				return nil
			}

			// 隐藏文件
			if info.Name()[0] == '.' {
				return nil
			}

			key := filepath.Join(backupPrefix, path)
			xlog.Infof("backup: %s", path)
			g.Go(func(ctx context.Context) {
				xerror.Panic(kk.PutObjectFromFile(key, path, oss.ContentMD5(Md5(path))))
				time.AfterFunc(time.Second*5, func() { _ = os.Remove(path) })
			})

			return nil
		}))
	}

	xerror.Exit(filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		if info.Name()[0] == '.' {
			return filepath.SkipDir
		}

		if info.Name() == backupPrefix {
			handle(path)
			return filepath.SkipDir
		}

		return nil
	}))
}

func checkAndSync(dir string, kk *oss.Bucket, db *badger.DB, ext string, c *atomic.Bool) {
	if tikdog_util.IsNotExist(dir) {
		return
	}

	fmt.Println("checking", dir)

	var handle = func(ctx context.Context, sf SyncFile) {
		key := filepath.Join(syncPrefix, sf.Path)

		if !sf.Synced {
			var ccc uint64
			head, err := kk.GetObjectMeta(key)
			if err != nil && !strings.Contains(err.Error(), "StatusCode=404") {
				xerror.Panic(err)
			}

			if head != nil {
				ccc, err = strconv.ParseUint(head.Get("X-Oss-Hash-Crc64ecma"), 10, 64)
				xerror.Panic(err)
			}

			if ccc != sf.Crc64ecma {
				xlog.Infof("sync: %s %s", key, sf.Path)
				xerror.Exit(kk.PutObjectFromFile(
					key, sf.Path,
					oss.ContentMD5(Md5(sf.Path)),
				))
			}
			sf.Changed = true
			sf.Synced = true
		}

		if sf.Changed {
			c.Store(true)
			xerror.Exit(db.Update(func(txn *badger.Txn) error {
				sf.Changed = false
				xlog.Infof("store: %s %s", key, sf.Path)
				return xerror.Wrap(txn.Set([]byte(Hash([]byte(key))), getBytes(sf)))
			}))
		}
	}

	var g = xprocess.NewGroup()
	defer g.Wait()
	xerror.Exit(filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name()[0] == '.' || info.Name() == backupPrefix {
				return filepath.SkipDir
			}

			return nil
		}

		// 隐藏文件
		if info.Name()[0] == '.' {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ext) {
			return nil
		}

		key := []byte(filepath.Join(syncPrefix, path))

		return xerror.Wrap(db.View(func(txn *badger.Txn) error {
			itm, err := txn.Get([]byte(Hash(key)))
			if err == badger.ErrKeyNotFound {
				fmt.Println("ErrKeyNotFound:", string(key))
				g.Go(func(ctx context.Context) {
					handle(ctx, SyncFile{
						Name:      info.Name(),
						Size:      info.Size(),
						Mode:      info.Mode(),
						ModTime:   info.ModTime().Unix(),
						IsDir:     info.IsDir(),
						Synced:    false,
						Changed:   true,
						Path:      path,
						Crc64ecma: getCrc64Sum(path),
					})
				})
				return nil
			}

			xerror.Panic(err)

			xerror.Panic(itm.Value(func(_val []byte) error {
				var sf SyncFile
				xerror.Panic(jsoniter.Unmarshal(_val, &sf))
				if sf.ModTime == info.ModTime().Unix() {
					return nil
				}

				sf.Name = info.Name()
				sf.Size = info.Size()
				sf.Mode = info.Mode()
				sf.ModTime = info.ModTime().Unix()
				sf.IsDir = info.IsDir()
				sf.Changed = true

				if hash := getCrc64Sum(path); sf.Crc64ecma != hash {
					sf.Synced = false
					sf.Crc64ecma = hash
				}

				g.Go(func(ctx context.Context) { handle(ctx, sf) })
				return nil
			}))
			return nil
		}))
	}))
}

func checkAndMove(kk *oss.Bucket, db *badger.DB, c *atomic.Bool) {
	var handle = func(sf SyncFile) {
		if !tikdog_util.IsNotExist(sf.Path) {
			return
		}

		c.Store(true)
		xlog.Infof("delete:%s", sf.Path)

		xerror.Panic(OssMove(kk, filepath.Join(syncPrefix, sf.Path), filepath.Join(delPrefix, sf.Path)))
		xerror.Panic(db.Update(func(txn *badger.Txn) error { return xerror.Wrap(txn.Delete([]byte(sf.Name))) }))
	}

	g := xprocess.NewGroup()
	defer g.Wait()
	xerror.Exit(db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			if !bytes.HasPrefix(item.Key(), []byte(syncPrefix)) {
				continue
			}

			xerror.Panic(item.Value(func(v []byte) error {
				var sf SyncFile
				xerror.Panic(jsoniter.Unmarshal(v, &sf))
				sf.Name = string(item.Key())
				handle(sf)
				//g.Go(func(ctx context.Context) { handle(sf) })
				return nil
			}))
		}

		return nil
	}))

}

func GetDbCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "db"}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		var prefix string
		if len(args) > 0 {
			prefix = args[0]
		}

		//var code = "true"
		//if len(args) > 1 {
		//	code = args[1]
		//}

		//program, err := expr.Compile(code, expr.Env(&SyncFile{}))
		//xerror.Panic(err)

		dbPath := filepath.Join(golug_env.Home, "db")
		opts := badger.DefaultOptions(dbPath)
		opts.WithLoggingLevel(badger.DEBUG)

		db, err := badger.Open(opts)
		xerror.Panic(err)
		defer db.Close()

		xerror.Exit(db.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchSize = 10

			it := txn.NewIterator(opts)
			defer it.Close()

			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()

				if !bytes.HasPrefix(item.Key(), []byte(prefix)) {
					continue
				}

				xerror.Panic(item.Value(func(v []byte) error {
					var sf SyncFile
					xerror.Panic(jsoniter.Unmarshal(v, &sf))
					//output, err := expr.Run(program, &sf)
					//xerror.Panic(err)
					//if output.(bool) {
					fmt.Println(string(item.Key()), string(v))
					//}

					return nil
				}))
			}

			return nil
		}))
	}
	return cmd
}

func OssMove(k *oss.Bucket, srcObjectKey, destObjectKey string) error {
	xlog.Infof("copy: %s %s", srcObjectKey, destObjectKey)
	_, err := k.CopyObject(srcObjectKey, destObjectKey)
	if err != nil {
		if strings.Contains(err.Error(), "StatusCode=404") {
			return nil
		}

		return xerror.Wrap(err)
	}

	xlog.Infof("delete: %s", srcObjectKey)
	return xerror.Wrap(k.DeleteObject(srcObjectKey))
}
