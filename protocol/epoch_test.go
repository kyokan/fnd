package protocol

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCurrentEpoch(t *testing.T) {
	now = func() time.Time { return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC) }
	epoch := CurrentEpoch("bazinga")
	assert.Equal(t, uint16(52), epoch)
}
