package https

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/lava"
	"github.com/pubgo/opendoc/opendoc"
)

type Handler[Req any, Rsp any] func(ctx *fiber.Ctx, req *Req) (rsp *Rsp, err error)

func Get[Req any, Rsp any](r *lava.Router, path string, w Handler[Req, Rsp], opts ...func(op *opendoc.Operation)) {
	r.R.Get(path, WrapHandler(w))
	r.Doc.GetOf(func(op *opendoc.Operation) {
		op.SetModel(new(Req), new(Rsp))
		op.SetPath("", path)
		for i := range opts {
			opts[i](op)
		}
	})
}

func Post[Req any, Rsp any](r *lava.Router, path string, w Handler[Req, Rsp], opts ...func(op *opendoc.Operation)) {
	r.R.Post(path, WrapHandler(w))
	r.Doc.PostOf(func(op *opendoc.Operation) {
		op.SetModel(new(Req), new(Rsp))
		op.SetPath("", path)
		for i := range opts {
			opts[i](op)
		}
	})
}

func Put[Req any, Rsp any](r *lava.Router, path string, w Handler[Req, Rsp], opts ...func(op *opendoc.Operation)) {
	r.R.Put(path, WrapHandler(w))
	r.Doc.PutOf(func(op *opendoc.Operation) {
		op.SetModel(new(Req), new(Rsp))
		op.SetPath("", path)
		for i := range opts {
			opts[i](op)
		}
	})
}

func Delete[Req any, Rsp any](r *lava.Router, path string, w Handler[Req, Rsp], opts ...func(op *opendoc.Operation)) {
	r.R.Delete(path, WrapHandler(w))
	r.Doc.DeleteOf(func(op *opendoc.Operation) {
		op.SetModel(new(Req), new(Rsp))
		op.SetPath("", path)
		for i := range opts {
			opts[i](op)
		}
	})
}

func Patch[Req any, Rsp any](r *lava.Router, path string, w Handler[Req, Rsp], opts ...func(op *opendoc.Operation)) {
	r.R.Patch(path, WrapHandler(w))
	r.Doc.PatchOf(func(op *opendoc.Operation) {
		op.SetModel(new(Req), new(Rsp))
		op.SetPath("", path)
		for i := range opts {
			opts[i](op)
		}
	})
}
