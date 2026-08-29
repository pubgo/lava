package netutil

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRemoteIp(t *testing.T) {
	tests := []struct {
		name string
		req  *http.Request
		want string
	}{
		{
			name: "x_real_ip",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.RemoteAddr = "10.0.0.1:1234"
				r.Header.Set(XRealIP, "203.0.113.1")
				return r
			}(),
			want: "203.0.113.1",
		},
		{
			name: "x_forwarded_for",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.RemoteAddr = "10.0.0.1:1234"
				r.Header.Set(XForwardedFor, "198.51.100.2")
				return r
			}(),
			want: "198.51.100.2",
		},
		{
			name: "remote_addr",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.RemoteAddr = "192.0.2.10:9999"
				return r
			}(),
			want: "192.0.2.10",
		},
		{
			name: "ipv6_loopback",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.RemoteAddr = "[::1]:80"
				r.Header.Set(XRealIP, "::1")
				return r
			}(),
			want: "127.0.0.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoteIp(tt.req); got != tt.want {
				t.Fatalf("got=%q want=%q", got, tt.want)
			}
		})
	}
}

func TestGetIP(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Real-IP", "203.0.113.9")
	ip, err := GetIP(r)
	if err != nil || ip != "203.0.113.9" {
		t.Fatalf("GetIP=%q err=%v", ip, err)
	}
}

func TestIsErrServerClosed(t *testing.T) {
	if !IsErrServerClosed(nil) {
		t.Fatal("nil should be treated as closed")
	}
	if !IsErrServerClosed(http.ErrServerClosed) {
		t.Fatal("http.ErrServerClosed")
	}
	if !IsErrServerClosed(net.ErrClosed) {
		t.Fatal("net.ErrClosed")
	}
	if !IsErrServerClosed(context.Canceled) {
		t.Fatal("context.Canceled")
	}
	if IsErrServerClosed(errors.New("boom")) {
		t.Fatal("generic error should not match")
	}
	if err := SkipServerClosedError(http.ErrServerClosed); err != nil {
		t.Fatalf("SkipServerClosedError=%v", err)
	}
}

func TestGetLocalIP(t *testing.T) {
	ip := GetLocalIP()
	if ip == "" {
		t.Fatal("empty local ip")
	}
}
