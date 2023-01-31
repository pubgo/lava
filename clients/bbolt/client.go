package bbolt

import (
	"context"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/merge"
	"github.com/pubgo/funk/result"
	bolt "go.etcd.io/bbolt"

	"github.com/pubgo/lava/core/tracing"
	"github.com/pubgo/lava/logging/logutil"
	utils2 "github.com/pubgo/lava/pkg/utils"
)

func New(cfg *Config, log *log.Logger) *Client {
	cfg = merge.Copy(DefaultConfig(), cfg).Unwrap()
	assert.MustF(cfg.Build(), "build failed, cfg=%#v", cfg)

	return &Client{DB: cfg.Get(), log: log}
}

type Client struct {
	*bolt.DB
	log *log.Logger
}

func (t *Client) bucket(name string, tx *bolt.Tx) *bolt.Bucket {
	var _, err = tx.CreateBucketIfNotExists([]byte(name))
	logutil.ErrRecord(t.log, err)
	return tx.Bucket([]byte(name))
}

func (t *Client) Set(ctx context.Context, key string, val []byte, names ...string) error {
	return t.Update(ctx, func(bucket *bolt.Bucket) error {
		return bucket.Put(utils2.StoB(key), val)
	}, names...)
}

func (t *Client) Get(ctx context.Context, key string, names ...string) result.Result[[]byte] {
	var (
		val []byte
		err = t.View(ctx, func(bucket *bolt.Bucket) error {
			val = bucket.Get(utils2.StoB(key))
			return nil
		}, names...)
	)
	return result.Wrap(val, err)
}

func (t *Client) List(ctx context.Context, fn func(k, v []byte) error, names ...string) error {
	return t.View(ctx, func(bucket *bolt.Bucket) error { return bucket.ForEach(fn) }, names...)
}

func (t *Client) Delete(ctx context.Context, key string, names ...string) error {
	return t.Update(ctx, func(bucket *bolt.Bucket) error {
		return bucket.Delete(utils2.StoB(key))
	}, names...)
}

func (t *Client) View(ctx context.Context, fn func(*bolt.Bucket) error, names ...string) error {
	name := utils2.GetDefault(names...)

	var span = tracing.CreateChild(ctx, name)
	defer span.Finish()
	ext.DBType.Set(span, Name)

	return t.DB.View(func(tx *bolt.Tx) error {
		return fn(t.bucket(name, tx))
	})
}

func (t *Client) Update(ctx context.Context, fn func(*bolt.Bucket) error, names ...string) error {
	name := utils2.GetDefault(names...)

	var span = tracing.CreateChild(ctx, name)
	defer span.Finish()
	ext.DBType.Set(span, Name)

	return t.DB.Update(func(tx *bolt.Tx) (err error) {
		return fn(t.bucket(name, tx))
	})
}
