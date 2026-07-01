// P2P 示例：本地内存信令 / tunnel gateway 全链路 / dev coturn 验证。
//
// 用法:
//
//	# 1) 本地快速验证（无需 gateway，内存信令）
//	go run ./core/p2p/example/main.go -mode=local
//
//	# 2) 单进程全链路（gateway + 双 agent + P2P 拨号 + stream）
//	go run ./core/p2p/example/main.go -mode=all
//
//	# 3) 分终端运行（模拟两台机器）
//	go run ./core/p2p/example/main.go -mode=gateway
//	go run ./core/p2p/example/main.go -mode=peer -peer-id=node-b
//	go run ./core/p2p/example/main.go -mode=peer -peer-id=node-a -dial-to=node-b
//
//	# 4) 经 dev coturn STUN/TURN 验证（需外网）
//	go run ./core/p2p/example/main.go -mode=dev
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pubgo/lava/v2/core/p2p"
	"github.com/pubgo/lava/v2/core/p2p/p2pbuilder"
	"github.com/pubgo/lava/v2/core/p2p/signaling"
	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelgateway"
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux"
)

const defaultAuthToken = "p2p-example-token"

var (
	mode        = flag.String("mode", "local", "Run mode: local, all, gateway, peer, dev")
	gatewayAddr = flag.String("gateway-addr", "127.0.0.1:27000", "Tunnel gateway address")
	peerID      = flag.String("peer-id", "", "P2P peer ID (required for mode=peer)")
	dialTo      = flag.String("dial-to", "", "Dial remote peer ID after listen (mode=peer)")
	authToken   = flag.String("auth-token", defaultAuthToken, "Tunnel/P2P auth token")
	insecure    = flag.Bool("insecure", true, "Skip QUIC TLS verify (dev only)")
)

func main() {
	flag.Parse()
	ctx, cancel := signalContext()
	defer cancel()

	switch *mode {
	case "local":
		if err := runLocal(ctx); err != nil {
			log.Fatal(err)
		}
	case "all":
		if err := runAll(ctx); err != nil {
			log.Fatal(err)
		}
	case "gateway":
		runGateway(ctx)
	case "peer":
		if *peerID == "" {
			log.Fatal("-peer-id is required for mode=peer")
		}
		runPeer(ctx, *peerID, *dialTo)
	case "dev":
		if err := runDev(ctx); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown mode %q", *mode)
	}
}

func signalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()
	return ctx, cancel
}

func p2pCfg() p2p.Config {
	cfg := p2p.ConfigFromEnv()
	cfg.Insecure = *insecure
	if cfg.AuthToken == "" {
		cfg.AuthToken = *authToken
	}
	if cfg.ICETimeout <= 0 {
		cfg.ICETimeout = 30 * time.Second
	}
	return cfg
}

// runLocal 内存信令 + Coordinator 全链路（最快验证）。
func runLocal(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cfg := p2p.Config{ICETimeout: 15 * time.Second, Insecure: true}
	brokerA, brokerB := signaling.Pair()
	coordA := p2p.NewCoordinator(cfg, brokerA, "node-a")
	coordB := p2p.NewCoordinator(cfg, brokerB, "node-b")
	defer func() { _ = coordA.Close() }()
	defer func() { _ = coordB.Close() }()

	ln, err := coordB.Listen(ctx, "node-b")
	if err != nil {
		return err
	}

	dialCh := make(chan error, 1)
	var peerA p2p.PeerConn
	go func() {
		var err error
		peerA, err = coordA.Dial(ctx, "node-b")
		dialCh <- err
	}()

	peerB, err := ln.Accept(ctx)
	if err != nil {
		return err
	}
	if err := <-dialCh; err != nil {
		return err
	}
	defer func() { _ = peerA.Close() }()
	defer func() { _ = peerB.Close() }()

	msg, err := exchangeStream(ctx, peerA, peerB, []byte("hello-local-p2p"))
	if err != nil {
		return err
	}

	printResult("local", peerA, msg)
	return nil
}

// runAll 单进程 gateway + 双 agent + P2P。
func runAll(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	gwAddr := *gatewayAddr
	gw := tunnelgateway.NewGateway(&tunnel.GatewayConfig{
		ListenAddr: gwAddr,
		Transport:  tunnel.TransportYamux,
	})
	if err := tunnel.ConfigureGatewayAuth(gw, *authToken); err != nil {
		return err
	}
	if err := gw.Start(ctx); err != nil {
		return err
	}
	defer func() { _ = gw.Stop(context.Background()) }()

	coordA, stopA, err := startPeerStack(ctx, gwAddr, "node-a", true)
	if err != nil {
		return err
	}
	defer stopA()

	coordB, stopB, err := startPeerStack(ctx, gwAddr, "node-b", true)
	if err != nil {
		return err
	}
	defer stopB()

	time.Sleep(300 * time.Millisecond)

	ln, err := coordB.Listen(ctx, "node-b")
	if err != nil {
		return err
	}

	dialCh := make(chan error, 1)
	var peerA p2p.PeerConn
	go func() {
		var err error
		peerA, err = coordA.Dial(ctx, "node-b")
		dialCh <- err
	}()

	peerB, err := ln.Accept(ctx)
	if err != nil {
		return err
	}
	if err := <-dialCh; err != nil {
		return err
	}
	defer func() { _ = peerA.Close() }()
	defer func() { _ = peerB.Close() }()

	msg, err := exchangeStream(ctx, peerA, peerB, []byte("hello-tunnel-p2p"))
	if err != nil {
		return err
	}

	printResult("all", peerA, msg)
	fmt.Printf("gateway: %s\n", gwAddr)
	return nil
}

// runDev 经 dev coturn STUN/TURN 验证（需外网）。
func runDev(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cfg := p2p.DefaultConfig()
	cfg.Insecure = true
	cfg.ICETimeout = 45 * time.Second

	fmt.Println("Using dev coturn:")
	fmt.Printf("  STUN: %v\n", cfg.STUNURLs)
	fmt.Printf("  TURN: %s (user=%s)\n", cfg.TURN.URL, cfg.TURN.Username)

	brokerA, brokerB := signaling.Pair()
	coordA := p2p.NewCoordinator(cfg, brokerA, "dev-a")
	coordB := p2p.NewCoordinator(cfg, brokerB, "dev-b")
	defer func() { _ = coordA.Close() }()
	defer func() { _ = coordB.Close() }()

	ln, err := coordB.Listen(ctx, "dev-b")
	if err != nil {
		return err
	}

	dialCh := make(chan error, 1)
	var peerA p2p.PeerConn
	go func() {
		var err error
		peerA, err = coordA.Dial(ctx, "dev-b")
		dialCh <- err
	}()

	peerB, err := ln.Accept(ctx)
	if err != nil {
		return err
	}
	if err := <-dialCh; err != nil {
		return err
	}
	defer func() { _ = peerA.Close() }()
	defer func() { _ = peerB.Close() }()

	msg, err := exchangeStream(ctx, peerA, peerB, []byte("hello-dev-coturn"))
	if err != nil {
		return err
	}

	printResult("dev", peerA, msg)
	printResult("dev", peerB, msg)
	return nil
}

func runGateway(ctx context.Context) {
	gw := tunnelgateway.NewGateway(&tunnel.GatewayConfig{
		ListenAddr: *gatewayAddr,
		Transport:  tunnel.TransportYamux,
	})
	if err := tunnel.ConfigureGatewayAuth(gw, *authToken); err != nil {
		log.Fatal(err)
	}
	if err := gw.Start(ctx); err != nil {
		log.Fatal(err)
	}
	defer func() { _ = gw.Stop(context.Background()) }()

	fmt.Printf("Gateway listening on %s (auth enabled)\n", *gatewayAddr)
	fmt.Println("Start peers in other terminals:")
	fmt.Printf("  go run ./core/p2p/example/main.go -mode=peer -peer-id=node-b -gateway-addr=%s\n", *gatewayAddr)
	fmt.Printf("  go run ./core/p2p/example/main.go -mode=peer -peer-id=node-a -dial-to=node-b -gateway-addr=%s\n", *gatewayAddr)
	<-ctx.Done()
}

func runPeer(ctx context.Context, id, remote string) {
	// 主动拨号方只拨号，不再注册被动监听，避免同一节点多 ICE agent 经 Multiplex 串台。
	listen := remote == ""
	coord, stop, err := startPeerStack(ctx, *gatewayAddr, id, listen)
	if err != nil {
		log.Fatal(err)
	}
	defer stop()

	fmt.Printf("P2P peer %q ready (gateway=%s)\n", id, *gatewayAddr)
	fmt.Printf("Stats: %+v\n", coord.Stats())

	if remote == "" {
		fmt.Println("Listening for inbound P2P... Press Ctrl+C to stop")
		<-ctx.Done()
		return
	}

	dialCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	fmt.Printf("Dialing %q ...\n", remote)
	peer, err := coord.Dial(dialCtx, remote)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = peer.Close() }()

	st, err := peer.Open(dialCtx)
	if err != nil {
		log.Fatal(err)
	}
	payload := []byte(fmt.Sprintf("ping from %s at %s", id, time.Now().Format(time.RFC3339)))
	if _, err := st.Write(payload); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent: %q\n", payload)
	printResult("peer", peer, string(payload))
}

func startPeerStack(ctx context.Context, gwAddr, id string, listen bool) (p2p.Coordinator, func(), error) {
	cfg := p2pCfg()
	node, err := p2pbuilder.Standalone(ctx, p2pbuilder.StandaloneOptions{
		GatewayAddr:   gwAddr,
		PeerID:        id,
		AuthToken:     *authToken,
		ListenOnStart: listen,
		Config:        &cfg,
	})
	if err != nil {
		return nil, nil, err
	}
	return node.Coordinator, func() { _ = node.Close() }, nil
}

func exchangeStream(ctx context.Context, a, b p2p.PeerConn, payload []byte) (string, error) {
	readCh := make(chan string, 1)
	go func() {
		st, err := b.Accept()
		if err != nil {
			readCh <- err.Error()
			return
		}
		buf := make([]byte, 256)
		n, err := st.Read(buf)
		if err != nil && err != io.EOF {
			readCh <- err.Error()
			return
		}
		readCh <- string(buf[:n])
	}()

	st, err := a.Open(ctx)
	if err != nil {
		return "", err
	}
	if _, err := st.Write(payload); err != nil {
		return "", err
	}
	got := <-readCh
	if got != string(payload) {
		return got, fmt.Errorf("payload mismatch: got %q want %q", got, payload)
	}
	return got, nil
}

func printResult(label string, pc p2p.PeerConn, msg string) {
	pair := pc.SelectedPair()
	fmt.Println("----------------------------------------")
	fmt.Printf("[%s] P2P OK\n", label)
	fmt.Printf("  local peer:  %s\n", pc.LocalPeerID())
	fmt.Printf("  remote peer: %s\n", pc.RemotePeerID())
	fmt.Printf("  message:     %q\n", msg)
	fmt.Printf("  ICE local:   %s (%s)\n", pair.LocalType, pair.LocalAddr)
	fmt.Printf("  ICE remote:  %s (%s)\n", pair.RemoteType, pair.RemoteAddr)
	fmt.Println("----------------------------------------")
}
