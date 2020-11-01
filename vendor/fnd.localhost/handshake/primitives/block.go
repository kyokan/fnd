package primitives

import (
	"bytes"
	"fnd.localhost/handshake/encoding"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
	"io"
)

const (
	NonceSize = 24
	MaskSize  = 32
)

type Block struct {
	Nonce        uint32
	Time         uint64
	PrevHash     [32]byte
	TreeRoot     [32]byte
	ExtraNonce   [NonceSize]byte
	ReservedRoot [32]byte
	WitnessRoot  [32]byte
	MerkleRoot   [32]byte
	Version      uint32
	Bits         uint32
	Mask         [MaskSize]byte
	Transactions []*Transaction
}

func (b *Block) Hash() []byte {
	leftData := new(bytes.Buffer)
	if err := encoding.WriteUint32(leftData, b.Nonce); err != nil {
		panic(err)
	}
	if err := encoding.WriteUint64(leftData, b.Time); err != nil {
		panic(err)
	}
	leftData.Write(b.padding(20))
	leftData.Write(b.PrevHash[:])
	leftData.Write(b.TreeRoot[:])
	leftData.Write(b.commitHash())
	left := leftData.Bytes()

	leftH, _ := blake2b.New512(nil)
	leftH.Write(left)

	rightH := sha3.New256()
	rightH.Write(left)
	rightH.Write(b.padding(8))

	outH, _ := blake2b.New256(nil)
	outH.Write(leftH.Sum(nil))
	outH.Write(b.padding(32))
	outH.Write(rightH.Sum(nil))
	return outH.Sum(nil)
}

func (b *Block) commitHash() []byte {
	h, _ := blake2b.New256(nil)
	_, _ = h.Write(b.subHash())
	_, _ = h.Write(b.maskHash())
	return h.Sum(nil)
}

func (b *Block) subHash() []byte {
	h, _ := blake2b.New256(nil)
	h.Write(b.ExtraNonce[:])
	h.Write(b.ReservedRoot[:])
	h.Write(b.WitnessRoot[:])
	h.Write(b.MerkleRoot[:])
	_ = encoding.WriteUint32(h, b.Version)
	_ = encoding.WriteUint32(h, b.Bits)
	return h.Sum(nil)
}

func (b *Block) maskHash() []byte {
	h, _ := blake2b.New256(nil)
	h.Write(b.PrevHash[:])
	h.Write(b.Mask[:])
	return h.Sum(nil)
}

func (b *Block) padding(size int) []byte {
	buf := make([]byte, size, size)
	for i := 0; i < len(buf); i++ {
		buf[i] = b.PrevHash[i%32] ^ b.TreeRoot[i%32]
	}
	return buf
}

func (b *Block) Encode(w io.Writer) error {
	if err := encoding.WriteUint32(w, b.Nonce); err != nil {
		return err
	}
	if err := encoding.WriteUint64(w, b.Time); err != nil {
		return err
	}
	if _, err := w.Write(b.PrevHash[:]); err != nil {
		return err
	}
	if _, err := w.Write(b.TreeRoot[:]); err != nil {
		return err
	}
	if _, err := w.Write(b.ExtraNonce[:]); err != nil {
		return err
	}
	if _, err := w.Write(b.ReservedRoot[:]); err != nil {
		return err
	}
	if _, err := w.Write(b.WitnessRoot[:]); err != nil {
		return err
	}
	if _, err := w.Write(b.MerkleRoot[:]); err != nil {
		return err
	}
	if err := encoding.WriteUint32(w, b.Version); err != nil {
		return err
	}
	if err := encoding.WriteUint32(w, b.Bits); err != nil {
		return err
	}
	if _, err := w.Write(b.Mask[:]); err != nil {
		return err
	}
	if err := encoding.WriteVarint(w, uint64(len(b.Transactions))); err != nil {
		return err
	}
	for _, tx := range b.Transactions {
		if err := tx.Encode(w); err != nil {
			return err
		}
	}
	return nil
}

func (b *Block) Decode(r io.Reader) error {
	nonce, err := encoding.ReadUint32(r)
	if err != nil {
		return err
	}
	ts, err := encoding.ReadUint64(r)
	if err != nil {
		return err
	}
	var hash [32]byte
	if _, err := r.Read(hash[:]); err != nil {
		return err
	}
	var treeRoot [32]byte
	if _, err := r.Read(treeRoot[:]); err != nil {
		return err
	}
	var extraNonce [NonceSize]byte
	if _, err := r.Read(extraNonce[:]); err != nil {
		return err
	}
	var reservedRoot [32]byte
	if _, err := r.Read(reservedRoot[:]); err != nil {
		return err
	}
	var witnessRoot [32]byte
	if _, err := r.Read(witnessRoot[:]); err != nil {
		return err
	}
	var merkleRoot [32]byte
	if _, err := r.Read(merkleRoot[:]); err != nil {
		return err
	}
	version, err := encoding.ReadUint32(r)
	if err != nil {
		return err
	}
	bits, err := encoding.ReadUint32(r)
	if err != nil {
		return err
	}
	var mask [MaskSize]byte
	if _, err := r.Read(mask[:]); err != nil {
		return err
	}
	txCount, err := encoding.ReadVarint(r)
	if err != nil {
		return err
	}
	var txs []*Transaction
	for i := 0; i < int(txCount); i++ {
		tx := new(Transaction)
		if err := tx.Decode(r); err != nil {
			return err
		}
		txs = append(txs, tx)
	}
	b.Nonce = nonce
	b.Time = ts
	b.PrevHash = hash
	b.TreeRoot = treeRoot
	b.ExtraNonce = extraNonce
	b.ReservedRoot = reservedRoot
	b.WitnessRoot = witnessRoot
	b.MerkleRoot = merkleRoot
	b.Version = version
	b.Bits = bits
	b.Mask = mask
	b.Transactions = txs
	return nil
}
