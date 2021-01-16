package wire

import (
	"io"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/ddrp-org/ddrp/dwire"
)

type Hello struct {
	HashCacher

	ProtocolVersion uint32
	LocalNonce      [32]byte
	RemoteNonce     [32]byte
	PublicKey       *btcec.PublicKey
	UserAgent       string
}

var _ Message = (*Hello)(nil)

func (h *Hello) MsgType() MessageType {
	return MessageTypeHello
}

func (h *Hello) Equals(other Message) bool {
	cast, ok := other.(*Hello)
	if !ok {
		return false
	}

	return h.ProtocolVersion == cast.ProtocolVersion &&
		h.LocalNonce == cast.LocalNonce &&
		h.RemoteNonce == cast.RemoteNonce &&
		h.PublicKey.IsEqual(cast.PublicKey) &&
		h.UserAgent == cast.UserAgent
}

func (h *Hello) Encode(w io.Writer) error {
	pubEnc := &PublicKeyEncoder{
		PublicKey: h.PublicKey,
	}
	return dwire.EncodeFields(
		w,
		h.ProtocolVersion,
		h.LocalNonce,
		h.RemoteNonce,
		pubEnc,
		h.UserAgent,
	)
}

func (h *Hello) Decode(r io.Reader) error {
	var pubEnc PublicKeyEncoder
	err := dwire.DecodeFields(
		r,
		&h.ProtocolVersion,
		&h.LocalNonce,
		&h.RemoteNonce,
		&pubEnc,
		&h.UserAgent,
	)
	if err != nil {
		return err
	}
	h.PublicKey = pubEnc.PublicKey
	return nil
}

func (h *Hello) Hash() (crypto.Hash, error) {
	return h.HashCacher.Hash(h)
}
