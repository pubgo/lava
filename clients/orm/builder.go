package orm

import (
	"log"
	"os"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/merge"
	"github.com/pubgo/funk/recovery"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/pubgo/lava/logging"
)

func New(cfg *Cfg, log *logging.Logger) *Client {
	assert.If(cfg == nil, "config is nil")

	var builder = DefaultCfg()
	builder.log = log.Named(Name)
	builder = merge.Struct(builder, cfg).Unwrap(func(err error) error {
		return errors.WrapTags(err, errors.Tags{"cfg": cfg})
	})

	assert.Must(builder.Build())
	return &Client{DB: builder.Get()}
}

func TestDb() *gorm.DB {
	defer recovery.Exit()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"))
	assert.Must(err, "open sqlite db failed")

	db.Logger = logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: false,
		Colorful:                  true,
	})
	return db
}
