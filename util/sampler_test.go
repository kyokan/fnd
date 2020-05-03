package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandomIndices(t *testing.T) {
	assert.Nil(t, SampleIndices(0, 10))
	assert.Equal(t, []int{0, 1}, SampleIndices(2, 10))

	indices := SampleIndices(3, 2)
	assert.Equal(t, 2, len(indices))
	uniqMap := make(map[int]bool)
	for _, i := range indices {
		uniqMap[i] = true
	}
	assert.Equal(t, 2, len(uniqMap))

	indices = SampleIndices(100, 10)
	assert.Equal(t, 10, len(indices))
	uniqMap = make(map[int]bool)
	for _, i := range indices {
		uniqMap[i] = true
	}
	assert.Equal(t, 10, len(uniqMap))
}
