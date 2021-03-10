package wire

import (
	"testing"
)

func TestBlobReq_Encoding(t *testing.T) {
	blobReq := &BlobReq{
		Name:        "testname.",
		EpochHeight: 0,
		SectorSize:  1,
	}

	testMessageEncoding(t, "blob_req", blobReq, &BlobReq{})
}
