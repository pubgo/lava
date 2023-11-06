package httprouter

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/lava"
	"github.com/pubgo/opendoc/opendoc"
	"net/http"
)

type Handler[Req any, Rsp any] func(ctx *fiber.Ctx, req *Req) (rsp *Rsp, err error)

func Get[Req any, Rsp any](r *lava.Router, path string, w Handler[Req, Rsp], opts ...func(op *opendoc.Operation)) {
	r.R.Get(path, WrapHandler(w))
	r.Doc.GetOf(func(op *opendoc.Operation) {
		op.SetModel(new(Req), new(Rsp))
		op.SetPath(path)
		for i := range opts {
			opts[i](op)
		}
	})
}

func Post[Req any, Rsp any](r *lava.Router, path string, w Handler[Req, Rsp], opts ...func(op *opendoc.Operation)) {
	r.R.Post(path, WrapHandler(w))
	r.Doc.PostOf(func(op *opendoc.Operation) {
		op.SetModel(new(Req), new(Rsp))
		op.SetPath(path)
		for i := range opts {
			opts[i](op)
		}
	})
}

func Put[Req any, Rsp any](r *lava.Router, path string, w Handler[Req, Rsp], opts ...func(op *opendoc.Operation)) {
	r.R.Put(path, WrapHandler(w))
	r.Doc.PutOf(func(op *opendoc.Operation) {
		op.SetModel(new(Req), new(Rsp))
		op.SetPath(path)
		for i := range opts {
			opts[i](op)
		}
	})
}

func Delete[Req any, Rsp any](r *lava.Router, path string, w Handler[Req, Rsp], opts ...func(op *opendoc.Operation)) {
	r.R.Delete(path, WrapHandler(w))
	r.Doc.DeleteOf(func(op *opendoc.Operation) {
		op.SetModel(new(Req), new(Rsp))
		op.SetPath(path)
		for i := range opts {
			opts[i](op)
		}
	})
}

func Patch[Req any, Rsp any](r *lava.Router, path string, w Handler[Req, Rsp], opts ...func(op *opendoc.Operation)) {
	r.R.Patch(path, WrapHandler(w))
	r.Doc.PatchOf(func(op *opendoc.Operation) {
		op.SetModel(new(Req), new(Rsp))
		op.SetPath(path)
		for i := range opts {
			opts[i](op)
		}
	})
}

var validate = validator.New()

func WrapHandler[Req any, Rsp any](handle func(ctx *fiber.Ctx, req *Req) (rsp *Rsp, err error)) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		var req Req

		if err := ctx.ParamsParser(&req); err != nil {
			return fmt.Errorf("failed to parse params, err:%w", err)
		}

		if err := ctx.QueryParser(&req); err != nil {
			return fmt.Errorf("failed to parse query, err:%w", err)
		}

		if err := ctx.ReqHeaderParser(&req); err != nil {
			return fmt.Errorf("failed to parse req header, err:%w", err)
		}

		switch ctx.Method() {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			if err := ctx.BodyParser(&req); err != nil {
				return fmt.Errorf("failed to parse body, err:%w", err)
			}
		}

		if err := validate.Struct(&req); err != nil {
			return fmt.Errorf("failed to validate request, err:%w", err)
		}

		rsp, err := handle(ctx, &req)
		if err != nil {
			return err
		}

		if rsp == nil {
			return ctx.JSON(make(map[string]any))
		}

		return ctx.JSON(rsp)
	}
}
