package fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
)

type Builder struct {
	srv *fiber.App
}

func (t *Builder) Get() *fiber.App {
	if t.srv == nil {
		panic("please init fiber")
	}

	return t.srv
}

func (t *Builder) Build(cfg Cfg) (err error) {
	defer xerror.RespErr(&err)

	var fc = fiber.New().Config()
	xerror.Panic(merge.CopyStruct(&fc, &cfg))

	if cfg.Templates.Dir != "" && cfg.Templates.Ext != "" {
		fc.Views = html.New(cfg.Templates.Dir, cfg.Templates.Ext)
	}

	t.srv = fiber.New(fc)
	return nil
}

func New() Builder { return Builder{} }
