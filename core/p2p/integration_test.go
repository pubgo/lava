package p2p_test

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/p2p"
	"github.com/pubgo/lava/v2/core/p2p/signaling"
)

func TestClientHTTPIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	cfg := p2p.Config{ICETimeout: 15 * time.Second, Insecure: true}
	brokerA, brokerB := signaling.Pair()
	coordA := p2p.NewCoordinator(cfg, brokerA, "node-a")
	coordB := p2p.NewCoordinator(cfg, brokerB, "node-b")
	defer func() { _ = coordA.Close() }()
	defer func() { _ = coordB.Close() }()

	ln, err := coordB.Listen(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		pc, err := ln.Accept(ctx)
		if err != nil {
			return
		}
		defer func() { _ = pc.Close() }()
		for {
			st, err := pc.Accept()
			if err != nil {
				return
			}
			go func(s io.ReadWriteCloser) {
				defer func() { _ = s.Close() }()
				_, _ = io.WriteString(s, "HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok")
			}(st)
		}
	}()

	client := p2p.NewClient(coordA)
	defer func() { _ = client.Close() }()

	resp, err := client.HTTPClient().Get("http://node-b/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Fatalf("body=%q", body)
	}
}

func TestClientRawStreamIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := p2p.Config{ICETimeout: 15 * time.Second, Insecure: true}
	brokerA, brokerB := signaling.Pair()
	coordA := p2p.NewCoordinator(cfg, brokerA, "node-a")
	coordB := p2p.NewCoordinator(cfg, brokerB, "node-b")
	defer func() { _ = coordA.Close() }()
	defer func() { _ = coordB.Close() }()

	ln, err := coordB.Listen(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}

	readCh := make(chan string, 1)
	go func() {
		pc, err := ln.Accept(ctx)
		if err != nil {
			readCh <- err.Error()
			return
		}
		defer func() { _ = pc.Close() }()
		st, err := pc.Accept()
		if err != nil {
			readCh <- err.Error()
			return
		}
		defer func() { _ = st.Close() }()
		line, err := bufio.NewReader(st).ReadString('\n')
		if err != nil {
			readCh <- err.Error()
			return
		}
		readCh <- strings.TrimSpace(line)
	}()

	client := p2p.NewClient(coordA)
	defer func() { _ = client.Close() }()

	conn, err := client.OpenStream(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := io.WriteString(conn, "ping\n"); err != nil {
		t.Fatal(err)
	}
	if got := <-readCh; got != "ping" {
		t.Fatalf("got %q", got)
	}
}
