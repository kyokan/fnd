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
		{":", "start with DDRPBAN"},
		{"DDRPBAN", "colon-separated components"},
		{"DDRPBAN:", "end with v followed by a digit"},
		{"DDRPBAN:beep", "end with v followed by a digit"},
		{"DDRPBAN:1", "end with v followed by a digit"},
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
		{"DDRPBAN:v0", 0},
		{"DDRPBAN:v1", 1},
		{"DDRPBAN:v10", 10},
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
			"DDRPBAN:",
			"v followed by a digit",
		},
		{
			"DDRPBAN:v1\n-------.",
			"start with a hyphen",
		},
		{
			"DDRPBAN:v0\nhonk",
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
			"DDRPBAN:v1\nwar\nis\npeace",
			[]string{
				"war",
				"is",
				"peace",
			},
		},
		{
			"DDRPBAN:v1",
			[]string{},
		},
		{
			"DDRPBAN:v1\n",
			[]string{},
		},
		{
			"DDRPBAN:v1\n    test2   \n",
			[]string{
				"test2",
			},
		},
		{
			"DDRPBAN:v1\ntest2   \n",
			[]string{
				"test2",
			},
		},
		{
			"DDRPBAN:v1\n\ttest2   \nhello",
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
	names, err := FetchListFile("https://gist.githubusercontent.com/mslipper/fd68d2eea1fbb0c435924d71c3152f19/raw/ccfa9fff79ce805c149ebdfff0dfa938f408fa52/banlist")
	require.NoError(t, err)
	require.Equal(t, 3, len(names))
	require.Equal(t, names[0], "testname")
	require.Equal(t, names[1], "another-test-name")
	require.Equal(t, names[2], "hellothere")
}
