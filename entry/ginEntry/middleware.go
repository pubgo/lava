package ginEntry

import (
	"bytes"
	"context"
	"github.com/pubgo/lava/service/service_type"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/errors"
)

func handlerMiddle(middlewares []service_type.Middleware) func(c *gin.Context) {
	var handler = func(ctx context.Context, req service_type.Request, rsp func(response service_type.Response) error) error {
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
			func(_ service_type.Response) error { return nil })

		if err != nil {
			var e = errors.FromError(err)
			c.AbortWithError(e.HTTPStatus(), e)
		}
	}
}
