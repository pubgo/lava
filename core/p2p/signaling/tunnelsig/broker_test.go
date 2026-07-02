package tunnelsig_test

import (
	"context"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/p2p/ice"
	"github.com/pubgo/lava/v2/core/p2p/signaling"
	"github.com/pubgo/lava/v2/core/p2p/signaling/tunnelsig"
	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelagent"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelgateway"
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux"
)

const testToken = "p2p-test-token"

func startGateway(t *testing.T, ctx context.Context, addr string) tunnel.Gateway {
	t.Helper()
	gw := tunnelgateway.NewGateway(&tunnel.GatewayConfig{
		ListenAddr: addr,
		Transport:  tunnel.TransportYamux,
	})
	auth, err := tunnel.NewTokenAuthProvider(testToken)
	if err != nil {
		t.Fatal(err)
	}
	gw.SetAuthProvider(auth)
	if err := gw.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = gw.Stop(ctx) })
	return gw
}

func startAgentBroker(t *testing.T, ctx context.Context, gwAddr, peerID string) (*tunnelagent.Agent, *tunnelsig.Broker) {
	t.Helper()
	cfg := &tunnel.AgentConfig{
		GatewayAddr: gwAddr,
		Transport:   tunnel.TransportYamux,
		ServiceName: "svc-" + peerID,
		Metadata:    map[string]string{"auth_token": testToken},
	}
	agent := tunnelagent.New(cfg)
	if err := agent.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = agent.Stop(ctx) })

	broker := tunnelsig.New(agent.Session(), peerID, testToken)
	tunnelsig.AttachAgentHandler(cfg, broker)
	if err := broker.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = broker.Close() })
	return agent, broker
}

func TestBrokerSignalExchange(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	startGateway(t, ctx, "127.0.0.1:22000")
	_, brokerA := startAgentBroker(t, ctx, "127.0.0.1:22000", "node-a")
	_, brokerB := startAgentBroker(t, ctx, "127.0.0.1:22000", "node-b")

	time.Sleep(300 * time.Millisecond)

	if err := brokerA.Send(ctx, signaling.Message{
		Type:  signaling.TypeOffer,
		To:    "node-b",
		Ufrag: "uf-a",
		Pwd:   "pw-a",
	}); err != nil {
		t.Fatal(err)
	}

	recvCtx, recvCancel := context.WithTimeout(ctx, 5*time.Second)
	defer recvCancel()
	msg, err := brokerB.Recv(recvCtx, "node-b")
	if err != nil {
		t.Fatal(err)
	}
	if msg.Type != signaling.TypeOffer || msg.From != "node-a" || msg.Ufrag != "uf-a" {
		t.Fatalf("unexpected offer: %+v", msg)
	}

	if err := brokerB.Send(ctx, signaling.Message{
		Type:  signaling.TypeAnswer,
		To:    "node-a",
		Ufrag: "uf-b",
		Pwd:   "pw-b",
	}); err != nil {
		t.Fatal(err)
	}
	msg, err = brokerA.Recv(recvCtx, "node-a")
	if err != nil {
		t.Fatal(err)
	}
	if msg.Type != signaling.TypeAnswer || msg.From != "node-b" {
		t.Fatalf("unexpected answer: %+v", msg)
	}

	if err := brokerA.Send(ctx, signaling.Message{
		Type:      signaling.TypeCandidate,
		To:        "node-b",
		Candidate: "candidate:1 1 UDP 2130706431 127.0.0.1 9 typ host",
	}); err != nil {
		t.Fatal(err)
	}
	msg, err = brokerB.Recv(recvCtx, "node-b")
	if err != nil {
		t.Fatal(err)
	}
	if msg.Type != signaling.TypeCandidate || msg.Candidate == "" {
		t.Fatalf("unexpected candidate: %+v", msg)
	}
}

func TestBrokerICEConnect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	startGateway(t, ctx, "127.0.0.1:22001")
	_, brokerA := startAgentBroker(t, ctx, "127.0.0.1:22001", "a")
	_, brokerB := startAgentBroker(t, ctx, "127.0.0.1:22001", "b")

	time.Sleep(300 * time.Millisecond)

	cfg := ice.Config{ICETimeout: 10 * time.Second, AuthToken: testToken}
	errCh := make(chan error, 2)
	go func() {
		c, err := ice.Connect(ctx, cfg, brokerA, "a", "b", ice.RoleDialer)
		if err != nil {
			errCh <- err
			return
		}
		_ = c.ICE.Close()
		_ = c.Agent.Close()
		errCh <- nil
	}()
	go func() {
		c, err := ice.Connect(ctx, cfg, brokerB, "b", "a", ice.RoleListener)
		if err != nil {
			errCh <- err
			return
		}
		_ = c.ICE.Close()
		_ = c.Agent.Close()
		errCh <- nil
	}()
	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			t.Fatalf("ice via tunnel signaling: %v", err)
		}
	}
}
