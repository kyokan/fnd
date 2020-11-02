package primitives

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fnd.localhost/handshake/dns"
	"fnd.localhost/handshake/encoding"
	"io"
)

const (
	CovenantNone uint8 = iota
	CovenantClaim
	CovenantOpen
	CovenantBid
	CovenantReveal
	CovenantRedeem
	CovenantRegister
	CovenantUpdate
	CovenantRenew
	CovenantTransfer
	CovenantFinalize
	CovenantRevoke

	MaxScriptStack = 1000
)

type Covenant struct {
	Type  uint8
	Items [][]byte
}

func (c *Covenant) Encode(w io.Writer) error {
	if err := encoding.WriteUint8(w, c.Type); err != nil {
		return err
	}
	if err := encoding.WriteVarint(w, uint64(len(c.Items))); err != nil {
		return err
	}
	for _, item := range c.Items {
		if err := encoding.WriteVarBytes(w, item); err != nil {
			return err
		}
	}
	return nil
}

func (c *Covenant) Decode(r io.Reader) error {
	typ, err := encoding.ReadUint8(r)
	if err != nil {
		return err
	}
	if typ > CovenantRevoke {
		return errors.New("invalid covenant type")
	}
	count, err := encoding.ReadVarint(r)
	if err != nil {
		return err
	}
	if count > MaxScriptStack {
		return errors.New("too many covenant items")
	}

	var items [][]byte
	for i := 0; i < int(count); i++ {
		item, err := encoding.ReadVarBytes(r)
		if err != nil {
			return err
		}
		items = append(items, item)
	}
	c.Type = typ
	c.Items = items
	return nil
}

type Open struct {
	NameHash []byte
	Reserved uint32
	Name     string
}

func OpenFromCovenant(cov *Covenant) (*Open, error) {
	if cov.Type != CovenantOpen {
		return nil, errors.New("covenant is not an open covenant")
	}
	reserved := binary.LittleEndian.Uint32(cov.Items[1])
	return &Open{
		NameHash: cov.Items[0],
		Reserved: reserved,
		Name:     string(cov.Items[2]),
	}, nil
}

type Bid struct {
	NameHash []byte
	Start    uint32
	Name     string
	Blind    []byte
}

func BidFromCovenant(cov *Covenant) (*Bid, error) {
	if cov.Type != CovenantBid {
		return nil, errors.New("covenant is not a bid covenant")
	}
	return &Bid{
		NameHash: cov.Items[0],
		Start:    binary.LittleEndian.Uint32(cov.Items[1]),
		Name:     string(cov.Items[2]),
		Blind:    cov.Items[3],
	}, nil
}

type Reveal struct {
	NameHash []byte
	Height   uint32
	Nonce    []byte
}

func RevealFromCovenant(cov *Covenant) (*Reveal, error) {
	if cov.Type != CovenantReveal {
		return nil, errors.New("covenant is not a reveal covenant")
	}
	return &Reveal{
		NameHash: cov.Items[0],
		Height:   binary.LittleEndian.Uint32(cov.Items[1]),
		Nonce:    cov.Items[2],
	}, nil
}

type Register struct {
	NameHash         []byte
	Height           uint32
	Resource         *dns.Resource
	RenewalBlockHash []byte
}

func RegisterFromCovenant(cov *Covenant) (*Register, error) {
	if cov.Type != CovenantRegister {
		return nil, errors.New("covenant is not a register covenant")
	}
	resourceB := cov.Items[2]
	if len(resourceB) > dns.MaxResourceSize {
		return nil, errors.New("resource too large")
	}
	var resource *dns.Resource
	if len(resourceB) > 0 {
		resource = new(dns.Resource)
		if err := resource.Decode(bytes.NewReader(resourceB)); err != nil {
			return nil, err
		}
	}
	return &Register{
		NameHash:         cov.Items[0],
		Height:           binary.LittleEndian.Uint32(cov.Items[1]),
		Resource:         resource,
		RenewalBlockHash: cov.Items[3],
	}, nil
}

type Redeem struct {
	NameHash []byte
	Height   uint32
}

func RedeemFromCovenant(cov *Covenant) (*Redeem, error) {
	if cov.Type != CovenantRedeem {
		return nil, errors.New("covenant is not a redeem covenant")
	}
	return &Redeem{
		NameHash: cov.Items[0],
		Height:   binary.LittleEndian.Uint32(cov.Items[1]),
	}, nil
}

type Update struct {
	NameHash []byte
	Height   uint32
	Resource *dns.Resource
}

func UpdateFromCovenant(cov *Covenant) (*Update, error) {
	if cov.Type != CovenantUpdate {
		return nil, errors.New("covenant is not an update covenant")
	}
	resourceB := cov.Items[2]
	if len(resourceB) > dns.MaxResourceSize {
		return nil, errors.New("resource too large")
	}
	var resource *dns.Resource
	if len(resourceB) > 0 {
		resource = new(dns.Resource)
		if err := resource.Decode(bytes.NewReader(resourceB)); err != nil {
			return nil, err
		}
	}
	return &Update{
		NameHash: cov.Items[0],
		Height:   binary.LittleEndian.Uint32(cov.Items[1]),
		Resource: resource,
	}, nil
}

type Renewal struct {
	NameHash         []byte
	Height           uint32
	RenewalBlockHash []byte
}

func RenewalFromCovenant(cov *Covenant) (*Renewal, error) {
	if cov.Type != CovenantRenew {
		return nil, errors.New("covenant is not a renewal covenant")
	}
	return &Renewal{
		NameHash:         cov.Items[0],
		Height:           binary.LittleEndian.Uint32(cov.Items[1]),
		RenewalBlockHash: cov.Items[2],
	}, nil
}

type Transfer struct {
	NameHash []byte
	Height   uint32
	Address  *Address
}

func TransferFromCovenant(cov *Covenant) (*Transfer, error) {
	if cov.Type != CovenantTransfer {
		return nil, errors.New("covenant is not a transfer covenant")
	}
	return &Transfer{
		NameHash: cov.Items[0],
		Height:   binary.LittleEndian.Uint32(cov.Items[1]),
		Address: &Address{
			Version: cov.Items[2][0],
			Hash:    cov.Items[3],
		},
	}, nil
}

type Finalize struct {
	NameHash         []byte
	Height           uint32
	Name             string
	Flags            uint8
	Claimed          uint32
	Renewals         uint32
	RenewalBlockHash []byte
}

func FinalizeFromCovenant(cov *Covenant) (*Finalize, error) {
	if cov.Type != CovenantFinalize {
		return nil, errors.New("covenant is not a finalize covenant")
	}
	return &Finalize{
		NameHash:         cov.Items[0],
		Height:           binary.LittleEndian.Uint32(cov.Items[1]),
		Name:             string(cov.Items[2]),
		Flags:            cov.Items[3][0],
		Claimed:          binary.LittleEndian.Uint32(cov.Items[4]),
		Renewals:         binary.LittleEndian.Uint32(cov.Items[5]),
		RenewalBlockHash: cov.Items[6],
	}, nil
}
