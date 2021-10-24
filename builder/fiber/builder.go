package fiber

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/gofiber/websocket/v2"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
)

func New() Builder { return Builder{} }

type Builder struct {
	srv *fiber.App
}

func (t *Builder) Get() *fiber.App {
	if t.srv == nil {
		panic("please init fiber")
	}

	return t.srv
}

func (t *Builder) Build(cfg Cfg) (err error) {
	defer xerror.RespErr(&err)

	var fc = fiber.New().Config()
	xerror.Panic(merge.CopyStruct(&fc, &cfg))

	if cfg.Templates.Dir != "" && cfg.Templates.Ext != "" {
		fc.Views = html.New(cfg.Templates.Dir, cfg.Templates.Ext)
	}

	t.srv = fiber.New(fc)

	t.srv.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}

		// c.Locals is added to the *websocket.Conn
		log.Println(c.Locals("allowed"))  // true
		log.Println(c.Params("id"))       // 123
		log.Println(c.Query("v"))         // 1.0
		log.Println(c.Cookies("session")) // ""

		return fiber.ErrUpgradeRequired
	})

	t.srv.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		// c.Locals is added to the *websocket.Conn
		log.Println(c.Locals("allowed"))  // true
		log.Println(c.Params("id"))       // 123
		log.Println(c.Query("v"))         // 1.0
		log.Println(c.Cookies("session")) // ""

		var (
			mt  int
			msg []byte
			err error
		)

		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}

			log.Printf("recv: %s", msg)

			if err = c.WriteMessage(mt, msg); err != nil {
				log.Println("write:", err)
				break
			}
		}

	}))
	return nil
}
