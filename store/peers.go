package store

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math"
	"time"
)

type Peer struct {
	ID                  crypto.Hash
	IP                  string
	LastSeen            time.Time
	Verify              bool
	InboundBannedUntil  time.Time
	OutboundBannedUntil time.Time
}

func (p *Peer) MarshalJSON() ([]byte, error) {
	out := struct {
		PeerID   string    `json:"peer_id"`
		IP       string    `json:"ip"`
		LastSeen time.Time `json:"last_seen"`
		Verify   bool      `json:"verify"`
	}{
		p.ID.String(),
		p.IP,
		p.LastSeen,
		p.Verify,
	}

	return json.Marshal(out)
}

func (p *Peer) UnmarshalJSON(b []byte) error {
	out := &struct {
		PeerID   string    `json:"peer_id"`
		IP       string    `json:"ip"`
		LastSeen time.Time `json:"last_seen"`
		Verify   bool      `json:"verify"`
	}{}
	if err := json.Unmarshal(b, out); err != nil {
		return err
	}
	idB, err := hex.DecodeString(out.PeerID)
	if err != nil {
		return err
	}
	hash, err := crypto.NewHashFromBytes(idB)
	if err != nil {
		return err
	}

	p.ID = hash
	p.IP = out.IP
	p.LastSeen = out.LastSeen
	p.Verify = out.Verify
	return nil
}

func (p *Peer) IsBanned() bool {
	now := time.Now()
	return p.InboundBannedUntil.After(now) || p.OutboundBannedUntil.After(now)
}

var (
	peersPrefix      = Prefixer("peers")
	peerDataPrefix   = Prefixer(string(peersPrefix("peer")))
	peerInBanPrefix  = Prefixer(string(peersPrefix("inbound-ban")))
	peerOutBanPrefix = Prefixer(string(peersPrefix("outbound-ban")))
)

func SetPeer(db *leveldb.DB, id crypto.Hash, ip string, verify bool) error {
	return WithTx(db, func(tx *leveldb.Transaction) error {
		return SetPeerTx(tx, id, ip, verify)
	})
}

func SetPeerTx(batch *leveldb.Transaction, id crypto.Hash, ip string, verify bool) error {
	err := batch.Put(peerDataPrefix(id.String()), mustMarshalJSON(&Peer{
		ID:       id,
		IP:       ip,
		LastSeen: time.Now(),
		Verify:   verify,
	}), nil)
	if err != nil {
		return errors.Wrap(err, "error writing peer")
	}
	return nil
}

type PeerStream struct {
	includeBanned bool
	db            *leveldb.DB
	iter          iterator.Iterator
}

func (ps *PeerStream) Next() (*Peer, error) {
	if !ps.iter.Next() {
		return nil, nil
	}

	peer := new(Peer)
	mustUnmarshalJSON(ps.iter.Value(), peer)
	inRes, err := ps.db.Get(peerInBanPrefix(peer.IP), nil)
	if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		return nil, errors.Wrap(err, "error getting inbound ban state during stream")
	}
	outRes, err := ps.db.Get(peerOutBanPrefix(peer.IP), nil)
	if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		return nil, errors.Wrap(err, "error getting outbound ban state during stream")
	}
	peer.InboundBannedUntil = mustDecodeTime(inRes)
	peer.OutboundBannedUntil = mustDecodeTime(outRes)

	now := time.Now()
	if !ps.includeBanned && (peer.InboundBannedUntil.After(now) || peer.OutboundBannedUntil.After(now)) {
		return ps.Next()
	}
	return peer, nil
}

func (ps *PeerStream) Close() error {
	ps.iter.Release()
	return ps.iter.Error()
}

func StreamPeers(db *leveldb.DB, includeBanned bool) (*PeerStream, error) {
	iter := db.NewIterator(util.BytesPrefix(peerDataPrefix()), nil)
	return &PeerStream{
		includeBanned: includeBanned,
		db:            db,
		iter:          iter,
	}, nil
}

func BanOutboundPeer(db *leveldb.DB, ip string, duration time.Duration) error {
	return WithTx(db, func(tx *leveldb.Transaction) error {
		return BanOutboundPeerTx(tx, ip, duration)
	})
}

func BanOutboundPeerTx(tx *leveldb.Transaction, ip string, duration time.Duration) error {
	if err := tx.Put(peerOutBanPrefix(ip), encodeTime(time.Now().Add(duration)), nil); err != nil {
		return errors.Wrap(err, "error putting outbound ban")
	}
	return nil
}

func BanInboundPeer(db *leveldb.DB, ip string, duration time.Duration) error {
	return WithTx(db, func(tx *leveldb.Transaction) error {
		return BanInboundPeerTx(tx, ip, duration)
	})
}

func BanInboundPeerTx(tx *leveldb.Transaction, ip string, duration time.Duration) error {
	if err := tx.Put(peerInBanPrefix(ip), encodeTime(time.Now().Add(duration)), nil); err != nil {
		return errors.Wrap(err, "erorr putting inbound ban")
	}
	return nil
}

func UnbanOutboundPeerTx(tx *leveldb.Transaction, ip string) error {
	k := peerOutBanPrefix(ip)
	has, err := tx.Has(k, nil)
	if err != nil {
		return errors.Wrap(err, "error checking for inbound ban existence")
	}
	if !has {
		return nil
	}
	if err := tx.Delete(k, nil); err != nil {
		return errors.Wrap(err, "error deleting outbound ban")
	}
	return nil
}

func UnbanInboundPeerTx(tx *leveldb.Transaction, ip string) error {
	k := peerInBanPrefix(ip)
	has, err := tx.Has(k, nil)
	if err != nil {
		return errors.Wrap(err, "error checking for inbound ban existence")
	}
	if !has {
		return nil
	}
	if err := tx.Delete(k, nil); err != nil {
		return errors.Wrap(err, "error deleting outbound ban")
	}
	return nil
}

func IsBanned(db *leveldb.DB, ip string) (bool, bool, error) {
	now := time.Now()
	inRes, err := db.Get(peerInBanPrefix(ip), nil)
	if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		return false, false, errors.Wrap(err, "error getting inbound ban state")
	}
	outRes, err := db.Get(peerOutBanPrefix(ip), nil)
	if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		return false, false, errors.Wrap(err, "error getting outbound ban state")
	}

	return mustDecodeTime(inRes).After(now), mustDecodeTime(outRes).After(now), nil
}

func TruncatePeerStore(db *leveldb.DB) error {
	err := WithTx(db, func(tx *leveldb.Transaction) error {
		iter := tx.NewIterator(util.BytesPrefix(peersPrefix()), nil)
		for iter.Next() {
			if err := tx.Delete(iter.Key(), nil); err != nil {
				return errors.Wrap(err, "error deleting peer store key")
			}
		}
		iter.Release()
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "error truncating peer store")
	}
	return nil
}

func mustMarshalJSON(in interface{}) []byte {
	out, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	return out
}

func mustUnmarshalJSON(data []byte, in interface{}) {
	if err := json.Unmarshal(data, in); err != nil {
		panic(err)
	}
}

func encodeTime(t time.Time) []byte {
	buf := make([]byte, 8, 8)
	binary.BigEndian.PutUint64(buf, uint64(t.Unix()))
	return buf
}

func mustDecodeTime(buf []byte) time.Time {
	if buf == nil {
		return time.Time{}
	}
	out := binary.BigEndian.Uint64(buf)
	if out > math.MaxInt64 {
		panic("overflow")
	}
	return time.Unix(int64(out), 0)
}
