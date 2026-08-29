package lava

import (
	"context"
	"testing"
)

type orderRecorder struct {
	name string
	seq  *[]string
}

func (o orderRecorder) String() string { return o.name }
func (o orderRecorder) Middleware(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, req Request) (Response, error) {
		*o.seq = append(*o.seq, o.name+"-in")
		resp, err := next(ctx, req)
		*o.seq = append(*o.seq, o.name+"-out")
		return resp, err
	}
}

func TestChainOrder(t *testing.T) {
	var seq []string
	mw := Chain(
		orderRecorder{name: "a", seq: &seq},
		orderRecorder{name: "b", seq: &seq},
		nil,
	)
	handler := mw.Middleware(func(context.Context, Request) (Response, error) {
		seq = append(seq, "handler")
		return nil, nil
	})
	if _, err := handler(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	want := []string{"a-in", "b-in", "handler", "b-out", "a-out"}
	if len(seq) != len(want) {
		t.Fatalf("seq=%v want=%v", seq, want)
	}
	for i := range want {
		if seq[i] != want[i] {
			t.Fatalf("seq=%v want=%v", seq, want)
		}
	}
}

func TestWithMiddleware(t *testing.T) {
	mw := WithMiddleware("named", func(next HandlerFunc) HandlerFunc {
		return next
	})
	if mw.String() != "named" {
		t.Fatalf("name=%q", mw.String())
	}
	if Chain().String() != "chain" {
		t.Fatal("empty chain name")
	}
}
