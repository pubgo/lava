package resty

import (
	"net/http"
	"testing"
)

func TestIsRedirect(t *testing.T) {
	for _, code := range []int{
		http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusSeeOther,
		http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect,
	} {
		if !IsRedirect(code) {
			t.Fatalf("expected redirect for %d", code)
		}
	}
	if IsRedirect(http.StatusOK) || IsRedirect(http.StatusBadRequest) {
		t.Fatal("non-redirect codes matched")
	}
}

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
