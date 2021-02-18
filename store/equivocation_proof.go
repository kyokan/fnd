package store

import (
	"bufio"
	"bytes"
	"fnd/wire"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	equivocationProofsPrefix = Prefixer("equivocationproofs")
)

func GeEquivocationProof(db *leveldb.DB, name string) (*wire.EquivocationProof, error) {
	proof := new(wire.EquivocationProof)
	equivocationProofData, err := db.Get(equivocationProofsPrefix(name), nil)
	if err != nil {
		return nil, errors.Wrap(err, "error getting equivocation proof data")
	}
	mustUnmarshalJSON(equivocationProofData, proof)
	return proof, nil
}

func SetEquivocationProofTx(tx *leveldb.Transaction, name string, proof *wire.EquivocationProof) error {
	var buf bytes.Buffer
	wr := bufio.NewWriter(&buf)
	proof.Encode(wr)
	if err := tx.Put(equivocationProofsPrefix(proof.Name), buf.Bytes(), nil); err != nil {
		return errors.Wrap(err, "error writing header tree")
	}
	return nil
}
