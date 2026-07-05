package tunnelgateway

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/tunnel"
)

// FiberHandler returns a Fiber handler for the gateway HTTP proxy
func (g *tunnelGateway) FiberHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		serviceName := c.Params("service")
		if serviceName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "service name required",
			})
		}

		fullPath := c.Path()
		subPath := ""
		if idx := strings.Index(fullPath, serviceName); idx >= 0 {
			subPath = fullPath[idx+len(serviceName):]
		}

		g.mu.RLock()
		svc, ok := g.services[serviceName]
		g.mu.RUnlock()

		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "service not found",
			})
		}

		if svc.session == nil || svc.session.IsClosed() {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "service unavailable",
			})
		}

		ctx := c.Context()
		stream, err := svc.session.Open(ctx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("failed to open stream: %v", err),
			})
		}
		defer func() {
			if err := stream.Close(); err != nil {
				log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: failed to close stream")
			}
		}()

		meta := tunnel.RequestMeta{
			ServiceID:    svc.info.ID,
			EndpointType: tunnel.EndpointTypeHTTP,
			Path:         subPath,
			Method:       c.Method(),
		}
		payload, err := json.Marshal(meta)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("failed to build request meta: %v", err),
			})
		}

		msg := &tunnel.Message{
			Type:    tunnel.MessageTypeHTTPRequest,
			Payload: payload,
		}

		if err := g.sendMessage(stream, msg); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("failed to send message: %v", err),
			})
		}

		req := c.Request()
		if _, err := stream.Write(req.Header.Header()); err != nil {
			return err
		}
		if _, err := stream.Write(req.Body()); err != nil {
			return err
		}

		buf := make([]byte, 32*1024)
		for {
			n, err := stream.Read(buf)
			if n > 0 {
				if _, writeErr := c.Response().BodyWriter().Write(buf[:n]); writeErr != nil {
					return writeErr
				}
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}

		return nil
	}
}
