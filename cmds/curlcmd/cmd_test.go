package curlcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKVFlagSetAndString(t *testing.T) {
	t.Parallel()

	var f kvFlag
	if err := f.Set("a=1"); err != nil {
		t.Fatal(err)
	}
	if err := f.Set("b=2"); err != nil {
		t.Fatal(err)
	}
	if f.String() != "a=1,b=2" {
		t.Fatalf("String() = %q", f.String())
	}
	if f.Map()["a"] != "1" || f.Map()["b"] != "2" {
		t.Fatalf("Map() = %#v", f.Map())
	}
	if f.Type() != "kv" {
		t.Fatalf("Type() = %q", f.Type())
	}
}

func TestKVFlagSetInvalid(t *testing.T) {
	t.Parallel()

	var f kvFlag
	if err := f.Set("invalid"); err == nil {
		t.Fatal("expected error for invalid pair")
	}
	if err := f.Set("=value"); err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestNormalizeMethod(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"__post__": "POST",
		"get":      "GET",
		"Put":      "PUT",
	}
	for in, want := range tests {
		if got := normalizeMethod(in); got != want {
			t.Errorf("normalizeMethod(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestJoinPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		prefix, path, want string
	}{
		{"/api", "users", "/api/users"},
		{"api", "/users", "/api/users"},
		{"", "/users", "/users"},
		{"/", "users", "/users"},
	}
	for _, tt := range tests {
		if got := joinPath(tt.prefix, tt.path); got != tt.want {
			t.Errorf("joinPath(%q, %q) = %q, want %q", tt.prefix, tt.path, got, tt.want)
		}
	}
}

func TestApplyPathParams(t *testing.T) {
	t.Parallel()

	got, err := applyPathParams("/users/{id}/posts/{post_id}", map[string]string{
		"id":      "42",
		"post_id": "7",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "/users/42/posts/7" {
		t.Fatalf("got %q", got)
	}

	_, err = applyPathParams("/users/{id}", map[string]string{})
	if err == nil {
		t.Fatal("expected missing param error")
	}
}

func TestPrintRoutes(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	routes := []routeOperation{
		{Method: "__get__", Path: "/hello", Operation: "demo.Hello", Verb: "GET"},
	}
	if err := printRoutes(&buf, routes, "/api"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "GET") || !strings.Contains(out, "/api/hello") || !strings.Contains(out, "demo.Hello") {
		t.Fatalf("output = %q", out)
	}
}

func TestPrintHeadersSorted(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	h := http.Header{
		"B": []string{"2"},
		"A": []string{"1"},
	}
	if err := printHeaders(&buf, h); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "A: 1\nB: 2\n" {
		t.Fatalf("output = %q", buf.String())
	}
}

func TestBuildRequestBody(t *testing.T) {
	t.Parallel()

	t.Run("inline", func(t *testing.T) {
		rc, err := buildRequestBody(`{"a":1}`, "", false)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rc.Close() }()
		b, _ := io.ReadAll(rc)
		if string(b) != `{"a":1}` {
			t.Fatalf("body = %q", b)
		}
	})

	t.Run("file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "body.json")
		if err := os.WriteFile(path, []byte("file-body"), 0o600); err != nil {
			t.Fatal(err)
		}
		rc, err := buildRequestBody("", path, false)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rc.Close() }()
		b, _ := io.ReadAll(rc)
		if string(b) != "file-body" {
			t.Fatalf("body = %q", b)
		}
	})

	t.Run("empty", func(t *testing.T) {
		rc, err := buildRequestBody("", "", false)
		if err != nil {
			t.Fatal(err)
		}
		if rc != nil {
			t.Fatal("expected nil reader")
		}
	})
}

func TestFetchGatewayRoutes(t *testing.T) {
	t.Parallel()

	wantRoutes := []routeOperation{
		{Method: "POST", Path: "/echo", Operation: "demo.Echo", Verb: "POST"},
	}
	info, err := json.Marshal(gatewayInfo{Method: wantRoutes})
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/debug/vars/api/list":
			_ = json.NewEncoder(w).Encode([]gatewayVarInfo{{Name: "gateway-server-info", Value: "x"}})
		case "/debug/vars/api/get/gateway-server-info":
			_, _ = w.Write(info)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	routes, err := fetchGatewayRoutes(context.Background(), srv.Client(), srv.URL, "gateway-server-info")
	if err != nil {
		t.Fatal(err)
	}
	if len(routes) != 1 || routes[0].Operation != "demo.Echo" {
		t.Fatalf("routes = %+v", routes)
	}
}

func TestTokenSaveAndLoad(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := saveToken("secret-token"); err != nil {
		t.Fatal(err)
	}
	got, err := loadToken()
	if err != nil {
		t.Fatal(err)
	}
	if got != "secret-token" {
		t.Fatalf("token = %q", got)
	}
}

func TestNewCommandHasLoginChild(t *testing.T) {
	t.Parallel()

	cmd := New()
	if cmd.Use != "curl [flags] <operation|path>" {
		t.Fatalf("Use = %q", cmd.Use)
	}
	found := false
	for _, child := range cmd.Children {
		if child.Use == "login" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected login subcommand")
	}
}
