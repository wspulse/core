package wspulse_test

import (
	"bytes"
	"encoding/json"
	"testing"

	wspulse "github.com/wspulse/core"
)

func TestJSONCodec_FrameType_IsTextMessage(t *testing.T) {
	if got := wspulse.JSONCodec.FrameType(); got != wspulse.TextMessage {
		t.Errorf("FrameType: want %d (TextMessage), got %d", wspulse.TextMessage, got)
	}
}

func TestJSONCodec_Encode_ProducesValidJSON(t *testing.T) {
	f := wspulse.Frame{ID: "1", Type: "msg", Payload: []byte(`{"text":"hello"}`)}
	data, err := wspulse.JSONCodec.Encode(f)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("encoded bytes are not valid JSON: %v", err)
	}
}

func TestJSONCodec_Encode_PopulatesAllFields(t *testing.T) {
	f := wspulse.Frame{ID: "abc", Type: "sys", Payload: []byte(`{"x":1}`)}
	data, _ := wspulse.JSONCodec.Encode(f)
	var m map[string]json.RawMessage
	_ = json.Unmarshal(data, &m)
	assertJSONString(t, m, "id", "abc")
	assertJSONString(t, m, "type", "sys")
	if _, ok := m["payload"]; !ok {
		t.Error("expected 'payload' key in encoded JSON")
	}
}

func TestJSONCodec_Encode_EmptyIDOmitted(t *testing.T) {
	f := wspulse.Frame{Type: "msg", Payload: []byte(`"data"`)}
	data, _ := wspulse.JSONCodec.Encode(f)
	if bytes.Contains(data, []byte(`"id"`)) {
		t.Error("expected 'id' key to be omitted when ID is empty")
	}
}

func TestJSONCodec_Encode_EmptyTypeOmitted(t *testing.T) {
	f := wspulse.Frame{ID: "1", Payload: []byte(`"data"`)}
	data, _ := wspulse.JSONCodec.Encode(f)
	if bytes.Contains(data, []byte(`"type"`)) {
		t.Error("expected 'type' key to be omitted when Type is empty")
	}
}

func TestJSONCodec_Encode_PayloadIsNotBase64(t *testing.T) {
	payload := []byte(`{"nested":{"k":"v"}}`)
	f := wspulse.Frame{Type: "msg", Payload: payload}
	data, _ := wspulse.JSONCodec.Encode(f)
	if !bytes.Contains(data, []byte(`"nested"`)) {
		t.Errorf("encoded JSON should embed payload as-is, got: %s", data)
	}
}

func TestJSONCodec_Decode_AllFields(t *testing.T) {
	input := `{"id":"x","type":"sys","payload":{"event":"join"}}`
	f, err := wspulse.JSONCodec.Decode([]byte(input))
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if f.ID != "x" {
		t.Errorf("ID: want %q, got %q", "x", f.ID)
	}
	if f.Type != "sys" {
		t.Errorf("Type: want %q, got %q", "sys", f.Type)
	}
	if !bytes.Contains(f.Payload, []byte(`"event"`)) {
		t.Errorf("Payload missing 'event' key: %s", f.Payload)
	}
}

func TestJSONCodec_Decode_MissingFieldsAreEmpty(t *testing.T) {
	f, err := wspulse.JSONCodec.Decode([]byte(`{}`))
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if f.ID != "" || f.Type != "" {
		t.Errorf("unexpected non-empty fields: ID=%q Type=%q", f.ID, f.Type)
	}
}

func TestJSONCodec_Decode_InvalidJSON_ReturnsError(t *testing.T) {
	_, err := wspulse.JSONCodec.Decode([]byte("not-json"))
	if err == nil {
		t.Error("expected error for invalid JSON input")
	}
}

func TestJSONCodec_Decode_EmptyInput_ReturnsError(t *testing.T) {
	_, err := wspulse.JSONCodec.Decode([]byte(""))
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestJSONCodec_Roundtrip_PreservesAllFields(t *testing.T) {
	original := wspulse.Frame{
		ID:      "round-1",
		Type:    "msg",
		Payload: []byte(`{"a":1,"b":true}`),
	}
	data, err := wspulse.JSONCodec.Encode(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	decoded, err := wspulse.JSONCodec.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if decoded.ID != original.ID {
		t.Errorf("ID: want %q, got %q", original.ID, decoded.ID)
	}
	if decoded.Type != original.Type {
		t.Errorf("Type: want %q, got %q", original.Type, decoded.Type)
	}
	if !bytes.Equal(decoded.Payload, original.Payload) {
		t.Errorf("Payload: want %s, got %s", original.Payload, decoded.Payload)
	}
}

func TestJSONCodec_Roundtrip_NilPayload(t *testing.T) {
	original := wspulse.Frame{ID: "no-payload", Type: "ping"}
	data, _ := wspulse.JSONCodec.Encode(original)
	decoded, err := wspulse.JSONCodec.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if decoded.ID != original.ID || decoded.Type != original.Type {
		t.Errorf("fields changed: got %+v", decoded)
	}
}

func assertJSONString(t *testing.T, m map[string]json.RawMessage, key, want string) {
	t.Helper()
	raw, ok := m[key]
	if !ok {
		t.Errorf("key %q not found in JSON", key)
		return
	}
	var got string
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Errorf("key %q: failed to unmarshal as string: %v", key, err)
		return
	}
	if got != want {
		t.Errorf("key %q: want %q, got %q", key, want, got)
	}
}
