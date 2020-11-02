package primitives

import (
	"fnd.localhost/handshake/encoding"
	"io"
)

type Input struct {
	Prevout  *Outpoint
	Sequence uint32
}

func (in *Input) Encode(w io.Writer) error {
	if err := in.Prevout.Encode(w); err != nil {
		return err
	}
	if err := encoding.WriteUint32(w, in.Sequence); err != nil {
		return err
	}
	return nil
}

func (in *Input) Decode(r io.Reader) error {
	prevout := new(Outpoint)
	if err := prevout.Decode(r); err != nil {
		return err
	}
	sequence, err := encoding.ReadUint32(r)
	if err != nil {
		return err
	}
	in.Prevout = prevout
	in.Sequence = sequence
	return nil
}
