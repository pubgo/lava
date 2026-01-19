package httputil

import (
	"io"
	"net"
	"net/http"
	"strings"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func StripPrefix(prefix string, hh fiber.Handler) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		ctx.Request().Header.Set("Path-Prefix", prefix)
		ctx.Request().SetRequestURI(strings.TrimPrefix(string(ctx.Request().RequestURI()), prefix))
		return hh(ctx)
	}
}

func FastHandler(h fasthttp.RequestHandler) http.Handler {
	return handlerFunc(h)
}

func HTTPHandlerFunc(h http.HandlerFunc) fiber.Handler { return HTTPHandler(h) }

func HTTPHandler(h http.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		handler := fasthttpadaptor.NewFastHTTPHandler(h)
		handler(c.Context())
		return nil
	}
}

func handlerFunc(h fasthttp.RequestHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// New fasthttp request
		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req)

		// Convert net/http -> fasthttp request
		if r.Body != nil {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, utils.StatusMessage(fiber.StatusInternalServerError), fiber.StatusInternalServerError)
				return
			}
			req.Header.SetContentLength(len(body))
			_, _ = req.BodyWriter().Write(body)
		}

		req.Header.SetMethod(r.Method)
		req.SetRequestURI(r.RequestURI)
		req.SetHost(r.Host)
		for key, val := range r.Header {
			for _, v := range val {
				req.Header.Set(key, v)
			}
		}

		if _, _, err := net.SplitHostPort(r.RemoteAddr); err != nil && err.(*net.AddrError).Err == "missing port in address" {
			r.RemoteAddr = net.JoinHostPort(r.RemoteAddr, "80")
		}

		remoteAddr, err := net.ResolveTCPAddr("tcp", r.RemoteAddr)
		if err != nil {
			http.Error(w, utils.StatusMessage(fiber.StatusInternalServerError), fiber.StatusInternalServerError)
			return
		}

		var ctx fasthttp.RequestCtx
		ctx.Init(req, remoteAddr, nil)

		h(&ctx)

		// Convert fasthttp Ctx > net/http
		for k, v := range ctx.Response.Header.All() {
			w.Header().Add(string(k), string(v))
		}
		w.WriteHeader(ctx.Response.StatusCode())
		_, _ = w.Write(ctx.Response.Body())
	}
}
