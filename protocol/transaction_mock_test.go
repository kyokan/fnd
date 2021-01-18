package protocol

import (
	"fnd/blob"
	"github.com/stretchr/testify/mock"
)

type TransactionMock struct {
	mock.Mock
}

func (t *TransactionMock) ReadAt(p []byte, off int64) (n int, err error) {
	args := t.Called(p, off)
	return args.Int(0), args.Error(1)
}

func (t *TransactionMock) ReadSector(sectorID uint8) (blob.Sector, error) {
	args := t.Called(sectorID)
	return args.Get(0).(blob.Sector), args.Error(1)
}

func (t *TransactionMock) Name() string {
	args := t.Called()
	return args.String(0)
}

func (t *TransactionMock) WriteAt(p []byte, off int64) (n int, err error) {
	args := t.Called(p, off)
	return args.Int(0), args.Error(1)
}

func (t *TransactionMock) WriteSector(id uint8, sector blob.Sector) error {
	args := t.Called(id, sector)
	return args.Error(0)
}

func (t *TransactionMock) Truncate() error {
	args := t.Called()
	return args.Error(0)
}

func (t *TransactionMock) Commit() error {
	args := t.Called()
	return args.Error(0)
}

func (t *TransactionMock) Rollback() error {
	args := t.Called()
	return args.Error(0)
}

func (t *TransactionMock) Remove() error {
	args := t.Called()
	return args.Error(0)
}
