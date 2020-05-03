package wire

import "testing"

func TestNilUpdate_Encoding(t *testing.T) {
	nilUpdate := NewNilUpdate("testname.")
	testMessageEncoding(t, "nil_update", nilUpdate, &NilUpdate{})
}
