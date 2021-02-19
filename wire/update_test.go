package wire

import (
	"testing"
)

func TestUpdate_Encoding(t *testing.T) {
	update := &Update{
		Name:        "testname",
		EpochHeight: fixedEpochHeight,
		SectorSize:  fixedSectorSize,
	}

	testMessageEncoding(t, "update", update, &Update{})
}
