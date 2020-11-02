package rpc

import (
	"context"
	"encoding/hex"
	apiv1 "fnd/rpc/v1"
	"io"
)

type Peer struct {
	ID          string
	IP          string
	Banned      bool
	Whitelisted bool
	Connected   bool
	TxBytes     uint64
	RxBytes     uint64
}

func ListPeers(client apiv1.Footnotev1Client) ([]*Peer, error) {
	return ListPeersContext(context.Background(), client)
}

func ListPeersContext(ctx context.Context, client apiv1.Footnotev1Client) ([]*Peer, error) {
	stream, err := client.ListPeers(ctx, &apiv1.ListPeersReq{})
	if err != nil {
		return nil, err
	}
	defer stream.CloseSend()

	var peers []*Peer
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		peers = append(peers, &Peer{
			ID:          hex.EncodeToString(res.PeerID),
			IP:          res.Ip,
			Banned:      res.Banned,
			Whitelisted: res.Whitelisted,
			Connected:   res.Connected,
			TxBytes:     res.TxBytes,
			RxBytes:     res.RxBytes,
		})
	}
	return peers, nil
}

func BanPeer(client apiv1.Footnotev1Client, ip string, duration int) error {
	return BanPeerContext(context.Background(), client, ip, duration)
}

func BanPeerContext(ctx context.Context, client apiv1.Footnotev1Client, ip string, duration int) error {
	_, err := client.BanPeer(ctx, &apiv1.BanPeerReq{
		Ip:         ip,
		DurationMS: uint32(duration),
	})
	return err
}

func UnbanPeer(client apiv1.Footnotev1Client, ip string) error {
	return UnbanPeerContext(context.Background(), client, ip)
}

func UnbanPeerContext(ctx context.Context, client apiv1.Footnotev1Client, ip string) error {
	_, err := client.UnbanPeer(ctx, &apiv1.UnbanPeerReq{
		Ip: ip,
	})
	return err
}

type Status struct {
	PeerID      string
	PeerCount   int
	HeaderCount int
	TxBytes     uint64
	RxBytes     uint64
}

func GetStatus(client apiv1.Footnotev1Client) (*Status, error) {
	return GetStatusContext(context.Background(), client)
}

func GetStatusContext(ctx context.Context, client apiv1.Footnotev1Client) (*Status, error) {
	res, err := client.GetStatus(ctx, &apiv1.Empty{})
	if err != nil {
		return nil, err
	}

	return &Status{
		PeerID:      hex.EncodeToString(res.PeerID),
		PeerCount:   int(res.PeerCount),
		HeaderCount: int(res.HeaderCount),
		TxBytes:     res.TxBytes,
		RxBytes:     res.RxBytes,
	}, nil
}

func AddPeer(client apiv1.Footnotev1Client, peerID string, ip string, verify bool) error {
	return AddPeerContext(context.Background(), client, peerID, ip, verify)
}

func AddPeerContext(ctx context.Context, client apiv1.Footnotev1Client, peerID string, ip string, verify bool) error {
	pIDBytes, err := hex.DecodeString(peerID)
	if err != nil {
		return err
	}

	_, err = client.AddPeer(ctx, &apiv1.AddPeerReq{
		PeerID:       pIDBytes,
		Ip:           ip,
		VerifyPeerID: verify,
	})
	return err
}
