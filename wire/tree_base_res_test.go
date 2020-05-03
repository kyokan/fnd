package wire

import (
	"ddrp/blob"
	"testing"
)

func TestTreeBaseRes_Encoding(t *testing.T) {
	treeBaseRes := &TreeBaseRes{
		Name:       "testname.",
		MerkleBase: blob.ZeroMerkleBase,
	}

	testMessageEncoding(t, "tree_base_res", treeBaseRes, &TreeBaseRes{})
}
