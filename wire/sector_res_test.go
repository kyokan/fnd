package wire

import (
	"github.com/ddrp-org/ddrp/blob"
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
