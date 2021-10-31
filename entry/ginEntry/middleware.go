package ginEntry

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pubgo/lava/types"
)

func (t *ginEntry) handlerMiddle(middlewares []types.Middleware) func(c *gin.Context) {
	var handler = func(ctx context.Context, req types.Request, rsp func(response types.Response) error) error {
		var reqCtx = req.(*httpRequest)

		// 执行最后的gin handler
		reqCtx.ctx.Next()

		// response回调
		return rsp(&httpResponse{ctx: reqCtx.ctx})
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return func(c *gin.Context) {
		if err := handler(
			c.Request.Context(),
			&httpRequest{ctx: c},
			func(_ types.Response) error { return nil },
		); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "msg": err})
		}
	}
}
