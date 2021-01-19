package blob

import (
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

type SectorReader interface {
	ReadSector(sectorID uint8) (Sector, error)
}

type Readable interface {
	io.ReaderAt
	SectorReader
	Name() string
}

type Blob interface {
	io.Closer
	Readable
	Transaction() (Transaction, error)
}

type blobImpl struct {
	f    *os.File
	name string
	mu   sync.Mutex
}

func newFromFile(name string, f *os.File) Blob {
	return &blobImpl{
		name: name,
		f:    f,
	}
}

func (b *blobImpl) Name() string {
	return b.name
}

func (b *blobImpl) ReadSector(id uint8) (Sector, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return ReadSector(b.f, id)
}

func (b *blobImpl) ReadAt(p []byte, off int64) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return ReadBlobAt(b.f, p, off)
}

func (b *blobImpl) Transaction() (Transaction, error) {
	return &txImpl{
		name:      b.name,
		cloner:    b.txCloner,
		committer: b.txCommitter,
		remover:   b.txRemover,
	}, nil
}

func (b *blobImpl) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.f.Close()
}

func (b *blobImpl) txCloner() (*os.File, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	clone, err := ioutil.TempFile("", "blobtx_*")
	if err != nil {
		return nil, errors.Wrap(err, "error cloning blob")
	}
	if _, err := b.f.Seek(0, io.SeekStart); err != nil {
		return nil, errors.Wrap(err, "error cloning blob")
	}
	if _, err := io.Copy(clone, b.f); err != nil {
		return nil, errors.Wrap(err, "error cloning blob")
	}
	return clone, nil
}

func (b *blobImpl) txCommitter(clone *os.File) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, err := clone.Seek(0, io.SeekStart); err != nil {
		return errors.Wrap(err, "error committing blob")
	}
	if _, err := b.f.Seek(0, io.SeekStart); err != nil {
		return errors.Wrap(err, "error committing blob")
	}
	if _, err := io.Copy(b.f, clone); err != nil {
		return errors.Wrap(err, "error committing blob")
	}
	return nil
}

func (b *blobImpl) txRemover() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err := os.Remove(b.f.Name()); err != nil {
		return err
	}
	return nil
}
