package protocol

import (
	"fnd/blob"
	"fnd/log"
	"fnd/store"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"time"
)

const (
	BanListUpdateInterval = 7 * 24 * time.Hour
)

func IngestBanLists(db *leveldb.DB, bs blob.Store, lists []string) error {
	lgr := log.WithModule("moderation")
	currRev, err := store.GetLastBanListImportAt(db)
	if err != nil {
		return errors.Wrap(err, "failed to fetch latest ban list revision")
	}
	if time.Now().Sub(currRev) < BanListUpdateInterval {
		lgr.Debug("ban list cached")
		return nil
	}

	lgr.Info("refreshing ban lists")
	err = store.WithTx(db, func(tx *leveldb.Transaction) error {
		if err := store.TruncateBannedNames(tx); err != nil {
			return errors.Wrap(err, "error truncating banned names")
		}

		for _, url := range lists {
			lgr.Debug("fetching ban list", "url", url)
			listNames, err := FetchListFile(url)
			if err != nil {
				return errors.Wrap(err, "failed to fetch ban list")
			}

			for _, name := range listNames {
				if err := store.BanName(tx, name); err != nil {
					return errors.Wrap(err, "error banning name")
				}
				exists, err := bs.Exists(name)
				if err != nil {
					return errors.Wrap(err, "error checking blob existence")
				}
				if !exists {
					continue
				}
				lgr.Info("deleting banned name", "name", name)
				bl, err := bs.Open(name)
				if err != nil {
					return errors.Wrap(err, "error opening blob")
				}
				tx, err := bl.Transaction()
				if err != nil {
					return errors.Wrap(err, "error opening transaction")
				}
				if err := tx.Remove(); err != nil {
					return errors.Wrap(err, "error removing banned name")
				}
				if err := tx.Commit(); err != nil {
					return errors.Wrap(err, "error committing transaction")
				}
				if err := bl.Close(); err != nil {
					return errors.Wrap(err, "error closing blob")
				}
			}
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "error ingesting ban lists")
	}
	return nil
}
