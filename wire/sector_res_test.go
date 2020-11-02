package wire

import (
	"fnd/blob"
	"testing"
)

func TestSectorRes_Encoding(t *testing.T) {
	sectorRes := &SectorRes{
		Name:     "testname.",
		SectorID: 16,
		Sector:   blob.Sector{},
	}

	testMessageEncoding(t, "sector_res", sectorRes, &SectorRes{})
}
