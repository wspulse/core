package wspulse_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wspulse "github.com/wspulse/core"
)

func TestMessage_ZeroValue_HasEmptyFields(t *testing.T) {
	var m wspulse.Message
	assert.Empty(t, m.Event)
	assert.Nil(t, m.Payload)
}

func TestMessage_FieldAssignment(t *testing.T) {
	payload := []byte(`{"key":"val"}`)
	m := wspulse.Message{Event: "msg", Payload: payload}
	assert.Equal(t, "msg", m.Event)
	assert.Equal(t, payload, m.Payload)
}

func TestSentinelErrors_AreDistinct(t *testing.T) {
	sentinels := []error{
		wspulse.ErrConnectionClosed,
		wspulse.ErrSendBufferFull,
	}
	for i, a := range sentinels {
		for j, b := range sentinels {
			if i != j {
				require.NotEqual(t, a, b, "errors[%d] and errors[%d] should be distinct", i, j)
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
		assert.NotEmpty(t, tc.err.Error(), "%s.Error() returned empty string", tc.name)
	}
}

func TestWireMessage_EmptyPayload_OmittedFromJSON(t *testing.T) {
	m := wspulse.Message{Event: "ping"}
	data, err := wspulse.JSONCodec.Encode(m)
	require.NoError(t, err)
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))
	assert.NotContains(t, raw, "payload")
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
		assert.True(t, strings.HasPrefix(tc.err.Error(), "wspulse:"),
			"%s.Error() = %q, want prefix %q", tc.name, tc.err.Error(), "wspulse:")
	}
}

func TestSentinelErrors_SupportErrorsIs(t *testing.T) {
	wrapped := errors.Join(
		wspulse.ErrConnectionClosed,
		errors.New("extra context"),
	)
	assert.ErrorIs(t, wrapped, wspulse.ErrConnectionClosed)
}

// ---- Message copy semantics (lifecycle awareness) ---------------------------

func TestMessage_PayloadSharing_AfterCopy(t *testing.T) {
	original := wspulse.Message{
		Event:   "msg",
		Payload: []byte(`{"data":"original"}`),
	}
	copied := original

	// Mutate the copied message's Payload in-place.
	copied.Payload[0] = 'X'

	// Since slices share backing arrays, the original is also affected.
	// This test documents the expected Go behavior so users are aware.
	assert.Equal(t, byte('X'), original.Payload[0],
		"expected shallow copy: modifying copy's Payload should affect original")
}

func TestMessage_IndependentPayload_RequiresExplicitCopy(t *testing.T) {
	original := wspulse.Message{
		Event:   "msg",
		Payload: []byte(`{"data":"safe"}`),
	}

	// Proper deep copy pattern for Message.
	independent := wspulse.Message{
		Event: original.Event,
	}
	independent.Payload = make([]byte, len(original.Payload))
	copy(independent.Payload, original.Payload)

	independent.Payload[0] = 'X'
	assert.NotEqual(t, byte('X'), original.Payload[0],
		"explicit copy should make Payload independent")
}
