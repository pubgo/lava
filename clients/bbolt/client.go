package bbolt

import (
	"context"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	bolt "go.etcd.io/bbolt"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/plugins/tracing"
	"github.com/pubgo/lava/resource"
)

type Client struct {
	resource.Resource
}

func (t *Client) Kind() string { return Name }

func (t *Client) bucket(name string, tx *bolt.Tx) *bolt.Bucket {
	var bk, err = tx.CreateBucketIfNotExists(strutil.ToBytes(name))
	xerror.Panic(err, "create bucket error")
	return bk
}

func (t *Client) View(ctx context.Context, fn func(*bolt.Bucket), names ...string) error {
	name := consts.KeyDefault
	if len(names) > 0 {
		name = names[0]
	}

	var span = tracing.CreateChild(ctx, name)
	defer span.Finish()

	ext.DBType.Set(span, Name)

	var c, cancel = t.Resource.LoadObj()
	defer cancel.Release()

	return c.(*bolt.DB).View(func(tx *bolt.Tx) (err error) { return xerror.Try(func() { fn(t.bucket(name, tx)) }) })
}

func (t *Client) Update(ctx context.Context, fn func(*bolt.Bucket), names ...string) error {
	name := consts.KeyDefault
	if len(names) > 0 {
		name = names[0]
	}

	var span = tracing.CreateChild(ctx, name)
	defer span.Finish()

	ext.DBType.Set(span, Name)

	var c, cancel = t.Resource.LoadObj()
	defer cancel.Release()

	return c.(*bolt.DB).Update(func(tx *bolt.Tx) (err error) { return xerror.Try(func() { fn(t.bucket(name, tx)) }) })
}
