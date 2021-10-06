package bbolt

import (
	"context"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	bolt "go.etcd.io/bbolt"

	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/tracing"
)

func Get(name ...string) *DB {
	var val, ok = cfgMap[consts.GetDefault(name...)]
	if !ok {
		return nil
	}
	return val.db
}

type DB struct {
	db *bolt.DB
}

func (t *DB) Get() *bolt.DB {
	return t.db
}

func (t *DB) bucket(name string, tx *bolt.Tx) *bolt.Bucket {
	var bk, err = tx.CreateBucketIfNotExists(strutil.ToBytes(name))
	xerror.Panic(err, "create bucket error")
	return bk
}

func (t *DB) View(ctx context.Context, fn func(*bolt.Bucket), names ...string) error {
	name := consts.Default
	if len(names) > 0 {
		name = names[0]
	}

	var span = tracing.CreateChild(ctx, name)
	defer span.Finish()

	ext.DBType.Set(span, Name)

	return t.db.View(func(tx *bolt.Tx) (err error) { return xerror.Try(func() { fn(t.bucket(name, tx)) }) })
}

func (t *DB) Update(ctx context.Context, fn func(*bolt.Bucket), names ...string) error {
	name := consts.Default
	if len(names) > 0 {
		name = names[0]
	}

	var span = tracing.CreateChild(ctx, name)
	defer span.Finish()

	ext.DBType.Set(span, Name)

	return t.db.Update(func(tx *bolt.Tx) (err error) { return xerror.Try(func() { fn(t.bucket(name, tx)) }) })
}
