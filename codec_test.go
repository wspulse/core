package wspulse_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	wspulse "github.com/wspulse/core"
)

func TestJSONCodec_FrameType_IsTextMessage(t *testing.T) {
	if got := wspulse.JSONCodec.FrameType(); got != wspulse.TextMessage {
		t.Errorf("FrameType: want %d (TextMessage), got %d", wspulse.TextMessage, got)
	}
}

func TestJSONCodec_Encode_ProducesValidJSON(t *testing.T) {
	f := wspulse.Frame{ID: "1", Event: "msg", Payload: []byte(`{"text":"hello"}`)}
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
	f := wspulse.Frame{ID: "abc", Event: "sys", Payload: []byte(`{"x":1}`)}
	data, _ := wspulse.JSONCodec.Encode(f)
	var m map[string]json.RawMessage
	_ = json.Unmarshal(data, &m)
	assertJSONString(t, m, "id", "abc")
	assertJSONString(t, m, "event", "sys")
	if _, ok := m["payload"]; !ok {
		t.Error("expected 'payload' key in encoded JSON")
	}
}

func TestJSONCodec_Encode_EmptyIDOmitted(t *testing.T) {
	f := wspulse.Frame{Event: "msg", Payload: []byte(`"data"`)}
	data, _ := wspulse.JSONCodec.Encode(f)
	if bytes.Contains(data, []byte(`"id"`)) {
		t.Error("expected 'id' key to be omitted when ID is empty")
	}
}

func TestJSONCodec_Encode_EmptyEventOmitted(t *testing.T) {
	f := wspulse.Frame{ID: "1", Payload: []byte(`"data"`)}
	data, _ := wspulse.JSONCodec.Encode(f)
	if bytes.Contains(data, []byte(`"event"`)) {
		t.Error("expected 'event' key to be omitted when Event is empty")
	}
}

func TestJSONCodec_Encode_PayloadIsNotBase64(t *testing.T) {
	payload := []byte(`{"nested":{"k":"v"}}`)
	f := wspulse.Frame{Event: "msg", Payload: payload}
	data, _ := wspulse.JSONCodec.Encode(f)
	if !bytes.Contains(data, []byte(`"nested"`)) {
		t.Errorf("encoded JSON should embed payload as-is, got: %s", data)
	}
}

func TestJSONCodec_Decode_AllFields(t *testing.T) {
	input := `{"id":"x","event":"sys","payload":{"event":"join"}}`
	f, err := wspulse.JSONCodec.Decode([]byte(input))
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if f.ID != "x" {
		t.Errorf("ID: want %q, got %q", "x", f.ID)
	}
	if f.Event != "sys" {
		t.Errorf("Event: want %q, got %q", "sys", f.Event)
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
	if f.ID != "" || f.Event != "" {
		t.Errorf("unexpected non-empty fields: ID=%q Event=%q", f.ID, f.Event)
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
		Event:   "msg",
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
	if decoded.Event != original.Event {
		t.Errorf("Event: want %q, got %q", original.Event, decoded.Event)
	}
	if !bytes.Equal(decoded.Payload, original.Payload) {
		t.Errorf("Payload: want %s, got %s", original.Payload, decoded.Payload)
	}
}

func TestJSONCodec_Roundtrip_NilPayload(t *testing.T) {
	original := wspulse.Frame{ID: "no-payload", Event: "ping"}
	data, _ := wspulse.JSONCodec.Encode(original)
	decoded, err := wspulse.JSONCodec.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if decoded.ID != original.ID || decoded.Event != original.Event {
		t.Errorf("fields changed: got %+v", decoded)
	}
}

// ---- Lifecycle: buffer independence -----------------------------------------

func TestJSONCodec_Decode_PayloadIsIndependentFromInput(t *testing.T) {
	input := []byte(`{"event":"msg","payload":{"key":"value"}}`)
	original := make([]byte, len(input))
	copy(original, input)

	frame, err := wspulse.JSONCodec.Decode(input)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Overwrite the input buffer to simulate buffer reuse in a read loop.
	for i := range input {
		input[i] = 'X'
	}

	// The decoded Frame's Payload must remain intact.
	if !bytes.Contains(frame.Payload, []byte(`"key"`)) {
		t.Errorf("Payload corrupted after input buffer modification: %s", frame.Payload)
	}
}

func TestJSONCodec_Encode_OutputIsIndependentFromInput(t *testing.T) {
	payload := []byte(`{"mutable":"yes"}`)
	frame := wspulse.Frame{Event: "msg", Payload: payload}

	data, err := wspulse.JSONCodec.Encode(frame)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	snapshot := make([]byte, len(data))
	copy(snapshot, data)

	// Mutate the original Payload after Encode.
	for i := range payload {
		payload[i] = '0'
	}

	// Encoded output must remain intact.
	if !bytes.Equal(data, snapshot) {
		t.Errorf("Encode output changed after input mutation: got %s", data)
	}
}

// ---- Edge cases: Payload variants -------------------------------------------

func TestJSONCodec_Encode_InvalidPayload_ReturnsError(t *testing.T) {
	frame := wspulse.Frame{Event: "msg", Payload: []byte("not valid json")}
	_, err := wspulse.JSONCodec.Encode(frame)
	if err == nil {
		t.Error("expected error when Payload is not valid JSON")
	}
}

func TestJSONCodec_Roundtrip_NullPayload(t *testing.T) {
	original := wspulse.Frame{Event: "msg", Payload: []byte("null")}
	data, err := wspulse.JSONCodec.Encode(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	decoded, err := wspulse.JSONCodec.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if string(decoded.Payload) != "null" {
		t.Errorf("Payload: want %q, got %q", "null", string(decoded.Payload))
	}
}

func TestJSONCodec_Roundtrip_ArrayPayload(t *testing.T) {
	original := wspulse.Frame{Event: "list", Payload: []byte(`[1,2,3]`)}
	data, _ := wspulse.JSONCodec.Encode(original)
	decoded, err := wspulse.JSONCodec.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(decoded.Payload, original.Payload) {
		t.Errorf("Payload: want %s, got %s", original.Payload, decoded.Payload)
	}
}

func TestJSONCodec_Roundtrip_StringPayload(t *testing.T) {
	original := wspulse.Frame{Event: "echo", Payload: []byte(`"hello world"`)}
	data, _ := wspulse.JSONCodec.Encode(original)
	decoded, err := wspulse.JSONCodec.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(decoded.Payload, original.Payload) {
		t.Errorf("Payload: want %s, got %s", original.Payload, decoded.Payload)
	}
}

func TestJSONCodec_Roundtrip_NumericPayload(t *testing.T) {
	original := wspulse.Frame{Event: "val", Payload: []byte(`42`)}
	data, _ := wspulse.JSONCodec.Encode(original)
	decoded, err := wspulse.JSONCodec.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(decoded.Payload, original.Payload) {
		t.Errorf("Payload: want %s, got %s", original.Payload, decoded.Payload)
	}
}

func TestJSONCodec_Roundtrip_UnicodePayload(t *testing.T) {
	original := wspulse.Frame{
		Event:   "msg",
		Payload: []byte(`{"text":"你好世界 🌍"}`),
	}
	data, err := wspulse.JSONCodec.Encode(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	decoded, err := wspulse.JSONCodec.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(decoded.Payload, original.Payload) {
		t.Errorf("Payload: want %s, got %s", original.Payload, decoded.Payload)
	}
}

func TestJSONCodec_Roundtrip_DeeplyNestedPayload(t *testing.T) {
	original := wspulse.Frame{
		Event:   "nested",
		Payload: []byte(`{"a":{"b":{"c":{"d":"deep"}}}}`),
	}
	data, _ := wspulse.JSONCodec.Encode(original)
	decoded, err := wspulse.JSONCodec.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(decoded.Payload, original.Payload) {
		t.Errorf("Payload: want %s, got %s", original.Payload, decoded.Payload)
	}
}

// ---- Forward compatibility: unknown fields ----------------------------------

func TestJSONCodec_Decode_ExtraFields_Ignored(t *testing.T) {
	input := `{"id":"1","event":"msg","payload":{"k":"v"},"version":2,"extra":"ignored"}`
	frame, err := wspulse.JSONCodec.Decode([]byte(input))
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if frame.ID != "1" || frame.Event != "msg" {
		t.Errorf("unexpected field values: %+v", frame)
	}
}

func TestJSONCodec_Decode_NullPayloadField(t *testing.T) {
	input := `{"event":"ping","payload":null}`
	frame, err := wspulse.JSONCodec.Decode([]byte(input))
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if string(frame.Payload) != "null" {
		t.Errorf("Payload: want %q, got %q", "null", string(frame.Payload))
	}
}

// ---- Concurrency safety -----------------------------------------------------

func TestJSONCodec_ConcurrentEncodeDecode(t *testing.T) {
	const goroutines = 50
	var waitGroup sync.WaitGroup
	waitGroup.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer waitGroup.Done()

			original := wspulse.Frame{
				ID:      "concurrent",
				Event:   "msg",
				Payload: []byte(`{"data":"test"}`),
			}
			data, err := wspulse.JSONCodec.Encode(original)
			if err != nil {
				t.Errorf("Encode failed: %v", err)
				return
			}
			decoded, err := wspulse.JSONCodec.Decode(data)
			if err != nil {
				t.Errorf("Decode failed: %v", err)
				return
			}
			if decoded.ID != original.ID || decoded.Event != original.Event {
				t.Errorf("roundtrip mismatch: got %+v", decoded)
			}
		}()
	}
	waitGroup.Wait()
}

// ---- Large payload ----------------------------------------------------------

func TestJSONCodec_Roundtrip_LargePayload(t *testing.T) {
	// Build a ~10 KB JSON payload.
	value := strings.Repeat("x", 10000)
	payload, _ := json.Marshal(map[string]string{"big": value})
	original := wspulse.Frame{Event: "bulk", Payload: payload}

	data, err := wspulse.JSONCodec.Encode(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	decoded, err := wspulse.JSONCodec.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(decoded.Payload, original.Payload) {
		t.Error("large payload roundtrip mismatch")
	}
}

// ---- Benchmarks -------------------------------------------------------------

func BenchmarkJSONCodec_Encode(b *testing.B) {
	frame := wspulse.Frame{
		ID:      "bench-1",
		Event:   "msg",
		Payload: []byte(`{"user":"alice","text":"hello"}`),
	}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = wspulse.JSONCodec.Encode(frame)
	}
}

func BenchmarkJSONCodec_Decode(b *testing.B) {
	data := []byte(`{"id":"bench-1","event":"msg","payload":{"user":"alice","text":"hello"}}`)
	b.ReportAllocs()
	for b.Loop() {
		_, _ = wspulse.JSONCodec.Decode(data)
	}
}

func BenchmarkJSONCodec_Roundtrip(b *testing.B) {
	frame := wspulse.Frame{
		ID:      "bench-rt",
		Event:   "msg",
		Payload: []byte(`{"user":"alice","text":"hello"}`),
	}
	b.ReportAllocs()
	for b.Loop() {
		data, _ := wspulse.JSONCodec.Encode(frame)
		_, _ = wspulse.JSONCodec.Decode(data)
	}
}

// ---- Helpers ----------------------------------------------------------------

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
