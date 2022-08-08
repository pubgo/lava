package bbolt

import (
	"io/fs"
	"path/filepath"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/x/pathutil"
	bolt "go.etcd.io/bbolt"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/internal/pkg/merge"
)

const Name = "bolt"

var _ config.Builder[*bolt.DB] = (*Cfg)(nil)

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
	db              *bolt.DB
}

func (t *Cfg) Get() *bolt.DB { return t.db }
func (t *Cfg) Build() (err error) {
	defer recovery.Err(&err)
	var opts = t.getOpts()
	var path = filepath.Join(config.CfgDir, t.Path)
	assert.Must(pathutil.IsNotExistMkDir(filepath.Dir(path)))
	t.db = assert.Must1(bolt.Open(path, t.FileMode, opts))
	return
}

func (t *Cfg) getOpts() *bolt.Options {
	var options = bolt.DefaultOptions
	options.Timeout = consts.DefaultTimeout
	assert.Must(merge.Struct(options, t))
	return options
}

func DefaultCfg() *Cfg {
	return &Cfg{
		Path:     "./db/bolt",
		FileMode: 0600,
		Timeout:  consts.DefaultTimeout,
	}
}
