package store

import (
	"bytes"
	"fnd/wire"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	equivocationProofsPrefix = Prefixer("equivocationproofs")
)

func GetEquivocationProof(db *leveldb.DB, name string) ([]byte, error) {
	bytes, err := db.Get(equivocationProofsPrefix(name), nil)
	if err != nil {
		return nil, errors.Wrap(err, "error getting equivocation proof")
	}
	return bytes, nil
}

func SetEquivocationProofTx(tx *leveldb.Transaction, name string, proof *wire.EquivocationProof) error {
	var buf bytes.Buffer
	proof.Encode(&buf)
	err := tx.Put(equivocationProofsPrefix(name), buf.Bytes(), nil)
	if err != nil {
		return errors.Wrap(err, "error setting equivocation proof")
	}
	return nil
}
