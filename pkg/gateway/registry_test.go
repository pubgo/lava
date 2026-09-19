package gateway

import (
	"testing"

	"google.golang.org/protobuf/types/known/emptypb"
)

func TestLookupOperation_ReturnsCachedRegistration(t *testing.T) {
	t.Parallel()

	mux := NewMux()
	full := "/lookup.v1.Echo/Ping"
	mth := &methodWrapper{
		grpcFullMethod: full,
		inputType:      (&emptypb.Empty{}).ProtoReflect().Type(),
		outputType:     (&emptypb.Empty{}).ProtoReflect().Type(),
	}
	mux.opts.handlers[full] = mth

	op1 := mux.LookupOperation(full)
	if op1 == nil || op1.FullMethod != full {
		t.Fatalf("LookupOperation=%v", op1)
	}
	op2 := mux.LookupOperation(full)
	if op1 != op2 {
		t.Fatal("LookupOperation should return the same Operation pointer")
	}
	if mux.LookupOperation("/missing") != nil {
		t.Fatal("expected nil for unknown method")
	}
}

func TestRegisterRouter_BuildsOperationOnce(t *testing.T) {
	t.Parallel()

	// operationFromMethod must reuse methodWrapper.op when set at register time.
	mth := &methodWrapper{
		grpcFullMethod: "/reg.v1.Echo/Ping",
		inputType:      (&emptypb.Empty{}).ProtoReflect().Type(),
		outputType:     (&emptypb.Empty{}).ProtoReflect().Type(),
		op: &Operation{
			FullMethod: "/reg.v1.Echo/Ping",
			InputType:  (&emptypb.Empty{}).ProtoReflect().Type(),
			OutputType: (&emptypb.Empty{}).ProtoReflect().Type(),
		},
	}
	if got := operationFromMethod(mth); got != mth.op {
		t.Fatalf("expected cached op pointer")
	}
}
