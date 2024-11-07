package wss

import (
	"log"

	_ "github.com/fasthttp/router"
	"github.com/fasthttp/websocket"
	fiber "github.com/gofiber/fiber/v3"
	"github.com/pubgo/lava/pkg/wsutil"
)

func init() {
	app := fiber.New()

	app.Use("/ws", func(c fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if wsutil.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:id", func(c fiber.Ctx) error {
		// c.Locals is added to the *websocket.Conn
		log.Println(c.Locals("allowed"))  // true
		log.Println(c.Params("id"))       // 123
		log.Println(c.Query("v"))         // 1.0
		log.Println(c.Cookies("session")) // ""

		return wsutil.New(c, func(conn *websocket.Conn) {
			// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
			var (
				mt  int
				msg []byte
				err error
			)
			for {
				if mt, msg, err = conn.ReadMessage(); err != nil {
					log.Println("read:", err)
					break
				}
				log.Printf("recv: %s", msg)

				if err = conn.WriteMessage(mt, msg); err != nil {
					log.Println("write:", err)
					break
				}
			}
		})
	})

	log.Fatal(app.Listen(":3000"))
	// Access the websocket server: ws://localhost:3000/ws/123?v=1.0
	// https://www.websocket.org/echo.html
}
