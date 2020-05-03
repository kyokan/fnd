package wire

import "testing"

func TestPeerReq_Encoding(t *testing.T) {
	peerReq := &PeerReq{}
	testMessageEncoding(t, "peer_req", peerReq, &PeerReq{})
}
