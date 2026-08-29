package gateway

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"
)

func TestNewIncomingContext_BinHeaderDecode(t *testing.T) {
	valid := base64.StdEncoding.EncodeToString([]byte("hello"))
	h := http.Header{}
	h.Set("x-custom-bin", valid)
	h.Add("x-custom-bin", "!!!not-base64!!!")
	h.Set("x-plain", "ok")

	_, md := newIncomingContext(context.Background(), h)
	if got := md.Get("x-plain"); len(got) != 1 || got[0] != "ok" {
		t.Fatalf("plain=%v", got)
	}
	got := md.Get("x-custom-bin")
	if len(got) != 1 || got[0] != "hello" {
		t.Fatalf("bin header=%v, want only decoded value", got)
	}
}
