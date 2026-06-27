package p2pbuilder_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/lifecycle/lifecyclebuilder"
	"github.com/pubgo/lava/v2/core/p2p"
	"github.com/pubgo/lava/v2/core/p2p/p2pbuilder"
	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelagent"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelgateway"
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux"
)

const testToken = "p2p-builder-token"

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

func startAgent(t *testing.T, ctx context.Context, gwAddr, serviceName string) *tunnelagent.Agent {
	t.Helper()
	agent := tunnelagent.New(&tunnel.AgentConfig{
		GatewayAddr: gwAddr,
		Transport:   tunnel.TransportYamux,
		ServiceName: serviceName,
		Metadata:    map[string]string{"auth_token": testToken},
	})
	if err := agent.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = agent.Stop(ctx) })
	return agent
}

func runLifecycleHooks(t *testing.T, ctx context.Context, execs []lifecycle.Executor) {
	t.Helper()
	for _, e := range execs {
		if err := e.Exec(ctx); err != nil {
			t.Fatal(err)
		}
	}
}

func startP2PNode(
	t *testing.T,
	ctx context.Context,
	lc lifecyclebuilder.Provider,
	agent *tunnelagent.Agent,
	peerID string,
	listenOnStart bool,
) p2p.Coordinator {
	t.Helper()
	coord, err := p2pbuilder.New(p2pbuilder.Params{
		LC:            lc.Setter,
		Agent:         agent,
		PeerID:        peerID,
		ListenOnStart: listenOnStart,
		Config: &p2p.Config{
			ICETimeout: 15 * time.Second,
			Insecure:   true,
			AuthToken:  testToken,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	runLifecycleHooks(t, ctx, lc.Getter.GetAfterStarts())
	t.Cleanup(func() { runLifecycleHooks(t, ctx, lc.Getter.GetBeforeStops()) })
	return coord
}

func TestP2PEndToEndViaTunnel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startGateway(t, ctx, "127.0.0.1:22010")
	agentA := startAgent(t, ctx, "127.0.0.1:22010", "svc-a")
	agentB := startAgent(t, ctx, "127.0.0.1:22010", "svc-b")
	time.Sleep(300 * time.Millisecond)

	lcA := lifecyclebuilder.New(nil)
	lcB := lifecyclebuilder.New(nil)
	coordA := startP2PNode(t, ctx, lcA, agentA, "node-a", false)
	coordB := startP2PNode(t, ctx, lcB, agentB, "node-b", true)
	time.Sleep(100 * time.Millisecond)

	ln, err := coordB.Listen(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}

	dialCh := make(chan p2p.PeerConn, 1)
	errCh := make(chan error, 1)
	go func() {
		pc, err := coordA.Dial(ctx, "node-b")
		if err != nil {
			errCh <- err
			return
		}
		dialCh <- pc
	}()

	acceptCtx, acceptCancel := context.WithTimeout(ctx, 20*time.Second)
	defer acceptCancel()
	peerB, err := ln.Accept(acceptCtx)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-errCh:
		t.Fatal(err)
	case peerA := <-dialCh:
		defer peerA.Close()
		defer peerB.Close()

		if peerA.RemotePeerID() != "node-b" || peerB.RemotePeerID() != "node-a" {
			t.Fatalf("peer ids: a->%s b->%s", peerA.RemotePeerID(), peerB.RemotePeerID())
		}

		readCh := make(chan string, 1)
		go func() {
			st, err := peerB.Accept()
			if err != nil {
				readCh <- err.Error()
				return
			}
			buf := make([]byte, 32)
			n, err := st.Read(buf)
			if err != nil && err != io.EOF {
				readCh <- err.Error()
				return
			}
			readCh <- string(buf[:n])
		}()

		stA, err := peerA.Open(ctx)
		if err != nil {
			t.Fatal(err)
		}
		msg := []byte("tunnel-p2p-e2e")
		if _, err := stA.Write(msg); err != nil {
			t.Fatal(err)
		}
		if got := <-readCh; got != string(msg) {
			t.Fatalf("got %q", got)
		}

		st := coordA.Stats()
		if st.SelfID != "node-a" || len(st.Connections) != 1 {
			t.Fatalf("stats a: %+v", st)
		}
	default:
		t.Fatal("dial did not complete")
	}
}
