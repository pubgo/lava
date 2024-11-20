package https

import "github.com/gofiber/fiber/v3/binder"

func init() {
	binder.SetParserDecoder(binder.ParserConfig{
		IgnoreUnknownKeys: true,
		ZeroEmpty:         true,
		ParserType:        parserTypes,
	})
}
