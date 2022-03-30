package htmlx

import (
	"bytes"
	"html/template"

	"github.com/gofiber/fiber/v2"
)

func Html(ctx *fiber.Ctx, temp *template.Template, data any) error {
	if data == nil {
		data = map[string]interface{}{}
	}

	var buf = bytes.NewBuffer(nil)
	if err := temp.Execute(buf, data); err != nil {
		return err
	}
	ctx.Response().Header.SetContentType(fiber.MIMETextHTMLCharsetUTF8)
	ctx.Response().SetBody(buf.Bytes())
	return nil
}
