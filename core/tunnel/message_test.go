package tunnel

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
)

func TestWriteReadMessageRoundTrip(t *testing.T) {
	orig := &Message{Type: MessageTypeHeartbeat, ID: "msg-1", Payload: []byte(`{"ok":true}`)}

	var buf bytes.Buffer
	if err := WriteMessage(&buf, orig); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	got, err := ReadMessage(&buf)
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if got.Type != orig.Type || got.ID != orig.ID {
		t.Fatalf("round-trip mismatch: %+v vs %+v", got, orig)
	}
}

func TestReadMessageOversize(t *testing.T) {
	var buf bytes.Buffer
	header := make([]byte, messageHeaderSize)
	// declare size larger than MaxMessageSize
	binary.BigEndian.PutUint32(header, MaxMessageSize+1)
	if _, err := buf.Write(header); err != nil {
		t.Fatalf("write header: %v", err)
	}

	_, err := ReadMessage(&buf)
	if err == nil || !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("expected ErrInvalidMessage, got %v", err)
	}
}
