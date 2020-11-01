package dns

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func TestResource_Golden(t *testing.T) {
	tests := []struct {
		name   string
		infile string
	}{
		{
			"proofofconcept name update in block 8578",
			"proofofconcept_4b131b575145a6d0b44654241e89c02a9e316e1acac0ad53edd4d8bb7af3ce8f.bin",
		},
		{
			"ix name update in block 8293",
			"ix_d74fad18bda1e83d2405aac3ea260513bec2cc71e39ce87d24b76fbe6a911c9c.bin",
		},
		{
			"lifelong name update in block 10162",
			"lifelong_1d0f8de2757488cbd59bea7b8f7c7ad5aa9ebd6459631e801a041062338a8630.bin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expData, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s", tt.infile))
			require.NoError(t, err)
			resource := new(Resource)
			require.NoError(t, resource.Decode(bytes.NewReader(expData)))
			actData := new(bytes.Buffer)
			require.NoError(t, resource.Encode(actData))
			require.EqualValues(t, expData, actData.Bytes())
		})
	}
}
