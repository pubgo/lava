package encoding

import (
	"testing"
)

// fakeCodec 是用于测试的最小 Codec 实现。
type fakeCodec struct{ name string }

func (c *fakeCodec) Name() string                       { return c.name }
func (c *fakeCodec) Encode(v any) ([]byte, error)       { return []byte(c.name), nil }
func (c *fakeCodec) Decode(data []byte, v any) error    { return nil }
func (c *fakeCodec) Marshal(v any) ([]byte, error)      { return []byte(c.name), nil }
func (c *fakeCodec) Unmarshal(data []byte, v any) error { return nil }

func TestRegisterAndGet(t *testing.T) {
	name := "test-codec-register"
	Register(name, &fakeCodec{name: name})

	got := Get(name)
	if got == nil {
		t.Fatalf("expected codec %q to be registered", name)
	}
	if got.Name() != name {
		t.Fatalf("expected name %q, got %q", name, got.Name())
	}
}

func TestGetMissingReturnsNil(t *testing.T) {
	if got := Get("does-not-exist-codec"); got != nil {
		t.Fatalf("expected nil for missing codec, got %v", got)
	}
}

func TestKeysContainsRegistered(t *testing.T) {
	name := "test-codec-keys"
	Register(name, &fakeCodec{name: name})

	found := false
	for _, k := range Keys() {
		if k == name {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Keys() should contain %q", name)
	}
}

func TestGetWithContentType(t *testing.T) {
	tests := map[string]string{
		"application/json":    "json",
		"application/proto":   "proto",
		"application/grpc":    "proto",
		"application/unknown": "",
	}

	for ct, wantName := range tests {
		gotName := cdcMapping[ct]
		if gotName != wantName {
			t.Fatalf("content-type %q: expected mapping %q, got %q", ct, wantName, gotName)
		}
	}

	// 未知 content-type 应返回 nil
	if got := GetWithCT("application/unknown"); got != nil {
		t.Fatalf("unknown content-type should return nil codec")
	}
}
