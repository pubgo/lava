package resty

import (
	"bytes"
	"net/url"
	"strings"
	"testing"
)

func TestFilterFlags(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"application/json", "application/json"},
		{"application/json; charset=utf-8", "application/json"},
		{"text/plain foo", "text/plain"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := FilterFlags(tt.in); got != tt.want {
			t.Fatalf("FilterFlags(%q)=%q want %q", tt.in, got, tt.want)
		}
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		in   any
		want string
	}{
		{"hi", "hi"},
		{true, "true"},
		{42, "42"},
		{int64(7), "7"},
		{uint(3), "3"},
		{struct{ A int }{1}, "{1}"},
	}
	for _, tt := range tests {
		if got := ToString(tt.in); got != tt.want {
			t.Fatalf("ToString(%v)=%q want %q", tt.in, got, tt.want)
		}
	}
}

func TestHandleContentType(t *testing.T) {
	got, err := HandleContentType("application/json", "", "")
	if err != nil || got != "application/json" {
		t.Fatalf("default=%q err=%v", got, err)
	}
	got, err = HandleContentType("application/json", "text/plain", "")
	if err != nil || got != "text/plain" {
		t.Fatalf("config=%q err=%v", got, err)
	}
	got, err = HandleContentType("application/json", "text/plain", "application/xml")
	if err != nil || got != "application/xml" {
		t.Fatalf("req=%q err=%v", got, err)
	}
	if _, err := HandleContentType("", "", ""); err == nil {
		t.Fatal("expected empty content-type error")
	}
}

func TestPathTemplate(t *testing.T) {
	tpl, err := CreatePathTemplate("/users/{id}/posts/{pid}")
	if err != nil {
		t.Fatal(err)
	}
	got, err := PathTemplateRun(tpl, map[string]any{"id": 42, "pid": "abc"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "/users/42/posts/abc" {
		t.Fatalf("got=%q", got)
	}
}

func TestGetBodyReader(t *testing.T) {
	if GetBodyReader(nil).IsErr() {
		t.Fatal("nil body should succeed")
	}
	if !bytes.Equal(GetBodyReader([]byte("hi")).Unwrap(), []byte("hi")) {
		t.Fatal("[]byte")
	}
	if !bytes.Equal(GetBodyReader("hi").Unwrap(), []byte("hi")) {
		t.Fatal("string")
	}
	if !bytes.Equal(GetBodyReader(bytes.NewBufferString("buf")).Unwrap(), []byte("buf")) {
		t.Fatal("buffer")
	}
	vals := url.Values{"a": {"1"}}
	if got := string(GetBodyReader(vals).Unwrap()); got != "a=1" {
		t.Fatalf("url.Values=%q", got)
	}
	if got := string(GetBodyReader(strings.NewReader("reader")).Unwrap()); got != "reader" {
		t.Fatalf("reader=%q", got)
	}
	type payload struct {
		Name string `json:"name"`
	}
	raw := GetBodyReader(payload{Name: "x"}).Unwrap()
	if !bytes.Contains(raw, []byte(`"name":"x"`)) {
		t.Fatalf("json=%s", raw)
	}
}
