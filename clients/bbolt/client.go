package bbolt

import (
	"context"
	"io"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	bolt "go.etcd.io/bbolt"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/plugins/tracing"
	"github.com/pubgo/lava/resource"
)

func Get(name ...string) *Client {
	var val = resource.Get(Name, lavax.GetDefault(name...))
	if val != nil {
		return val.(*Client)
	}
	return nil
}

var _ resource.Resource = (*Client)(nil)

type Client struct {
	v *bolt.DB
}

func (t *Client) Unwrap() io.Closer               { return t.Get() }
func (t *Client) Kind() string                    { return Name }
func (t *Client) UpdateObj(val resource.Resource) { t.v = val.(*Client).v }
func (t *Client) Get() *bolt.DB                   { return t.v }

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

	return t.Get().View(func(tx *bolt.Tx) (err error) { return xerror.Try(func() { fn(t.bucket(name, tx)) }) })
}

func (t *Client) Update(ctx context.Context, fn func(*bolt.Bucket), names ...string) error {
	name := consts.KeyDefault
	if len(names) > 0 {
		name = names[0]
	}

	var span = tracing.CreateChild(ctx, name)
	defer span.Finish()

	ext.DBType.Set(span, Name)

	return t.Get().Update(func(tx *bolt.Tx) (err error) { return xerror.Try(func() { fn(t.bucket(name, tx)) }) })
}
