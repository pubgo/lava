package gatewayutils

import "testing"

func TestNewDoubleArray_HasCommonPrefix(t *testing.T) {
	da := NewDoubleArray([][]string{
		{"body"},
		{"nested", "field"},
	})
	if !da.HasCommonPrefix([]string{"body"}) {
		t.Fatal("expected body prefix")
	}
	if !da.HasCommonPrefix([]string{"nested", "field", "extra"}) {
		t.Fatal("expected nested.field prefix")
	}
	if da.HasCommonPrefix([]string{"other"}) {
		t.Fatal("unexpected other prefix")
	}
}

func TestNewDoubleArray_Empty(t *testing.T) {
	da := NewDoubleArray(nil)
	if da.HasCommonPrefix([]string{"x"}) {
		t.Fatal("empty da should not match")
	}
}

func TestQuote(t *testing.T) {
	if got := string(quote([]byte(`hello`))); got != `"hello"` {
		t.Fatalf("quote unquoted=%q", got)
	}
	if got := string(quote([]byte(`"already"`))); got != `"already"` {
		t.Fatalf("quote already=%q", got)
	}
}
