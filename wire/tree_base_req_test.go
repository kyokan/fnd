package wire

import (
	"testing"
)

func TestTreeBaseReq_Encoding(t *testing.T) {
	treeBaseReq := &TreeBaseReq{
		Name: "testname.",
	}

	testMessageEncoding(t, "tree_base_req", treeBaseReq, &TreeBaseReq{})
}
