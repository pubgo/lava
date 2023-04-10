package bbolt

import (
	"io/fs"
	"time"

	"github.com/pubgo/funk/merge"
	bolt "go.etcd.io/bbolt"
)

const Name = "bolt"

type Config struct {
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

func (t *Config) getOpts() *bolt.Options {
	options := bolt.DefaultOptions
	options.Timeout = time.Second * 2
	return merge.Struct(options, t).Unwrap()
}

func DefaultConfig() *Config {
	return &Config{
		Path:     "./db/bolt",
		FileMode: 0o600,
		Timeout:  time.Second * 2,
	}
}
