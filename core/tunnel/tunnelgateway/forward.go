package tunnelgateway

import (
	"context"
	"encoding/json"
	"io"
	"net"

	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func (g *tunnelGateway) Forward(ctx context.Context, serviceName string, endpointType tunnel.EndpointType, conn net.Conn) error {
	g.mu.RLock()
	svc, ok := g.services[serviceName]
	g.mu.RUnlock()

	if !ok {
		return tunnel.ErrServiceNotFound
	}

	if svc.session == nil || svc.session.IsClosed() {
		return tunnel.ErrSessionClosed
	}

	stream, err := svc.session.Open(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: failed to close stream")
		}
	}()

	meta := tunnel.RequestMeta{
		ServiceID:    svc.info.ID,
		EndpointType: endpointType,
	}
	payload, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	msg := &tunnel.Message{
		Type:    g.endpointTypeToMessageType(endpointType),
		Payload: payload,
	}

	if err := g.sendMessage(stream, msg); err != nil {
		return err
	}

	errCh := make(chan error, 2)
	go func() {
		_, err := io.Copy(stream, conn)
		errCh <- err
	}()
	go func() {
		_, err := io.Copy(conn, stream)
		errCh <- err
	}()

	<-errCh
	return nil
}

func (g *tunnelGateway) endpointTypeToMessageType(et tunnel.EndpointType) tunnel.MessageType {
	switch et {
	case tunnel.EndpointTypeHTTP:
		return tunnel.MessageTypeHTTPRequest
	case tunnel.EndpointTypeGRPC:
		return tunnel.MessageTypeGRPCRequest
	case tunnel.EndpointTypeDebug:
		return tunnel.MessageTypeDebugRequest
	default:
		return tunnel.MessageTypeHTTPRequest
	}
}

func (g *tunnelGateway) sendMessage(stream tunnel.Stream, msg *tunnel.Message) error {
	return tunnel.WriteMessage(stream, msg)
}
