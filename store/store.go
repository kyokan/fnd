package store

import (
	"github.com/ddrp-org/ddrp/log"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

type TxCb func(tx *leveldb.Transaction) error

var logger = log.WithModule("store")

func Open(path string) (*leveldb.DB, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error opening database")
	}
	return db, nil
}

func WithTx(db *leveldb.DB, cb TxCb) error {
	tx, err := db.OpenTransaction()
	if err != nil {
		return errors.Wrap(err, "error opening transaction")
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Discard()
			panic(p)
		} else if err != nil {
			tx.Discard()
		} else {
			err = tx.Commit()
		}
	}()

	return cb(tx)
}
