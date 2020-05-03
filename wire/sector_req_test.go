package wire

import (
	"testing"
)

func TestSectorReq_Encoding(t *testing.T) {
	sectorReq := &SectorReq{
		Name:     "testname.",
		SectorID: 16,
	}

	testMessageEncoding(t, "sector_req", sectorReq, &SectorReq{})
}
