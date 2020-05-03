package wire

import "testing"

func TestPing_Encoding(t *testing.T) {
	ping := &Ping{}
	testMessageEncoding(t, "ping", ping, &Ping{})
}
