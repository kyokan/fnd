package blob

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPathifyPubHash(t *testing.T) {
	path := PathifyName("bazinga")
	assert.Equal(t, "f25b/57d9/eebc/274f/49cc/eafd/aacf/9a4a/d6c3/17a2/8a80/37ae/9060/ab30/9d63/4aea_blob", path)
}
