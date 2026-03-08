package wspulse_test

import (
	"encoding/json"
	"testing"

	wspulse "github.com/wspulse/core"
)

func TestFrame_ZeroValue_HasEmptyFields(t *testing.T) {
	var f wspulse.Frame
	if f.ID != "" {
		t.Errorf("ID: want empty, got %q", f.ID)
	}
	if f.Type != "" {
		t.Errorf("Type: want empty, got %q", f.Type)
	}
	if f.Payload != nil {
		t.Errorf("Payload: want nil, got %v", f.Payload)
	}
}

func TestFrame_FieldAssignment(t *testing.T) {
	payload := []byte(`{"key":"val"}`)
	f := wspulse.Frame{ID: "abc", Type: "msg", Payload: payload}
	if f.ID != "abc" {
		t.Errorf("ID: want %q, got %q", "abc", f.ID)
	}
	if f.Type != "msg" {
		t.Errorf("Type: want %q, got %q", "msg", f.Type)
	}
	if string(f.Payload) != string(payload) {
		t.Errorf("Payload: want %s, got %s", payload, f.Payload)
	}
}

func TestSentinelErrors_AreDistinct(t *testing.T) {
	sentinels := []error{
		wspulse.ErrConnectionClosed,
		wspulse.ErrSendBufferFull,
	}
	for i, a := range sentinels {
		for j, b := range sentinels {
			if i != j && a == b {
				t.Fatalf("errors[%d] and errors[%d] are the same: %v", i, j, a)
			}
		}
	}
}

func TestSentinelErrors_HaveNonEmptyMessages(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"ErrConnectionClosed", wspulse.ErrConnectionClosed},
		{"ErrSendBufferFull", wspulse.ErrSendBufferFull},
	}
	for _, tc := range cases {
		if tc.err.Error() == "" {
			t.Errorf("%s.Error() returned empty string", tc.name)
		}
	}
}

func TestMessageTypeConstants(t *testing.T) {
	if wspulse.TextMessage != 1 {
		t.Errorf("TextMessage: want 1, got %d", wspulse.TextMessage)
	}
	if wspulse.BinaryMessage != 2 {
		t.Errorf("BinaryMessage: want 2, got %d", wspulse.BinaryMessage)
	}
}

func TestWireFrame_EmptyPayload_OmittedFromJSON(t *testing.T) {
	f := wspulse.Frame{Type: "ping"}
	data, err := wspulse.JSONCodec.Encode(f)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if _, ok := m["payload"]; ok {
		t.Error("expected 'payload' to be absent when Payload is nil")
	}
}
