package primitives

import (
	"errors"
	"fnd.localhost/handshake/encoding"
	"io"
)

type Witness struct {
	Items [][]byte
}

func (wt *Witness) Encode(w io.Writer) error {
	if err := encoding.WriteVarint(w, uint64(len(wt.Items))); err != nil {
		return err
	}
	for _, item := range wt.Items {
		if err := encoding.WriteVarBytes(w, item); err != nil {
			return err
		}
	}
	return nil
}

func (wt *Witness) Decode(r io.Reader) error {
	count, err := encoding.ReadVarint(r)
	if err != nil {
		return err
	}
	if count > MaxScriptStack {
		return errors.New("too many witness items")
	}

	var items [][]byte
	for i := 0; i < int(count); i++ {
		item, err := encoding.ReadVarBytes(r)
		if err != nil {
			return err
		}
		items = append(items, item)
	}
	wt.Items = items
	return nil
}
