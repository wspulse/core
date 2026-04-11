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
	assert.Equal(t, wspulse.StatusCode(1000), wspulse.StatusNormalClosure)
	assert.Equal(t, wspulse.StatusCode(1001), wspulse.StatusGoingAway)
	assert.Equal(t, wspulse.StatusCode(1006), wspulse.StatusAbnormalClosure)
}
