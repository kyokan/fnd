package store

import (
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"time"
)

var (
	lastBanListImportAtKey = []byte("last-ban-list-import-at")
	bansPrefix             = Prefixer("bans")
	banPrefix              = Prefixer(string(bansPrefix("ban")))
)

func GetLastBanListImportAt(db *leveldb.DB) (time.Time, error) {
	res, err := db.Get(lastBanListImportAtKey, nil)
	if errors.Is(err, leveldb.ErrNotFound) {
		return time.Unix(0, 0), nil
	}
	if err != nil {
		return time.Unix(0, 0), errors.Wrap(err, "error getting last ban list import time")
	}
	return mustDecodeTime(res), nil
}

func SetLastBanListImportAt(tx *leveldb.Transaction, t time.Time) error {
	err := tx.Put(lastBanListImportAtKey, encodeTime(t), nil)
	if err != nil {
		return errors.Wrap(err, "error setting last ban list import time")
	}
	return nil
}

func NameIsBanned(db *leveldb.DB, name string) (bool, error) {
	has, err := db.Has(banPrefix(name), nil)
	if err != nil {
		return false, errors.Wrap(err, "error getting name ban state")
	}
	return has, nil
}

func TruncateBannedNames(tx *leveldb.Transaction) error {
	iter := tx.NewIterator(util.BytesPrefix(bansPrefix()), nil)
	for iter.Next() {
		if err := tx.Delete(iter.Key(), nil); err != nil {
			return errors.Wrap(err, "error deleting ban store key")
		}
	}
	iter.Release()
	return nil
}

func BanName(tx *leveldb.Transaction, name string) error {
	err := tx.Put(banPrefix(name), []byte{0x01}, nil)
	if err != nil {
		return errors.Wrap(err, "error inserting banned name")
	}
	return nil
}
