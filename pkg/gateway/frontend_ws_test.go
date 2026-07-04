package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/coder/websocket"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pubgo/lava/v2/pkg/encoding/protojson"
)

func TestResolveWSEncoding(t *testing.T) {
	cases := []struct {
		name        string
		query       string
		subprotocol string
		want        wsEncoding
	}{
		{"default json", "", "", wsEncodingJSON},
		{"query proto", "encoding=proto", "", wsEncodingProto},
		{"query binary", "encoding=binary", "", wsEncodingProto},
		{"subprotocol proto", "", "grpc-ws-proto", wsEncodingProto},
		{"subprotocol json wins over nothing", "", "grpc-ws-json", wsEncodingJSON},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &http.Request{URL: &url.URL{RawQuery: tc.query}}
			if got := resolveWSEncoding(r, tc.subprotocol); got != tc.want {
				t.Fatalf("resolveWSEncoding = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestStreamWS_RoundTripJSON verifies the streamWS adapter marshals/unmarshals
// protojson over websocket text frames end-to-end.
func TestStreamWS_RoundTripJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			t.Errorf("accept: %v", err)
			return
		}
		defer func() { _ = conn.CloseNow() }()

		s := &streamWS{conn: conn, ctx: r.Context(), encoding: wsEncodingJSON}

		req := &structpb.Struct{}
		if err = s.RecvMsg(req); err != nil {
			t.Errorf("server RecvMsg: %v", err)
			return
		}
		// Echo the same message back to the client.
		if err = s.SendMsg(req); err != nil {
			t.Errorf("server SendMsg: %v", err)
			return
		}
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + srv.URL[len("http"):]
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.CloseNow() }()

	want := &structpb.Struct{Fields: map[string]*structpb.Value{
		"msg": structpb.NewStringValue("hello-ws"),
	}}
	payload, err := protojson.Default.Marshal(want)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err = conn.Write(ctx, websocket.MessageText, payload); err != nil {
		t.Fatalf("client write: %v", err)
	}

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("client read: %v", err)
	}

	got := &structpb.Struct{}
	if err = protojson.Default.Unmarshal(data, got); err != nil {
		t.Fatalf("client unmarshal: %v", err)
	}
	if got.GetFields()["msg"].GetStringValue() != "hello-ws" {
		t.Fatalf("unexpected echo payload: %s", string(data))
	}
}

func TestWSCloseCode(t *testing.T) {
	if got := wsCloseCode(codes.OK); got != websocket.StatusNormalClosure {
		t.Fatalf("OK close code = %v, want normal closure", got)
	}
	if got := wsCloseCode(codes.NotFound); got != websocket.StatusInternalError {
		t.Fatalf("NotFound close code = %v, want internal error", got)
	}
}

// TestCloseWithStatus verifies the structured grpc-status close frame: the close
// reason carries a JSON {grpcStatus, grpcMessage} recoverable by the client.
func TestCloseWithStatus(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		code    codes.Code
		message string
		wsCode  websocket.StatusCode
	}{
		{"ok", nil, codes.OK, "", websocket.StatusNormalClosure},
		{"not found", status.Error(codes.NotFound, "missing user"), codes.NotFound, "missing user", websocket.StatusInternalError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
				if err != nil {
					t.Errorf("accept: %v", err)
					return
				}
				closeWithStatus(conn, tc.err)
			}))
			defer srv.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			wsURL := "ws" + srv.URL[len("http"):]
			conn, _, err := websocket.Dial(ctx, wsURL, nil)
			if err != nil {
				t.Fatalf("dial: %v", err)
			}
			defer func() { _ = conn.CloseNow() }()

			_, _, readErr := conn.Read(ctx)
			var ce websocket.CloseError
			if !errors.As(readErr, &ce) {
				t.Fatalf("expected close error, got %v", readErr)
			}
			if ce.Code != tc.wsCode {
				t.Fatalf("close code = %v, want %v", ce.Code, tc.wsCode)
			}

			var payload struct {
				GRPCStatus  uint32 `json:"grpcStatus"`
				GRPCMessage string `json:"grpcMessage"`
			}
			if err = json.Unmarshal([]byte(ce.Reason), &payload); err != nil {
				t.Fatalf("unmarshal close reason %q: %v", ce.Reason, err)
			}
			if codes.Code(payload.GRPCStatus) != tc.code {
				t.Fatalf("grpcStatus = %d, want %d", payload.GRPCStatus, tc.code)
			}
			if payload.GRPCMessage != tc.message {
				t.Fatalf("grpcMessage = %q, want %q", payload.GRPCMessage, tc.message)
			}
		})
	}
}
