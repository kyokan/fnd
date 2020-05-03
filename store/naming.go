package store

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcd/btcec"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math"
)

var (
	lastNameImportHeightKey  = []byte("last-name-import-height")
	initialImportCompleteKey = []byte("initial-import-complete")
	namesPrefix              = Prefixer("names")
	nameDataPrefix           = Prefixer(string(namesPrefix("name")))
)

func GetLastNameImportHeight(db *leveldb.DB) (int, error) {
	res, err := db.Get(lastNameImportHeightKey, nil)
	if errors.Is(err, leveldb.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrap(err, "error getting last name import height")
	}
	return mustDecodeInt(res), nil
}

func SetLastNameImportHeightTx(tx *leveldb.Transaction, height int) error {
	err := tx.Put(lastNameImportHeightKey, mustEncodeInt(height), nil)
	if err != nil {
		return errors.Wrap(err, "error setting last name import height")
	}
	return nil
}

func GetInitialImportComplete(db *leveldb.DB) (bool, error) {
	res, err := db.Has(initialImportCompleteKey, nil)
	if err != nil {
		return false, errors.Wrap(err, "error getting initial import complete")
	}
	return res, nil
}

func SetInitialImportCompleteTx(tx *leveldb.Transaction) error {
	err := tx.Put(initialImportCompleteKey, []byte{0x01}, nil)
	if err != nil {
		return errors.Wrap(err, "error setting initial import complete")
	}
	return nil
}

type NameInfo struct {
	Name         string
	PublicKey    *btcec.PublicKey
	ImportHeight int
}

func (n *NameInfo) MarshalJSON() ([]byte, error) {
	out := struct {
		Name         string `json:"name"`
		PublicKey    string `json:"public_key"`
		ImportHeight int    `json:"import_height"`
	}{
		n.Name,
		hex.EncodeToString(n.PublicKey.SerializeCompressed()),
		n.ImportHeight,
	}
	return json.Marshal(out)
}

func (n *NameInfo) UnmarshalJSON(data []byte) error {
	out := &struct {
		Name         string `json:"name"`
		PublicKey    string `json:"public_key"`
		ImportHeight int    `json:"import_height"`
	}{}
	if err := json.Unmarshal(data, out); err != nil {
		return err
	}
	n.Name = out.Name
	n.PublicKey = mustDecodePublicKey(out.PublicKey)
	n.ImportHeight = out.ImportHeight
	return nil
}

func GetNameInfo(db *leveldb.DB, name string) (*NameInfo, error) {
	res, err := db.Get(nameDataPrefix(name), nil)
	if err != nil {
		return nil, errors.Wrap(err, "error getting name info")
	}
	info := new(NameInfo)
	mustUnmarshalJSON(res, info)
	return info, nil
}

type NameInfoStream struct {
	iter iterator.Iterator
}

func (nis *NameInfoStream) Next() (*NameInfo, error) {
	if !nis.iter.Next() {
		return nil, nil
	}

	info := new(NameInfo)
	mustUnmarshalJSON(nis.iter.Value(), info)
	return info, nil
}

func (nis *NameInfoStream) Close() error {
	nis.iter.Release()
	return nis.iter.Error()
}

func StreamNameInfo(db *leveldb.DB, start string) (*NameInfoStream, error) {
	if start == "" {
		return &NameInfoStream{
			iter: db.NewIterator(util.BytesPrefix(nameDataPrefix()), nil),
		}, nil
	}

	iterRange := &util.Range{
		Start: nameDataPrefix(start),
		Limit: nameDataPrefix(string([]byte{0xff})),
	}
	last := iterRange.Start[len(iterRange.Start)-1]
	iterRange.Start[len(iterRange.Start)-1] = last + 1
	iter := db.NewIterator(iterRange, nil)
	return &NameInfoStream{
		iter: iter,
	}, nil
}

func SetNameInfoTx(tx *leveldb.Transaction, name string, key *btcec.PublicKey, height int) error {
	err := tx.Put(nameDataPrefix(name), mustMarshalJSON(&NameInfo{
		Name:         name,
		PublicKey:    key,
		ImportHeight: height,
	}), nil)
	if err != nil {
		return errors.Wrap(err, "error inserting name info")
	}
	return nil
}

func TruncateNameStore(db *leveldb.DB) error {
	err := WithTx(db, func(tx *leveldb.Transaction) error {
		iter := tx.NewIterator(util.BytesPrefix(namesPrefix()), nil)
		for iter.Next() {
			if err := tx.Delete(iter.Key(), nil); err != nil {
				return errors.Wrap(err, "error deleting name store key")
			}
		}
		iter.Release()
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "error truncating name store")
	}
	return nil
}

func mustEncodeInt(in int) []byte {
	buf := make([]byte, 8, 8)
	binary.BigEndian.PutUint64(buf, uint64(in))
	return buf
}

func mustDecodeInt(in []byte) int {
	if len(in) == 0 {
		return 0
	}
	out := binary.BigEndian.Uint64(in)
	if out > math.MaxInt32 {
		panic("overflow")
	}
	return int(out)
}

func mustDecodePublicKey(in string) *btcec.PublicKey {
	data, err := hex.DecodeString(in)
	if err != nil {
		panic(err)
	}
	pub, err := btcec.ParsePubKey(data, btcec.S256())
	if err != nil {
		panic(err)
	}
	return pub
}
