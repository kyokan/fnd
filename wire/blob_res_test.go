package wire

import (
	"testing"
)

func TestBlobRes_Encoding(t *testing.T) {
	blobRes := &BlobRes{
		Name:            "testname.",
		EpochHeight:     0,
		PayloadPosition: 0,
		PrevHash:        fixedHash,
		ReservedRoot:    fixedHash,
		Payload:         nil,
	}

	testMessageEncoding(t, "blob_res", blobRes, &BlobRes{})
}
