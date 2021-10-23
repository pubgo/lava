package ginEntry

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/pubgo/lava/types"
)

func (t *restEntry) handlerMiddle(middlewares []types.Middleware) func(c *gin.Context) {
	var handler = func(ctx context.Context, req types.Request, rsp func(response types.Response) error) error {
		var reqCtx = req.(*httpRequest)

		// 最后执行业务逻辑
		reqCtx.ctx.Next()

		return rsp(&httpResponse{header: reqCtx.Header(), ctx: reqCtx.ctx})
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return func(c *gin.Context) {
		handler(
			c.Request.Context(),
			&httpRequest{ctx: c, header: c.Request.Header},
			func(_ types.Response) error { return nil },
		)
	}
}
