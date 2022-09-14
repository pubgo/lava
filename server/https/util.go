package https

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	_ "github.com/emicklei/go-restful-openapi/v2"
	_ "github.com/emicklei/go-restful/v3"
	"github.com/gofiber/fiber/v2"
	_ "github.com/santhosh-tekuri/jsonschema/v3"
	_ "github.com/swaggest/rest"
)

// ErrResponse is HTTP error response body.
type ErrResponse struct {
	Status struct {
		StatusText     string `json:"status,omitempty" description:"Status text."`
		AppCode        int    `json:"code,omitempty" description:"Application-specific error code."`
		ErrorText      string `json:"error,omitempty" description:"Error message."`
		err            error  // Original error.
		httpStatusCode int    // HTTP response status code.
	}
	Data     interface{}
	Paginate struct {
		Next     int
		Previous int
		Count    int
	}
	Context map[string]interface{} `json:"context,omitempty" description:"Application context."`
}

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

// Info exposes information about use case.
type Info struct {
	name           string
	title          string
	description    string
	tags           []string
	expectedErrors []error
	isDeprecated   bool
}

var (
	_ HasTags           = Info{}
	_ HasTitle          = Info{}
	_ HasName           = Info{}
	_ HasDescription    = Info{}
	_ HasIsDeprecated   = Info{}
	_ HasExpectedErrors = Info{}
)

// IsDeprecated implements HasIsDeprecated.
func (i Info) IsDeprecated() bool {
	return i.isDeprecated
}

// SetIsDeprecated sets status of deprecation.
func (i *Info) SetIsDeprecated(isDeprecated bool) {
	i.isDeprecated = isDeprecated
}

// ExpectedErrors implements HasExpectedErrors.
func (i Info) ExpectedErrors() []error {
	return i.expectedErrors
}

// SetExpectedErrors sets errors that are expected to cause use case failure.
func (i *Info) SetExpectedErrors(expectedErrors ...error) {
	i.expectedErrors = expectedErrors
}

// Tags implements HasTag.
func (i Info) Tags() []string {
	return i.tags
}

// SetTags sets tags of use cases group.
func (i *Info) SetTags(tags ...string) {
	i.tags = tags
}

// Description implements HasDescription.
func (i Info) Description() string {
	return i.description
}

// SetDescription sets use case description.
func (i *Info) SetDescription(description string) {
	i.description = description
}

// Title implements HasTitle.
func (i Info) Title() string {
	return i.title
}

// SetTitle sets use case title.
func (i *Info) SetTitle(title string) {
	i.title = title
}

// Name implements HasName.
func (i Info) Name() string {
	return i.name
}

// SetName sets use case title.
func (i *Info) SetName(name string) {
	i.name = name
}
