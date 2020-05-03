package wire

import (
	"net"
	"testing"
)

func TestPeerRes_Encoding(t *testing.T) {
	peerRes := &PeerRes{
		Peers: []*Peer{
			{
				IP: net.ParseIP("192.168.0.1"),
				ID: fixedHash,
			},
			{
				IP: net.ParseIP("1.1.1.1"),
				ID: fixedHash,
			},
		},
	}

	testMessageEncoding(t, "peer_res", peerRes, &PeerRes{})
}
