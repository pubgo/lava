// Package tunnelsig 通过 tunnel gateway 控制流交换 ICE 信令。
package tunnelsig

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/pubgo/lava/v2/core/p2p/signaling"
	"github.com/pubgo/lava/v2/core/tunnel"
)

// Broker 经 tunnel gateway 转发 P2P 信令。
type Broker struct {
	session   tunnel.Session
	selfID    string
	authToken string
	agentMode bool // true：入站信令由 agent P2PSignalHandler 注入，不占用 session.Accept

	inbound chan signaling.Message
	stopCh  chan struct{}
	wg      sync.WaitGroup
	once    sync.Once
}

// New 创建 Broker；调用 Start 完成 peer 注册。
// 与 tunnel agent 联用时请调用 AttachAgentHandler，Broker 将自动进入 agentMode。
func New(session tunnel.Session, selfID, authToken string) *Broker {
	return &Broker{
		session:   session,
		selfID:    selfID,
		authToken: authToken,
		inbound:   make(chan signaling.Message, 64),
		stopCh:    make(chan struct{}),
	}
}

// Start 向 gateway 注册 peerID；非 agentMode 时启动 session.Accept 接收循环。
func (b *Broker) Start(ctx context.Context) error {
	if err := b.register(ctx); err != nil {
		return err
	}
	if b.agentMode {
		return nil
	}
	b.wg.Add(1)
	go b.acceptLoop()
	return nil
}

// Deliver 供 tunnel agent 的 P2PSignalHandler 回调注入信令。
func (b *Broker) Deliver(payload []byte) error {
	var msg signaling.Message
	if err := json.Unmarshal(payload, &msg); err != nil {
		return err
	}
	return b.enqueue(msg)
}

func (b *Broker) Send(ctx context.Context, msg signaling.Message) error {
	if msg.From == "" {
		msg.From = b.selfID
	}
	if msg.AuthToken == "" {
		msg.AuthToken = b.authToken
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	stream, err := b.session.Open(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = stream.Close() }()

	return tunnel.WriteMessage(stream, &tunnel.Message{
		Type:    tunnel.MessageTypeP2PSignal,
		Payload: payload,
	})
}

func (b *Broker) Recv(ctx context.Context, selfID string) (signaling.Message, error) {
	if selfID != "" && selfID != b.selfID {
		return signaling.Message{}, context.Canceled
	}
	select {
	case msg := <-b.inbound:
		return msg, nil
	case <-ctx.Done():
		return signaling.Message{}, ctx.Err()
	case <-b.stopCh:
		return signaling.Message{}, context.Canceled
	}
}

func (b *Broker) Close() error {
	b.once.Do(func() { close(b.stopCh) })
	b.wg.Wait()
	return nil
}

func (b *Broker) register(ctx context.Context) error {
	payload, err := json.Marshal(tunnel.P2PRegisterPayload{
		PeerID:    b.selfID,
		AuthToken: b.authToken,
	})
	if err != nil {
		return err
	}

	stream, err := b.session.Open(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = stream.Close() }()

	return tunnel.WriteMessage(stream, &tunnel.Message{
		Type:    tunnel.MessageTypeP2PRegister,
		Payload: payload,
	})
}

func (b *Broker) acceptLoop() {
	defer b.wg.Done()

	for {
		select {
		case <-b.stopCh:
			return
		default:
		}

		if b.session == nil || b.session.IsClosed() {
			return
		}

		stream, err := b.session.Accept()
		if err != nil {
			select {
			case <-b.stopCh:
				return
			default:
			}
			if b.session.IsClosed() {
				return
			}
			continue
		}

		go b.handleInboundStream(stream)
	}
}

func (b *Broker) handleInboundStream(stream tunnel.Stream) {
	defer func() { _ = stream.Close() }()

	msg, err := tunnel.ReadMessage(stream)
	if err != nil {
		return
	}
	if msg.Type != tunnel.MessageTypeP2PSignal {
		return
	}

	var sig signaling.Message
	if err := json.Unmarshal(msg.Payload, &sig); err != nil {
		return
	}
	_ = b.enqueue(sig)
}

func (b *Broker) enqueue(msg signaling.Message) error {
	if msg.To != "" && msg.To != b.selfID {
		return nil
	}
	select {
	case b.inbound <- msg:
		return nil
	case <-b.stopCh:
		return context.Canceled
	default:
		select {
		case <-b.inbound:
		default:
		}
		b.inbound <- msg
		return nil
	}
}

// AttachAgentHandler 将 Broker 挂到 agent：入站信令经 agent acceptLoop 分发，避免与 gateway 争用 session.Accept。
func AttachAgentHandler(cfg *tunnel.AgentConfig, b *Broker) {
	if cfg == nil || b == nil {
		return
	}
	b.agentMode = true
	cfg.P2PSignalHandler = func(payload []byte) {
		_ = b.Deliver(payload)
	}
}
