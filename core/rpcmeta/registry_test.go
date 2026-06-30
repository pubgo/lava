package rpcmeta

import "testing"

func TestRegisterAndGet(t *testing.T) {
	meta := &RpcMeta{Name: "svc.Test", Method: "/svc/Test"}
	if err := Register(meta); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if got := Get("svc.Test"); got != meta {
		t.Fatalf("get by name failed")
	}
	if got := Get("/svc/Test"); got != meta {
		t.Fatalf("get by method failed")
	}
}

func TestGetMissingReturnsNil(t *testing.T) {
	if got := Get("does-not-exist"); got != nil {
		t.Fatalf("expected nil for missing meta")
	}
}

func TestRegisterDuplicatePanics(t *testing.T) {
	if err := Register(&RpcMeta{Name: "svc.Dup", Method: "/svc/Dup"}); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic on duplicate registration")
		}
	}()
	_ = Register(&RpcMeta{Name: "svc.Dup", Method: "/svc/Dup2"})
}

func TestRegisterInvalidPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic on nil meta")
		}
	}()
	_ = Register(nil)
}
