package store

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrefixer(t *testing.T) {
	base := Prefixer("foo")

	tests := []struct {
		in  []byte
		out string
	}{
		{
			base("bar"),
			"foo/bar",
		},
		{
			base(),
			"foo",
		},
		{
			base(""),
			"foo/",
		},
	}
	for _, tt := range tests {
		require.Equal(t, tt.out, string(tt.in))
	}
}
