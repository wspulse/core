package wspulse_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	wspulse "github.com/wspulse/core"
)

func TestFrame_ZeroValue_HasEmptyFields(t *testing.T) {
	var f wspulse.Frame
	if f.Event != "" {
		t.Errorf("Event: want empty, got %q", f.Event)
	}
	if f.Payload != nil {
		t.Errorf("Payload: want nil, got %v", f.Payload)
	}
}

func TestFrame_FieldAssignment(t *testing.T) {
	payload := []byte(`{"key":"val"}`)
	f := wspulse.Frame{Event: "msg", Payload: payload}
	if f.Event != "msg" {
		t.Errorf("Event: want %q, got %q", "msg", f.Event)
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
	f := wspulse.Frame{Event: "ping"}
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

// ---- Sentinel error conventions ---------------------------------------------

func TestSentinelErrors_HaveWspulsePrefix(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"ErrConnectionClosed", wspulse.ErrConnectionClosed},
		{"ErrSendBufferFull", wspulse.ErrSendBufferFull},
	}
	for _, tc := range cases {
		if !strings.HasPrefix(tc.err.Error(), "wspulse:") {
			t.Errorf("%s.Error() = %q, want prefix %q", tc.name, tc.err.Error(), "wspulse:")
		}
	}
}

func TestSentinelErrors_SupportErrorsIs(t *testing.T) {
	wrapped := errors.Join(
		wspulse.ErrConnectionClosed,
		errors.New("extra context"),
	)
	if !errors.Is(wrapped, wspulse.ErrConnectionClosed) {
		t.Error("errors.Is should match ErrConnectionClosed in joined error")
	}
}

// ---- Frame copy semantics (lifecycle awareness) -----------------------------

func TestFrame_PayloadSharing_AfterCopy(t *testing.T) {
	original := wspulse.Frame{
		Event:   "msg",
		Payload: []byte(`{"data":"original"}`),
	}
	copied := original

	// Mutate the copied frame's Payload in-place.
	copied.Payload[0] = 'X'

	// Since slices share backing arrays, the original is also affected.
	// This test documents the expected Go behavior so users are aware.
	if original.Payload[0] != 'X' {
		t.Error("expected shallow copy: modifying copy's Payload should affect original")
	}
}

func TestFrame_IndependentPayload_RequiresExplicitCopy(t *testing.T) {
	original := wspulse.Frame{
		Event:   "msg",
		Payload: []byte(`{"data":"safe"}`),
	}

	// Proper deep copy pattern for Frame.
	independent := wspulse.Frame{
		Event: original.Event,
	}
	independent.Payload = make([]byte, len(original.Payload))
	copy(independent.Payload, original.Payload)

	independent.Payload[0] = 'X'
	if original.Payload[0] == 'X' {
		t.Error("explicit copy should make Payload independent")
	}
}
