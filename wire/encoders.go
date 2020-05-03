package wire

import (
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"io"
	"net"
)

type PublicKeyEncoder struct {
	PublicKey *btcec.PublicKey
}

func (p *PublicKeyEncoder) Encode(w io.Writer) error {
	_, err := w.Write(p.PublicKey.SerializeCompressed())
	return err
}

func (p *PublicKeyEncoder) Decode(r io.Reader) error {
	buf := make([]byte, 33, 33)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	key, err := btcec.ParsePubKey(buf, btcec.S256())
	if err != nil {
		return err
	}
	p.PublicKey = key
	return nil
}

type IPEncoder net.IP

func (i IPEncoder) Encode(w io.Writer) error {
	to16 := net.IP(i).To16()
	if to16 == nil {
		return errors.New("not an IP address")
	}
	_, err := w.Write(to16)
	return err
}

func (i *IPEncoder) Decode(r io.Reader) error {
	buf := make([]byte, 16, 16)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	ip := net.IP(buf)
	*i = IPEncoder(ip)
	return nil
}
