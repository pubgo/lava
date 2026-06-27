package flags

import (
	"testing"

	"github.com/pubgo/redant"
)

func TestRegisterAndGetFlags(t *testing.T) {
	before := len(GetFlags())

	Register(redant.Option{Flag: "test-flag-xyz"})

	after := GetFlags()
	if len(after) != before+1 {
		t.Fatalf("expected %d flags after register, got %d", before+1, len(after))
	}

	found := false
	for _, f := range after {
		if f.Flag == "test-flag-xyz" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("GetFlags() should contain the registered flag")
	}
}
