package store

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcd/btcec"
	"fnd/blob"
	"fnd/crypto"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"sync"
	"time"
)

type Header struct {
	Name         string
	Timestamp    time.Time
	MerkleRoot   crypto.Hash
	Signature    crypto.Signature
	ReservedRoot crypto.Hash
	ReceivedAt   time.Time
	Timebank     int
}

func (h *Header) MarshalJSON() ([]byte, error) {
	out := &struct {
		Name         string    `json:"name"`
		Timestamp    time.Time `json:"timestamp"`
		MerkleRoot   string    `json:"merkle_root"`
		Signature    string    `json:"signature"`
		ReservedRoot string    `json:"reserved_root"`
		ReceivedAt   time.Time `json:"received_at"`
		Timebank     int       `json:"timebank"`
	}{
		h.Name,
		h.Timestamp,
		h.MerkleRoot.String(),
		h.Signature.String(),
		h.ReservedRoot.String(),
		h.ReceivedAt,
		h.Timebank,
	}

	return json.Marshal(out)
}

func (h *Header) UnmarshalJSON(b []byte) error {
	in := &struct {
		Name         string    `json:"name"`
		Timestamp    time.Time `json:"timestamp"`
		MerkleRoot   string    `json:"merkle_root"`
		Signature    string    `json:"signature"`
		ReservedRoot string    `json:"reserved_root"`
		ReceivedAt   time.Time `json:"received_at"`
		Timebank     int       `json:"timebank"`
	}{}
	if err := json.Unmarshal(b, in); err != nil {
		return err
	}
	mrB, err := hex.DecodeString(in.MerkleRoot)
	if err != nil {
		return err
	}
	mr, err := crypto.NewHashFromBytes(mrB)
	if err != nil {
		return err
	}
	sigB, err := hex.DecodeString(in.Signature)
	if err != nil {
		return err
	}
	sig, err := crypto.NewSignatureFromBytes(sigB)
	if err != nil {
		return err
	}
	rrB, err := hex.DecodeString(in.ReservedRoot)
	if err != nil {
		return err
	}
	rr, err := crypto.NewHashFromBytes(rrB)
	if err != nil {
		return err
	}

	h.Name = in.Name
	h.Timestamp = in.Timestamp
	h.MerkleRoot = mr
	h.Signature = sig
	h.ReservedRoot = rr
	h.ReceivedAt = in.ReceivedAt
	h.Timebank = in.Timebank
	return nil
}

var (
	headersPrefix          = Prefixer("headers")
	headerCountKey         = Prefixer(string(headersPrefix("count")))()
	headerMerkleBasePrefix = Prefixer(string(headersPrefix("merkle-base")))
	headerDataPrefix       = Prefixer(string(headersPrefix("header")))
)

func GetHeaderCount(db *leveldb.DB) (int, error) {
	res, err := db.Get(headerCountKey, nil)
	if errors.Is(err, leveldb.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrap(err, "error getting header count")
	}
	return mustDecodeInt(res), nil
}

var hCountMu sync.Mutex

func IncrementHeaderCount(tx *leveldb.Transaction) error {
	hCountMu.Lock()
	defer hCountMu.Unlock()
	count, err := tx.Get(headerCountKey, nil)
	if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		return errors.Wrap(err, "error getting header count")
	}
	if err := tx.Put(headerCountKey, mustEncodeInt(mustDecodeInt(count)+1), nil); err != nil {
		return errors.Wrap(err, "error putting header count")
	}
	return nil
}

func GetHeader(db *leveldb.DB, name string) (*Header, error) {
	header := new(Header)
	headerData, err := db.Get(headerDataPrefix(name), nil)
	if err != nil {
		return nil, errors.Wrap(err, "error getting header data")
	}
	mustUnmarshalJSON(headerData, header)
	return header, nil
}

func GetMerkleBase(db *leveldb.DB, name string) (blob.MerkleBase, error) {
	var base blob.MerkleBase
	baseB, err := db.Get(headerMerkleBasePrefix(name), nil)
	if err != nil {
		return base, errors.Wrap(err, "error getting merkle base")
	}
	if err := base.Decode(bytes.NewReader(baseB)); err != nil {
		panic(err)
	}
	return base, nil
}

func SetHeaderTx(tx *leveldb.Transaction, header *Header, merkleBase blob.MerkleBase) error {
	var buf bytes.Buffer
	if err := merkleBase.Encode(&buf); err != nil {
		return errors.Wrap(err, "error encoding merkle tree")
	}
	exists, err := tx.Has(headerDataPrefix(header.Name), nil)
	if err != nil {
		return errors.Wrap(err, "error checking header existence")
	}
	if err := tx.Put(headerMerkleBasePrefix(header.Name), buf.Bytes(), nil); err != nil {
		return errors.Wrap(err, "error writing merkle tree")
	}
	if err := tx.Put(headerDataPrefix(header.Name), mustMarshalJSON(header), nil); err != nil {
		return errors.Wrap(err, "error writing header tree")
	}
	if !exists {
		if err := IncrementHeaderCount(tx); err != nil {
			return errors.Wrap(err, "error incrementing header count")
		}
	}
	return nil
}

type BlobInfo struct {
	Name         string           `json:"name"`
	PublicKey    *btcec.PublicKey `json:"public_key"`
	ImportHeight int              `json:"import_height"`
	Timestamp    time.Time        `json:"timestamp"`
	MerkleRoot   crypto.Hash      `json:"merkle_root"`
	Signature    crypto.Signature `json:"signature"`
	ReservedRoot crypto.Hash      `json:"reserved_root"`
	ReceivedAt   time.Time        `json:"received_at"`
	Timebank     int              `json:"timebank"`
}

func (b *BlobInfo) MarshalJSON() ([]byte, error) {
	jsonInfo := struct {
		Name         string    `json:"name"`
		PublicKey    string    `json:"public_key"`
		ImportHeight int       `json:"import_height"`
		Timestamp    time.Time `json:"timestamp"`
		MerkleRoot   string    `json:"merkle_root"`
		Signature    string    `json:"signature"`
		ReservedRoot string    `json:"reserved_root"`
		ReceivedAt   time.Time `json:"received_at"`
		Timebank     int       `json:"timebank"`
	}{
		b.Name,
		hex.EncodeToString(b.PublicKey.SerializeCompressed()),
		b.ImportHeight,
		b.Timestamp,
		hex.EncodeToString(b.MerkleRoot[:]),
		hex.EncodeToString(b.Signature[:]),
		hex.EncodeToString(b.ReservedRoot[:]),
		b.ReceivedAt,
		b.Timebank,
	}

	return json.Marshal(jsonInfo)
}

type BlobInfoStream struct {
	db   *leveldb.DB
	iter iterator.Iterator
}

func (bis *BlobInfoStream) Next() (*BlobInfo, error) {
	if !bis.iter.Next() {
		return nil, nil
	}

	header := new(Header)
	mustUnmarshalJSON(bis.iter.Value(), header)
	nameInfo, err := GetNameInfo(bis.db, header.Name)
	if err != nil {
		return nil, errors.Wrap(err, "error getting name info")
	}
	return &BlobInfo{
		Name:         header.Name,
		PublicKey:    nameInfo.PublicKey,
		ImportHeight: nameInfo.ImportHeight,
		Timestamp:    header.Timestamp,
		MerkleRoot:   header.MerkleRoot,
		Signature:    header.Signature,
		ReservedRoot: header.ReservedRoot,
		ReceivedAt:   header.ReceivedAt,
		Timebank:     header.Timebank,
	}, nil
}

func (bis *BlobInfoStream) Close() error {
	bis.iter.Release()
	return bis.iter.Error()
}

func StreamBlobInfo(db *leveldb.DB, start string) (*BlobInfoStream, error) {
	if start == "" {
		return &BlobInfoStream{
			db:   db,
			iter: db.NewIterator(util.BytesPrefix(headerDataPrefix()), nil),
		}, nil
	}

	iterRange := &util.Range{
		Start: headerDataPrefix(start),
		Limit: headerDataPrefix(string([]byte{0xff})),
	}
	last := iterRange.Start[len(iterRange.Start)-1]
	iterRange.Start[len(iterRange.Start)-1] = last + 1
	iter := db.NewIterator(iterRange, nil)
	return &BlobInfoStream{
		db:   db,
		iter: iter,
	}, nil
}

func TruncateHeaderStore(db *leveldb.DB) error {
	err := WithTx(db, func(tx *leveldb.Transaction) error {
		iter := tx.NewIterator(util.BytesPrefix(headersPrefix()), nil)
		for iter.Next() {
			if err := tx.Delete(iter.Key(), nil); err != nil {
				return errors.Wrap(err, "error deleting header store key")
			}
		}
		iter.Release()
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "error truncating header store")
	}
	return nil
}
