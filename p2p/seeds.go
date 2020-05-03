package p2p

import (
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/pkg/errors"
	"net"
	"strings"
)

type SeedPeer struct {
	ID crypto.Hash
	IP string
}

func ResolveDNSSeeds(domain string) ([]string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, errors.Wrap(err, "error looking up DNS seeds")
	}

	var out []string
	used := make(map[string]bool)
	for _, ip := range ips {
		if used[ip.String()] {
			continue
		}
		to4 := ip.To4()
		if to4 == nil {
			continue
		}
		out = append(out, to4.String())
	}
	return out, nil
}

func ParseSeedPeers(seedPeers []string) ([]SeedPeer, error) {
	if len(seedPeers) == 0 {
		return nil, nil
	}

	var peers []SeedPeer
	for _, seed := range seedPeers {
		pieces := strings.Split(seed, "@")
		if len(pieces) != 2 {
			return nil, errors.New("seed peer must specify both a peer ID and address")
		}
		peerID, err := crypto.NewHashFromHex(pieces[0])
		if err != nil {
			return nil, errors.Wrap(err, "mal-formed peer ID")
		}
		peers = append(peers, SeedPeer{
			ID: peerID,
			IP: pieces[1],
		})
	}

	return peers, nil
}
