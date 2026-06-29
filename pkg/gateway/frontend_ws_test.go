package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/coder/websocket"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pubgo/lava/v2/core/encoding/protojson"
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
		defer conn.CloseNow()

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
		conn.Close(websocket.StatusNormalClosure, "")
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + srv.URL[len("http"):]
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

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
