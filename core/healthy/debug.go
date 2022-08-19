package healthy

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/x/jsonx"

	"github.com/pubgo/lava/debug"
	"github.com/pubgo/lava/internal/pkg/utils"
)

func init() {
	debug.Get("/health", func(ctx *fiber.Ctx) error {
		var dt = make(map[string]*health)
		assert.Must(healthList.Each(func(name string, r interface{}) {
			var h = &health{}
			var dur, err = utils.Cost(func() { assert.Must(r.(Handler)(ctx)) })
			h.Cost = dur.String()
			if err != nil {
				h.Msg = err.Error()
				h.Err = err
			}
			dt[name] = h
		}))

		var bts, err = jsonx.Marshal(dt)
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
