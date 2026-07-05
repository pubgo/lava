package wsproxy

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsClosedConnError(t *testing.T) {
	t.Parallel()

	if !isClosedConnError(errors.New("use of closed network connection")) {
		t.Fatal("expected closed network connection to match")
	}
	if isClosedConnError(errors.New("other error")) {
		t.Fatal("unexpected match")
	}
}

func TestWebsocketProxyPassesNonUpgradeRequests(t *testing.T) {
	t.Parallel()

	const body = "plain-http"
	handler := WebsocketProxy(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte(body))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTeapot)
	}
	if got := rec.Body.String(); got != body {
		t.Fatalf("body = %q, want %q", got, body)
	}
}

func TestWebsocketProxyReadLimitDefaults(t *testing.T) {
	t.Parallel()

	h := WebsocketProxy(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	p := h.(*Proxy)
	if p.ReadLimit != maxMessageSize {
		t.Fatalf("ReadLimit = %d, want %d", p.ReadLimit, maxMessageSize)
	}
}

func TestWebsocketProxyWithReadLimit(t *testing.T) {
	t.Parallel()

	const limit = 128 * 1024
	h := WebsocketProxy(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), WithReadLimit(limit))
	p := h.(*Proxy)
	if p.ReadLimit != limit {
		t.Fatalf("ReadLimit = %d, want %d", p.ReadLimit, limit)
	}
}

func TestInMemoryResponseWriter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	w := newInMemoryResponseWriter(&buf)
	w.Header().Set("X-Test", "1")
	w.WriteHeader(http.StatusCreated)

	n, err := w.Write([]byte("line\n"))
	if err != nil || n != 5 {
		t.Fatalf("Write: n=%d err=%v", n, err)
	}
	if w.Header().Get("X-Test") != "1" {
		t.Fatal("header not preserved")
	}
	if w.code != http.StatusCreated {
		t.Fatalf("code = %d", w.code)
	}
	if buf.String() != "line\n" {
		t.Fatalf("buffer = %q", buf.String())
	}
}

func TestWebsocketProxyWithOptions(t *testing.T) {
	t.Parallel()

	h := WebsocketProxy(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		WithMethodParamOverride("verb"),
		WithTokenCookieName("auth"),
		WithPingPong(true),
	)
	p := h.(*Proxy)
	if p.methodOverrideParam != "verb" || p.tokenCookieName != "auth" || !p.enablePingPong {
		t.Fatalf("options not applied: %+v", p)
	}
}
