package bbolt

import (
	"context"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/lug/tracing"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	bolt "go.etcd.io/bbolt"
)

type DB struct {
	*bolt.DB
}

func (t *DB) Bucket(name string, tx *bolt.Tx) *bolt.Bucket {
	var bk, err = tx.CreateBucketIfNotExists(strutil.ToBytes(name))
	xerror.Panic(err, "create bucket error")
	return bk
}

func (t *DB) View(ctx context.Context, fn func(*bolt.Tx) error) error {
	var span = tracing.FromCtx(ctx).CreateChild(Name)
	defer span.Finish()

	ext.DBType.Set(span, Name)

	return xerror.Wrap(t.DB.View(func(tx *bolt.Tx) (err error) {
		defer xerror.RespErr(&err)
		return fn(tx)
	}))
}

func (t *DB) Update(ctx context.Context, fn func(*bolt.Tx) error) error {
	var span = tracing.FromCtx(ctx).CreateChild(Name)
	defer span.Finish()

	ext.DBType.Set(span, Name)

	return xerror.Wrap(t.DB.Update(func(tx *bolt.Tx) (err error) {
		defer xerror.RespErr(&err)
		return fn(tx)
	}))
}
