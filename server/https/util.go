package https

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Handler[Request any, Response any] func(ctx context.Context, req Request) (rsp Response, err error)

func Wrap[Request any, Response any](h Handler[Request, Response]) fiber.Handler {
	var hasReqBody bool
	var hasReqHeader bool
	var hasReqQuery bool
	var hasReqParams bool

	var rr = reflect.TypeOf(new(Request))
	for rr.Kind() == reflect.Ptr {
		rr = rr.Elem()
	}

	for i := 0; i < rr.NumField(); i++ {
		if rr.Field(i).Tag.Get("json") != "" {
			hasReqBody = true
		}

		if rr.Field(i).Tag.Get("xml") != "" {
			hasReqBody = true
		}

		if rr.Field(i).Tag.Get("form") != "" {
			hasReqBody = true
		}

		if rr.Field(i).Tag.Get("query") != "" {
			hasReqQuery = true
		}

		if rr.Field(i).Tag.Get("reqHeader") != "" {
			hasReqHeader = true
		}

		if rr.Field(i).Tag.Get("params") != "" {
			hasReqParams = true
		}
	}

	var mthName = getFunctionName(h)
	var reqName = reflect.TypeOf(new(Request)).String()
	var rspName = reflect.TypeOf(new(Response)).String()
	return func(ctx *fiber.Ctx) error {
		var req Request

		if hasReqBody {
			if err1 := ctx.BodyParser(&req); err1 != nil {
				return fmt.Errorf("method %s parse request %s body failed, err=%w", mthName, reqName, err1)
			}
		}

		if hasReqHeader {
			if err1 := ctx.ReqHeaderParser(&req); err1 != nil {
				return fmt.Errorf("method %s parse request %s header failed, err=%w", mthName, reqName, err1)
			}
		}

		if hasReqQuery {
			if err1 := ctx.QueryParser(&req); err1 != nil {
				return fmt.Errorf("method %s parse request %s query failed, err=%w", mthName, reqName, err1)
			}
		}

		if hasReqParams {
			if err1 := ctx.ParamsParser(&req); err1 != nil {
				return fmt.Errorf("method %s parse request %s params failed, err=%w", mthName, reqName, err1)
			}
		}

		var rsp, err = h(ctx.Context(), req)
		if err != nil {
			return fmt.Errorf("method %s handler failed,req=%s rsp=%s err=%w", mthName, reqName, rspName, err)
		}

		if err1 := ctx.JSON(rsp); err1 != nil {
			return fmt.Errorf("method %s json marshal %v failed, err=%w", mthName, rsp, err1)
		} else {
			return nil
		}
	}
}

func getFunctionName(i interface{}) string {
	fn := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	var names = strings.Split(fn, "/")
	return names[len(names)-1]
}
