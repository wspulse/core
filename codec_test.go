package wspulse_test

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wspulse "github.com/wspulse/core"
)

func TestJSONCodec_FrameType_IsTextMessage(t *testing.T) {
	assert.Equal(t, wspulse.TextMessage, wspulse.JSONCodec.FrameType())
}

func TestJSONCodec_Encode_ProducesValidJSON(t *testing.T) {
	f := wspulse.Message{Event: "msg", Payload: []byte(`{"text":"hello"}`)}
	data, err := wspulse.JSONCodec.Encode(f)
	require.NoError(t, err)
	var m map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &m), "encoded bytes are not valid JSON")
}

func TestJSONCodec_Encode_PopulatesAllFields(t *testing.T) {
	f := wspulse.Message{Event: "sys", Payload: []byte(`{"x":1}`)}
	data, _ := wspulse.JSONCodec.Encode(f)
	var m map[string]json.RawMessage
	_ = json.Unmarshal(data, &m)
	assertJSONString(t, m, "event", "sys")
	assert.Contains(t, m, "payload")
}

func TestJSONCodec_Encode_EmptyEventOmitted(t *testing.T) {
	f := wspulse.Message{Payload: []byte(`"data"`)}
	data, _ := wspulse.JSONCodec.Encode(f)
	assert.NotContains(t, string(data), `"event"`)
}

func TestJSONCodec_Encode_PayloadIsNotBase64(t *testing.T) {
	payload := []byte(`{"nested":{"k":"v"}}`)
	f := wspulse.Message{Event: "msg", Payload: payload}
	data, _ := wspulse.JSONCodec.Encode(f)
	assert.Contains(t, string(data), `"nested"`,
		"encoded JSON should embed payload as-is, got: %s", data)
}

func TestJSONCodec_Decode_AllFields(t *testing.T) {
	input := `{"event":"sys","payload":{"event":"join"}}`
	f, err := wspulse.JSONCodec.Decode([]byte(input))
	require.NoError(t, err)
	assert.Equal(t, "sys", f.Event)
	assert.Contains(t, string(f.Payload), `"event"`)
}

func TestJSONCodec_Decode_MissingFieldsAreEmpty(t *testing.T) {
	f, err := wspulse.JSONCodec.Decode([]byte(`{}`))
	require.NoError(t, err)
	assert.Empty(t, f.Event)
}

func TestJSONCodec_Decode_InvalidJSON_ReturnsError(t *testing.T) {
	_, err := wspulse.JSONCodec.Decode([]byte("not-json"))
	assert.Error(t, err)
}

func TestJSONCodec_Decode_EmptyInput_ReturnsError(t *testing.T) {
	_, err := wspulse.JSONCodec.Decode([]byte(""))
	assert.Error(t, err)
}

func TestJSONCodec_Roundtrip_PreservesAllFields(t *testing.T) {
	original := wspulse.Message{
		Event:   "msg",
		Payload: []byte(`{"a":1,"b":true}`),
	}
	data, err := wspulse.JSONCodec.Encode(original)
	require.NoError(t, err)
	decoded, err := wspulse.JSONCodec.Decode(data)
	require.NoError(t, err)
	assert.Equal(t, original.Event, decoded.Event)
	assert.Equal(t, original.Payload, decoded.Payload)
}

func TestJSONCodec_Roundtrip_NilPayload(t *testing.T) {
	original := wspulse.Message{Event: "ping"}
	data, _ := wspulse.JSONCodec.Encode(original)
	decoded, err := wspulse.JSONCodec.Decode(data)
	require.NoError(t, err)
	assert.Equal(t, original.Event, decoded.Event)
}

// ---- Lifecycle: buffer independence -----------------------------------------

func TestJSONCodec_Decode_PayloadIsIndependentFromInput(t *testing.T) {
	input := []byte(`{"event":"msg","payload":{"key":"value"}}`)
	original := make([]byte, len(input))
	copy(original, input)

	msg, err := wspulse.JSONCodec.Decode(input)
	require.NoError(t, err)

	// Overwrite the input buffer to simulate buffer reuse in a read loop.
	for i := range input {
		input[i] = 'X'
	}

	// The decoded Message's Payload must remain intact.
	assert.Contains(t, string(msg.Payload), `"key"`,
		"Payload corrupted after input buffer modification")
}

func TestJSONCodec_Encode_OutputIsIndependentFromInput(t *testing.T) {
	payload := []byte(`{"mutable":"yes"}`)
	msg := wspulse.Message{Event: "msg", Payload: payload}

	data, err := wspulse.JSONCodec.Encode(msg)
	require.NoError(t, err)
	snapshot := make([]byte, len(data))
	copy(snapshot, data)

	// Mutate the original Payload after Encode.
	for i := range payload {
		payload[i] = '0'
	}

	// Encoded output must remain intact.
	assert.Equal(t, snapshot, data, "Encode output changed after input mutation")
}

// ---- Edge cases: Payload variants -------------------------------------------

func TestJSONCodec_Encode_InvalidPayload_ReturnsError(t *testing.T) {
	msg := wspulse.Message{Event: "msg", Payload: []byte("not valid json")}
	_, err := wspulse.JSONCodec.Encode(msg)
	assert.Error(t, err)
}

func TestJSONCodec_Roundtrip_NullPayload(t *testing.T) {
	original := wspulse.Message{Event: "msg", Payload: []byte("null")}
	data, err := wspulse.JSONCodec.Encode(original)
	require.NoError(t, err)
	decoded, err := wspulse.JSONCodec.Decode(data)
	require.NoError(t, err)
	assert.Equal(t, "null", string(decoded.Payload))
}

func TestJSONCodec_Roundtrip_ArrayPayload(t *testing.T) {
	original := wspulse.Message{Event: "list", Payload: []byte(`[1,2,3]`)}
	data, _ := wspulse.JSONCodec.Encode(original)
	decoded, err := wspulse.JSONCodec.Decode(data)
	require.NoError(t, err)
	assert.Equal(t, original.Payload, decoded.Payload)
}

func TestJSONCodec_Roundtrip_StringPayload(t *testing.T) {
	original := wspulse.Message{Event: "echo", Payload: []byte(`"hello world"`)}
	data, _ := wspulse.JSONCodec.Encode(original)
	decoded, err := wspulse.JSONCodec.Decode(data)
	require.NoError(t, err)
	assert.Equal(t, original.Payload, decoded.Payload)
}

func TestJSONCodec_Roundtrip_NumericPayload(t *testing.T) {
	original := wspulse.Message{Event: "val", Payload: []byte(`42`)}
	data, _ := wspulse.JSONCodec.Encode(original)
	decoded, err := wspulse.JSONCodec.Decode(data)
	require.NoError(t, err)
	assert.Equal(t, original.Payload, decoded.Payload)
}

func TestJSONCodec_Roundtrip_UnicodePayload(t *testing.T) {
	original := wspulse.Message{
		Event:   "msg",
		Payload: []byte(`{"text":"你好世界 🌍"}`),
	}
	data, err := wspulse.JSONCodec.Encode(original)
	require.NoError(t, err)
	decoded, err := wspulse.JSONCodec.Decode(data)
	require.NoError(t, err)
	assert.Equal(t, original.Payload, decoded.Payload)
}

func TestJSONCodec_Roundtrip_DeeplyNestedPayload(t *testing.T) {
	original := wspulse.Message{
		Event:   "nested",
		Payload: []byte(`{"a":{"b":{"c":{"d":"deep"}}}}`),
	}
	data, _ := wspulse.JSONCodec.Encode(original)
	decoded, err := wspulse.JSONCodec.Decode(data)
	require.NoError(t, err)
	assert.Equal(t, original.Payload, decoded.Payload)
}

// ---- Forward compatibility: unknown fields ----------------------------------

func TestJSONCodec_Decode_ExtraFields_Ignored(t *testing.T) {
	input := `{"event":"msg","payload":{"k":"v"},"version":2,"extra":"ignored"}`
	msg, err := wspulse.JSONCodec.Decode([]byte(input))
	require.NoError(t, err)
	assert.Equal(t, "msg", msg.Event)
}

func TestJSONCodec_Decode_NullPayloadField(t *testing.T) {
	input := `{"event":"ping","payload":null}`
	msg, err := wspulse.JSONCodec.Decode([]byte(input))
	require.NoError(t, err)
	assert.Equal(t, "null", string(msg.Payload))
}

// ---- Concurrency safety -----------------------------------------------------

func TestJSONCodec_ConcurrentEncodeDecode(t *testing.T) {
	const goroutines = 50
	var waitGroup sync.WaitGroup
	waitGroup.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer waitGroup.Done()

			original := wspulse.Message{
				Event:   "msg",
				Payload: []byte(`{"data":"test"}`),
			}
			data, err := wspulse.JSONCodec.Encode(original)
			if !assert.NoError(t, err, "Encode failed") {
				return
			}
			decoded, err := wspulse.JSONCodec.Decode(data)
			if !assert.NoError(t, err, "Decode failed") {
				return
			}
			assert.Equal(t, original.Event, decoded.Event, "roundtrip mismatch")
		}()
	}
	waitGroup.Wait()
}

// ---- Large payload ----------------------------------------------------------

func TestJSONCodec_Roundtrip_LargePayload(t *testing.T) {
	// Build a ~10 KB JSON payload.
	value := strings.Repeat("x", 10000)
	payload, _ := json.Marshal(map[string]string{"big": value})
	original := wspulse.Message{Event: "bulk", Payload: payload}

	data, err := wspulse.JSONCodec.Encode(original)
	require.NoError(t, err)
	decoded, err := wspulse.JSONCodec.Decode(data)
	require.NoError(t, err)
	assert.Equal(t, original.Payload, decoded.Payload, "large payload roundtrip mismatch")
}

// ---- Benchmarks -------------------------------------------------------------

func BenchmarkJSONCodec_Encode(b *testing.B) {
	msg := wspulse.Message{
		Event:   "msg",
		Payload: []byte(`{"user":"alice","text":"hello"}`),
	}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = wspulse.JSONCodec.Encode(msg)
	}
}

func BenchmarkJSONCodec_Decode(b *testing.B) {
	data := []byte(`{"event":"msg","payload":{"user":"alice","text":"hello"}}`)
	b.ReportAllocs()
	for b.Loop() {
		_, _ = wspulse.JSONCodec.Decode(data)
	}
}

func BenchmarkJSONCodec_Roundtrip(b *testing.B) {
	msg := wspulse.Message{
		Event:   "msg",
		Payload: []byte(`{"user":"alice","text":"hello"}`),
	}
	b.ReportAllocs()
	for b.Loop() {
		data, _ := wspulse.JSONCodec.Encode(msg)
		_, _ = wspulse.JSONCodec.Decode(data)
	}
}

// ---- Helpers ----------------------------------------------------------------

func assertJSONString(t *testing.T, m map[string]json.RawMessage, key, want string) {
	t.Helper()
	raw, ok := m[key]
	if !assert.True(t, ok, "key %q not found in JSON", key) {
		return
	}
	var got string
	if !assert.NoError(t, json.Unmarshal(raw, &got), "key %q: failed to unmarshal as string", key) {
		return
	}
	assert.Equal(t, want, got, "key %q", key)
}
