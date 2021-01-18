package protocol

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseBanListVersion(t *testing.T) {
	invalidTests := []struct {
		in  string
		err string
	}{
		{"", "colon-separated components"},
		{"whatever", "colon-separated components"},
		{":", "start with FNBAN"},
		{"FNBAN", "colon-separated components"},
		{"FNBAN:", "end with v followed by a digit"},
		{"FNBAN:beep", "end with v followed by a digit"},
		{"FNBAN:1", "end with v followed by a digit"},
	}

	for _, test := range invalidTests {
		ver, err := ParseBanListVersion(test.in)
		require.Equal(t, 0, ver)
		require.Error(t, err)
		require.Contains(t, err.Error(), test.err)
	}

	validTests := []struct {
		in  string
		ver int
	}{
		{"FNBAN:v0", 0},
		{"FNBAN:v1", 1},
		{"FNBAN:v10", 10},
	}

	for _, test := range validTests {
		ver, err := ParseBanListVersion(test.in)
		require.NoError(t, err)
		require.Equal(t, test.ver, ver)
	}
}

func TestReadBanList(t *testing.T) {
	invalidTests := []struct {
		in  string
		err string
	}{
		{
			"",
			"must start with version line",
		},
		{
			"FNBAN:",
			"v followed by a digit",
		},
		{
			"FNBAN:v1\n-------.",
			"start with a hyphen",
		},
		{
			"FNBAN:v0\nhonk",
			"unsupported ban list version",
		},
	}

	for _, test := range invalidTests {
		names, err := ReadBanList(bytes.NewReader([]byte(test.in)))
		require.Nil(t, names)
		require.Error(t, err)
		require.Contains(t, err.Error(), test.err)
	}

	validTests := []struct {
		in  string
		out []string
	}{
		{
			"FNBAN:v1\nwar\nis\npeace",
			[]string{
				"war",
				"is",
				"peace",
			},
		},
		{
			"FNBAN:v1",
			[]string{},
		},
		{
			"FNBAN:v1\n",
			[]string{},
		},
		{
			"FNBAN:v1\n    test2   \n",
			[]string{
				"test2",
			},
		},
		{
			"FNBAN:v1\ntest2   \n",
			[]string{
				"test2",
			},
		},
		{
			"FNBAN:v1\n\ttest2   \nhello",
			[]string{
				"test2",
				"hello",
			},
		},
	}

	for _, test := range validTests {
		names, err := ReadBanList(bytes.NewReader([]byte(test.in)))
		require.NoError(t, err)
		require.Equal(t, len(test.out), len(names))
		for i := 0; i < len(names); i++ {
			require.Equal(t, test.out[i], names[i])
		}
	}
}

func TestFetchListFile(t *testing.T) {
	names, err := FetchListFile("http://sprunge.us/GgRLMw")
	require.NoError(t, err)
	require.Equal(t, 3, len(names))
	require.Equal(t, names[0], "testname")
	require.Equal(t, names[1], "another-test-name")
	require.Equal(t, names[2], "hellothere")
}
