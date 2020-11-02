package primitives

import (
	"fnd.localhost/handshake/encoding"
	"io"
)

type Output struct {
	Value    uint64
	Address  *Address
	Covenant *Covenant
}

func (o *Output) Encode(w io.Writer) error {
	if err := encoding.WriteUint64(w, o.Value); err != nil {
		return err
	}
	if err := o.Address.Encode(w); err != nil {
		return err
	}
	if err := o.Covenant.Encode(w); err != nil {
		return err
	}
	return nil
}

func (o *Output) Decode(r io.Reader) error {
	value, err := encoding.ReadUint64(r)
	if err != nil {
		return err
	}
	address := new(Address)
	if err := address.Decode(r); err != nil {
		return err
	}
	covenant := new(Covenant)
	if err := covenant.Decode(r); err != nil {
		return err
	}
	o.Value = value
	o.Address = address
	o.Covenant = covenant
	return nil
}
