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
	Name          string
	EpochHeight   uint16
	SectorSize    uint16
	SectorTipHash crypto.Hash
	Signature     crypto.Signature
	ReservedRoot  crypto.Hash
	EpochStartAt  time.Time
	Banned        bool
	BannedAt      time.Time
}

func (h *Header) MarshalJSON() ([]byte, error) {
	out := &struct {
		Name         string    `json:"name"`
		EpochHeight  uint16    `json:"epoch_height"`
		SectorSize   uint16    `json:"sector_size"`
		SectorTipHash   string    `json:"sector_tip_hash"`
		Signature    string    `json:"signature"`
		ReservedRoot string    `json:"reserved_root"`
		EpochStartAt time.Time `json:"epoch_start_at"`
		Banned       bool      `json:"banned"`
		BannedAt     time.Time `json:"banned_at"`
	}{
		h.Name,
		h.EpochHeight,
		h.SectorSize,
		h.SectorTipHash.String(),
		h.Signature.String(),
		h.ReservedRoot.String(),
		h.EpochStartAt,
		h.Banned,
		h.BannedAt,
	}

	return json.Marshal(out)
}

func (h *Header) UnmarshalJSON(b []byte) error {
	in := &struct {
		Name         string    `json:"name"`
		EpochHeight  uint16    `json:"epoch_height"`
		SectorSize   uint16    `json:"sector_size"`
		SectorTipHash   string    `json:"sector_tip_hash"`
		Signature    string    `json:"signature"`
		ReservedRoot string    `json:"reserved_root"`
		EpochStartAt time.Time `json:"epoch_start_at"`
		Banned       bool      `json:"banned"`
		BannedAt     time.Time `json:"banned_at"`
	}{}
	if err := json.Unmarshal(b, in); err != nil {
		return err
	}
	mrB, err := hex.DecodeString(in.SectorTipHash)
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
	h.EpochHeight = in.EpochHeight
	h.SectorSize = in.SectorSize
	h.SectorTipHash = mr
	h.Signature = sig
	h.ReservedRoot = rr
	h.EpochStartAt = in.EpochStartAt
	h.Banned = in.Banned
	h.BannedAt = in.BannedAt
	return nil
}

var (
	headersPrefix            = Prefixer("headers")
	headerCountKey           = Prefixer(string(headersPrefix("count")))()
	headerSectorHashesPrefix = Prefixer(string(headersPrefix("sector-hashes")))
	headerDataPrefix         = Prefixer(string(headersPrefix("header")))
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

func GetSectorHash(db *leveldb.DB, name string, index uint16) (crypto.Hash, error) {
	hashes, err := GetSectorHashes(db, name)
	if err != nil {
		return crypto.ZeroHash, err
	}
	if int(index) > len(hashes) {
		return crypto.ZeroHash, errors.Wrap(err, "error getting index")
	}
	return hashes[index], nil
}

func GetSectorHashes(db *leveldb.DB, name string) (blob.SectorHashes, error) {
	var base blob.SectorHashes
	baseB, err := db.Get(headerSectorHashesPrefix(name), nil)
	if err != nil {
		return base, errors.Wrap(err, "error getting sector hashes")
	}
	if err := base.Decode(bytes.NewReader(baseB)); err != nil {
		panic(err)
	}
	return base, nil
}

func SetHeaderTx(tx *leveldb.Transaction, header *Header, sectorHashes blob.SectorHashes) error {
	var buf bytes.Buffer
	if err := sectorHashes.Encode(&buf); err != nil {
		return errors.Wrap(err, "error encoding merkle tree")
	}
	exists, err := tx.Has(headerDataPrefix(header.Name), nil)
	if err != nil {
		return errors.Wrap(err, "error checking header existence")
	}
	if err := tx.Put(headerSectorHashesPrefix(header.Name), buf.Bytes(), nil); err != nil {
		return errors.Wrap(err, "error writing sector hashes")
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
	EpochHeight  uint16           `json:"epoch_height"`
	SectorSize   uint16           `json:"sector_size"`
	SectorTipHash   crypto.Hash      `json:"sector_tip_hash"`
	Signature    crypto.Signature `json:"signature"`
	ReservedRoot crypto.Hash      `json:"reserved_root"`
	ReceivedAt   time.Time        `json:"received_at"`
}

func (b *BlobInfo) MarshalJSON() ([]byte, error) {
	jsonInfo := struct {
		Name         string    `json:"name"`
		PublicKey    string    `json:"public_key"`
		ImportHeight int       `json:"import_height"`
		EpochHeight  uint16    `json:"epoch_height"`
		SectorSize   uint16    `json:"sector_size"`
		SectorTipHash   string    `json:"sector_tip_hash"`
		Signature    string    `json:"signature"`
		ReservedRoot string    `json:"reserved_root"`
		ReceivedAt   time.Time `json:"received_at"`
	}{
		b.Name,
		hex.EncodeToString(b.PublicKey.SerializeCompressed()),
		b.ImportHeight,
		b.EpochHeight,
		b.SectorSize,
		hex.EncodeToString(b.SectorTipHash[:]),
		hex.EncodeToString(b.Signature[:]),
		hex.EncodeToString(b.ReservedRoot[:]),
		b.ReceivedAt,
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
		EpochHeight:  header.EpochHeight,
		SectorSize:   header.SectorSize,
		SectorTipHash:   header.SectorTipHash,
		Signature:    header.Signature,
		ReservedRoot: header.ReservedRoot,
		ReceivedAt:   header.EpochStartAt,
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
