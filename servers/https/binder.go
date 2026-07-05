package https

import (
	"github.com/gofiber/fiber/v3/binder"
)

func init() {
	binder.SetParserDecoder(binder.ParserConfig{
		IgnoreUnknownKeys: true,
		ZeroEmpty:         true,
	})
}

// RegParser registers custom Fiber binder parsers for HTTP handlers.
func RegParser(parsers []binder.ParserType) {
	binder.SetParserDecoder(binder.ParserConfig{
		IgnoreUnknownKeys: true,
		ZeroEmpty:         true,
		ParserType:        parsers,
	})
}
