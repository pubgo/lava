package bbolt

import (
	"io/fs"
	"path/filepath"
	"time"

	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	bolt "go.etcd.io/bbolt"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/internal/pkg/merge"
)

const Name = "bolt"

type Cfg struct {
	FileMode        fs.FileMode       `json:"file_mode"`
	Timeout         time.Duration     `json:"timeout"`
	NoGrowSync      bool              `json:"no_grow_sync"`
	NoFreelistSync  bool              `json:"no_freelist_sync"`
	FreelistType    bolt.FreelistType `json:"freelist_type"`
	ReadOnly        bool              `json:"read_only"`
	MmapFlags       int               `json:"mmap_flags"`
	InitialMmapSize int               `json:"initial_mmap_size"`
	PageSize        int               `json:"page_size"`
	NoSync          bool              `json:"no_sync"`
	Path            string            `json:"path"`
}

func (t *Cfg) BuildOpts() *bolt.Options {
	var options = bolt.DefaultOptions
	options.Timeout = consts.DefaultTimeout
	xerror.Panic(merge.Struct(options, t))
	return options

}

func (t *Cfg) Create() *bolt.DB {
	var opts = t.BuildOpts()
	var path = filepath.Join(config.CfgDir, t.Path)
	xerror.Panic(pathutil.IsNotExistMkDir(filepath.Dir(path)))

	db, err := bolt.Open(path, t.FileMode, opts)
	xerror.Panic(err)
	return db
}

func DefaultCfg() *Cfg {
	return &Cfg{
		Path:     "./db/bolt",
		FileMode: 0600,
		Timeout:  consts.DefaultTimeout,
	}
}
