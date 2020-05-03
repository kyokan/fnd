package wire

import (
	"ddrp/crypto"
	"github.com/ddrp-org/dwire"
	"io"
	"time"
)

type Update struct {
	HashCacher

	Name         string
	Timestamp    time.Time
	MerkleRoot   crypto.Hash
	ReservedRoot crypto.Hash
	Signature    crypto.Signature
}

var _ Message = (*Update)(nil)

func (u *Update) MsgType() MessageType {
	return MessageTypeUpdate
}

func (u *Update) Equals(other Message) bool {
	cast, ok := other.(*Update)
	if !ok {
		return false
	}

	return u.Name == cast.Name &&
		u.Timestamp.Equal(cast.Timestamp) &&
		u.MerkleRoot == cast.MerkleRoot &&
		u.ReservedRoot == cast.ReservedRoot &&
		u.Signature == cast.Signature
}

func (u *Update) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		u.Name,
		u.Timestamp,
		u.MerkleRoot,
		u.ReservedRoot,
		u.Signature,
	)
}

func (u *Update) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&u.Name,
		&u.Timestamp,
		&u.MerkleRoot,
		&u.ReservedRoot,
		&u.Signature,
	)
}

func (u *Update) Hash() (crypto.Hash, error) {
	return u.HashCacher.Hash(u)
}
