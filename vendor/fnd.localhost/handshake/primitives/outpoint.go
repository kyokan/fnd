package primitives

import (
	"fnd.localhost/handshake/encoding"
	"io"
)

type Outpoint struct {
	Hash  [32]byte
	Index uint32
}

func (o *Outpoint) Encode(w io.Writer) error {
	if _, err := w.Write(o.Hash[:]); err != nil {
		return err
	}
	if err := encoding.WriteUint32(w, o.Index); err != nil {
		return err
	}
	return nil
}

func (o *Outpoint) Decode(r io.Reader) error {
	var hash [32]byte
	if _, err := io.ReadFull(r, hash[:]); err != nil {
		return err
	}
	index, err := encoding.ReadUint32(r)
	if err != nil {
		return err
	}
	o.Hash = hash
	o.Index = index
	return nil
}
