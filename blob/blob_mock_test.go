package blob

import "github.com/stretchr/testify/mock"

type BlobMock struct {
	mock.Mock
}

func (b BlobMock) Close() error {
	args := b.Called()
	return args.Error(0)
}

func (b BlobMock) ReadAt(p []byte, off int64) (n int, err error) {
	args := b.Called(p, off)
	return args.Int(0), args.Error(1)
}

func (b BlobMock) ReadSector(sectorID uint8) (Sector, error) {
	args := b.Called(sectorID)
	return args.Get(0).(Sector), args.Error(1)
}

func (b BlobMock) Name() string {
	args := b.Called()
	return args.String(0)
}

func (b BlobMock) Seek(sectorSize uint16) {
}

func (b BlobMock) At() uint16 {
	return 0
}

func (b BlobMock) Transaction() (Transaction, error) {
	args := b.Called()
	return args.Get(0).(Transaction), args.Error(1)
}
