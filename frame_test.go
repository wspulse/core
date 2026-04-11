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

func TestFrame_ZeroValue_HasEmptyFields(t *testing.T) {
	var f wspulse.Frame
	assert.Empty(t, f.Event)
	assert.Nil(t, f.Payload)
}

func TestFrame_FieldAssignment(t *testing.T) {
	payload := []byte(`{"key":"val"}`)
	f := wspulse.Frame{Event: "msg", Payload: payload}
	assert.Equal(t, "msg", f.Event)
	assert.Equal(t, payload, f.Payload)
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

func TestMessageTypeConstants(t *testing.T) {
	assert.Equal(t, wspulse.MessageType(1), wspulse.MessageText)
	assert.Equal(t, wspulse.MessageType(2), wspulse.MessageBinary)
}

func TestWireFrame_EmptyPayload_OmittedFromJSON(t *testing.T) {
	f := wspulse.Frame{Event: "ping"}
	data, err := wspulse.JSONCodec.Encode(f)
	require.NoError(t, err)
	var m map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &m))
	assert.NotContains(t, m, "payload")
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
	assert.Equal(t, byte('X'), original.Payload[0],
		"expected shallow copy: modifying copy's Payload should affect original")
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
	assert.NotEqual(t, byte('X'), original.Payload[0],
		"explicit copy should make Payload independent")
}
