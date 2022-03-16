package ginEntry

import (
	"bytes"
	"context"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/errors"
	"github.com/pubgo/lava/types"
)

func handlerMiddle(middlewares []types.Middleware) func(c *gin.Context) {
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
		var data []byte
		var err error

		if c.Request.Body != nil {
			data, err = ioutil.ReadAll(c.Request.Body)
			xerror.Panic(err)
			c.Request.Body = ioutil.NopCloser(bytes.NewReader(data))
		}

		err = handler(c.Request.Context(),
			&httpRequest{
				data: data,
				ctx:  c,
				ct:   c.ContentType(),
			},
			func(_ types.Response) error { return nil })

		if err != nil {
			var e = errors.FromError(err)
			c.AbortWithError(e.HTTPStatus(), e)
		}
	}
}