package zrpc

import (
	"testing"

	"github.com/nats-io/nats.go"
)

func TestIncomingMetadata(t *testing.T) {
	header := nats.Header{}
	header.Add("X-Request-Id", "req-1")
	header.Add("X-Auth-Token", "tok")
	header.Add(HeaderTimeout, "5s")
	header.Add(HeaderStream, "1")
	header.Add(HeaderStreamFrame, "open")
	header.Add(HeaderStreamReqSub, "_INBOX.abc")
	header.Add("Nats-Msg-Id", "dedup-1")

	md := incomingMetadata(header)

	if got := md.Get("x-request-id"); len(got) != 1 || got[0] != "req-1" {
		t.Fatalf("unexpected x-request-id: %v", got)
	}
	if got := md.Get("x-auth-token"); len(got) != 1 || got[0] != "tok" {
		t.Fatalf("unexpected x-auth-token: %v", got)
	}

	for _, key := range []string{
		"timeout",
		"zrpc-stream",
		"zrpc-stream-frame",
		"zrpc-stream-req-subject",
		"nats-msg-id",
	} {
		if vals := md[key]; len(vals) != 0 {
			t.Fatalf("expected transport header %q to be filtered, got %v", key, vals)
		}
	}
}

func TestIncomingMetadataEmpty(t *testing.T) {
	if md := incomingMetadata(nil); md != nil {
		t.Fatalf("expected nil metadata for empty header, got %v", md)
	}
}
