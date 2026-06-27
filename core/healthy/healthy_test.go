package healthy

import (
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestRegisterGetList(t *testing.T) {
	name := "test-healthy-register"
	h := func(req fiber.Ctx) error { return nil }

	Register(name, h)

	if got := Get(name); got == nil {
		t.Fatalf("expected handler %q to be registered", name)
	}

	found := false
	for _, n := range List() {
		if n == name {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("List() should contain %q", name)
	}
}

func TestGetMissingReturnsNil(t *testing.T) {
	if got := Get("does-not-exist-healthy"); got != nil {
		t.Fatalf("expected nil for missing handler")
	}
}

func TestRegisterDuplicatePanics(t *testing.T) {
	name := "test-healthy-dup"
	Register(name, func(req fiber.Ctx) error { return nil })

	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic on duplicate registration")
		}
	}()
	Register(name, func(req fiber.Ctx) error { return nil })
}

func TestRegisterEmptyPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic on empty name/handler")
		}
	}()
	Register("", nil)
}
