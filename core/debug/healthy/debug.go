package healthy

import (
	"net/http"
	"time"

	jjson "github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/try"

	"github.com/pubgo/lava/core/debug"
	"github.com/pubgo/lava/core/healthy"
)

func init() {
	debug.Get("/health", func(ctx fiber.Ctx) error {
		dt := make(map[string]*health)
		for _, name := range healthy.List() {
			h := &health{}
			h.Err = try.Try(func() error {
				defer func(s time.Time) { h.Cost = time.Since(s).String() }(time.Now())
				return healthy.Get(name)(ctx)
			})
			dt[name] = h
		}

		bts, err := jjson.Marshal(dt)
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			_, err = ctx.Write([]byte(err.Error()))
			return err
		}

		ctx.Response().Header.Set("content-type", "application/json")
		ctx.Status(http.StatusOK)
		_, err = ctx.Write(bts)
		return err
	})
}

type health struct {
	Cost string `json:"cost,omitempty"`
	Err  error  `json:"err,omitempty"`
	Msg  string `json:"err_msg,omitempty"`
}
