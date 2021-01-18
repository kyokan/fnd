package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrentEpoch(t *testing.T) {
	epoch := CurrentEpoch("bazinga")
	assert.Equal(t, uint16(54), epoch)
}
