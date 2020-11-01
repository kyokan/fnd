package primitives

import (
	"fnd.localhost/handshake/encoding"
	"golang.org/x/crypto/blake2b"
	"io"
)

type Transaction struct {
	Version   uint32
	Inputs    []*Input
	Outputs   []*Output
	Locktime  uint32
	Witnesses []*Witness
}

func (t *Transaction) ID() []byte {
	h, _ := blake2b.New256(nil)
	if err := t.EncodeNoWitnesses(h); err != nil {
		panic(err)
	}
	return h.Sum(nil)
}

func (t *Transaction) Encode(w io.Writer) error {
	if err := t.EncodeNoWitnesses(w); err != nil {
		return err
	}
	for _, witness := range t.Witnesses {
		if err := witness.Encode(w); err != nil {
			return err
		}
	}
	return nil
}

func (t *Transaction) EncodeNoWitnesses(w io.Writer) error {
	if err := encoding.WriteUint32(w, t.Version); err != nil {
		return err
	}
	if err := encoding.WriteVarint(w, uint64(len(t.Inputs))); err != nil {
		return err
	}
	for _, input := range t.Inputs {
		if err := input.Encode(w); err != nil {
			return err
		}
	}
	if err := encoding.WriteVarint(w, uint64(len(t.Outputs))); err != nil {
		return err
	}
	for _, output := range t.Outputs {
		if err := output.Encode(w); err != nil {
			return err
		}
	}
	if err := encoding.WriteUint32(w, t.Locktime); err != nil {
		return err
	}
	return nil
}

func (t *Transaction) Decode(r io.Reader) error {
	version, err := encoding.ReadUint32(r)
	if err != nil {
		return err
	}
	inCount, err := encoding.ReadVarint(r)
	if err != nil {
		return err
	}
	var inputs []*Input
	for i := 0; i < int(inCount); i++ {
		input := new(Input)
		if err := input.Decode(r); err != nil {
			return err
		}
		inputs = append(inputs, input)
	}
	outCount, err := encoding.ReadVarint(r)
	if err != nil {
		return err
	}
	var outputs []*Output
	for i := 0; i < int(outCount); i++ {
		output := new(Output)
		if err := output.Decode(r); err != nil {
			return err
		}
		outputs = append(outputs, output)
	}
	locktime, err := encoding.ReadUint32(r)
	if err != nil {
		return err
	}
	var witnesses []*Witness
	for i := 0; i < int(inCount); i++ {
		witness := new(Witness)
		if err := witness.Decode(r); err != nil {
			return err
		}
		witnesses = append(witnesses, witness)
	}
	t.Version = version
	t.Inputs = inputs
	t.Outputs = outputs
	t.Locktime = locktime
	t.Witnesses = witnesses
	return nil
}
