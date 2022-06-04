package fiber_builder

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/xerror"
)

func New() Builder { return Builder{} }

type Builder struct {
	srv *fiber.App
}

func (t *Builder) Get() *fiber.App {
	if t.srv == nil {
		panic("please init fiber")
	}

	return t.srv
}

func (t *Builder) Build(cfg *Cfg) (err error) {
	defer xerror.RecoverErr(&err)

	var fc = fiber.New().Config()
	xerror.Panic(merge.Struct(&fc, &cfg))
	t.srv = fiber.New(fc)

	if cfg.Templates.Dir != "" && cfg.Templates.Ext != "" {
		fc.Views = html.New(cfg.Templates.Dir, cfg.Templates.Ext)
	}

	return nil
}
