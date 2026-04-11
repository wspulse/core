package wspulse_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	wspulse "github.com/wspulse/core"
)

func TestMessageTypeConstants(t *testing.T) {
	assert.Equal(t, wspulse.MessageType(1), wspulse.TextMessage)
	assert.Equal(t, wspulse.MessageType(2), wspulse.BinaryMessage)
}

func TestStatusCodeConstants(t *testing.T) {
	// Standard codes — valid to send in a close frame.
	assert.Equal(t, wspulse.StatusCode(1000), wspulse.StatusNormalClosure)
	assert.Equal(t, wspulse.StatusCode(1001), wspulse.StatusGoingAway)
	assert.Equal(t, wspulse.StatusCode(1002), wspulse.StatusProtocolError)
	assert.Equal(t, wspulse.StatusCode(1003), wspulse.StatusUnsupportedData)
	assert.Equal(t, wspulse.StatusCode(1007), wspulse.StatusInvalidFramePayloadData)
	assert.Equal(t, wspulse.StatusCode(1008), wspulse.StatusPolicyViolation)
	assert.Equal(t, wspulse.StatusCode(1009), wspulse.StatusMessageTooBig)
	assert.Equal(t, wspulse.StatusCode(1010), wspulse.StatusMandatoryExtension)
	assert.Equal(t, wspulse.StatusCode(1011), wspulse.StatusInternalError)

	// Local-only codes — MUST NOT be sent in a close frame (RFC 6455 §7.4.1).
	assert.Equal(t, wspulse.StatusCode(1005), wspulse.StatusNoStatusReceived)
	assert.Equal(t, wspulse.StatusCode(1006), wspulse.StatusAbnormalClosure)
	assert.Equal(t, wspulse.StatusCode(1015), wspulse.StatusTLSHandshake)
}
