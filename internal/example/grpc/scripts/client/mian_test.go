package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pubgo/funk/assert"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func _TestName(t1 *testing.T) {
	t := &http2.Transport{
		AllowHTTP: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}

	ddd := assert.Must1(proto.Marshal(new(emptypb.Empty)))
	client := http.Client{Transport: t, Timeout: 15 * time.Second}
	resp, err := client.Post(
		"http://localhost:8080/gid/grpc/lava.service.ErrorService/Codes",
		"application/grpc",
		bytes.NewReader(ddd),
	)
	if err != nil {
		fmt.Printf("Failed get: %s\r\n", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed reading response body: %s\r\n", err)
	}
	fmt.Printf("Client Got response %d: %s %s\r\n", resp.StatusCode, resp.Proto, string(body))
}

func _TestName2(t *testing.T) {
	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get("http://localhost:8080/gid/grpc/lava/err_codes")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed reading response body: %s\r\n", err)
	}

	fmt.Printf("Client Got response %d: %s %s\r\n", resp.StatusCode, resp.Proto, string(body))
}

func _TestName3(t *testing.T) {
	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(
		"http://localhost:8080/gid/grpc/lava.service.ErrorService/Codes",
		"application/json",
		strings.NewReader(`{}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed reading response body: %s\r\n", err)
	}

	fmt.Printf("Client Got response %d: %s %s\r\n", resp.StatusCode, resp.Proto, string(body))
}
