package htmlx

import (
	pongo "github.com/flosch/pongo2/v5"
	"github.com/gofiber/fiber/v2"
)

func Html(ctx *fiber.Ctx, data []byte) error {
	ctx.Response().Header.SetContentType(fiber.MIMETextHTMLCharsetUTF8)
	ctx.Response().SetBody(data)
	return nil
}

type Context = pongo.Context
