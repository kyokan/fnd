package blob

import (
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/pkg/errors"
)

var (
	ErrTransactionClosed  = errors.New("transaction already closed")
	ErrTransactionRemoved = errors.New("transaction removed")
)

type Transaction interface {
	Readable
	io.Seeker
	io.WriterAt
	WriteSector(sector Sector) error
	Truncate() error
	Commit() error
	Rollback() error
	Remove() error
}

type txImpl struct {
	name        string
	f           *os.File
	sectorSize  uint16
	mu          sync.Mutex
	cloner      func() (*os.File, error)
	committer   func(clone *os.File) error
	remover     func() error
	initialized bool
	closed      bool
	removed     bool
}

func (t *txImpl) Name() string {
	return t.name
}

func (t *txImpl) Seek(off int64, whence int) (int64, error) {
	if off%SectorBytes != 0 {
		return 0, errors.New("seek not a multiple of sector len")
	}
	if off < int64(t.sectorSize)*int64(SectorBytes) {
		return 0, errors.New("seek before already written sector")
	}
	switch whence {
	case io.SeekStart:
		if off > Size {
			return 0, errors.New("seek beyond blob bounds")
		}
		t.sectorSize = uint16(off / SectorBytes)
	case io.SeekCurrent:
	case io.SeekEnd:
	default:
		panic("invalid whence")
	}
	return off, nil
}

func (t *txImpl) ReadSector(id uint8) (Sector, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ZeroSector, ErrTransactionClosed
	}
	if t.removed {
		return ZeroSector, ErrTransactionRemoved
	}
	if err := t.lazyInitialize(); err != nil {
		return ZeroSector, errors.Wrap(err, "error initializing transaction")
	}
	return ReadSector(t.f, id)
}

func (t *txImpl) ReadAt(p []byte, off int64) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return 0, ErrTransactionClosed
	}
	if t.removed {
		return 0, ErrTransactionRemoved
	}
	if err := t.lazyInitialize(); err != nil {
		return 0, errors.Wrap(err, "error initializing transaction")
	}
	return ReadBlobAt(t.f, p, off)
}

func (t *txImpl) WriteSector(sector Sector) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ErrTransactionClosed
	}
	if t.removed {
		return ErrTransactionRemoved
	}
	if err := t.lazyInitialize(); err != nil {
		return errors.Wrap(err, "error initializing transaction")
	}
	if err := WriteSector(t.f, t.sectorSize, sector); err != nil {
		return errors.Wrap(err, "error writing sector")
	}
	t.sectorSize++
	return nil
}

func (t *txImpl) WriteAt(p []byte, off int64) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return 0, ErrTransactionClosed
	}
	if t.removed {
		return 0, ErrTransactionRemoved
	}
	if err := t.lazyInitialize(); err != nil {
		return 0, errors.Wrap(err, "error initializing transaction")
	}
	return WriteBlobAt(t.f, p, off)
}

func (t *txImpl) Truncate() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.removed {
		return ErrTransactionRemoved
	}
	if t.initialized {
		if err := os.Remove(t.f.Name()); err != nil {
			return err
		}
	}
	clone, err := ioutil.TempFile("", "blobtx_*")
	if err != nil {
		return errors.Wrap(err, "error creating temporary file")
	}
	if err := clone.Truncate(Size); err != nil {
		return errors.Wrap(err, "error truncating file")
	}
	t.f = clone
	t.initialized = true
	return nil
}

func (t *txImpl) Commit() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ErrTransactionClosed
	}
	// all the below operations need to be atomic, so panic on errors
	// until we implement a more robust recovery method
	if t.removed {
		if err := t.remover(); err != nil {
			panic(err)
		}
	} else if t.initialized {
		if err := t.committer(t.f); err != nil {
			panic(err)
		}
		if err := t.f.Close(); err != nil {
			panic(err)
		}
		if err := os.Remove(t.f.Name()); err != nil {
			panic(err)
		}
	}
	// end atomic section
	t.closed = true
	return nil
}

func (t *txImpl) Rollback() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ErrTransactionClosed
	}
	if !t.initialized || t.removed {
		t.closed = true
		return nil
	}
	if err := t.f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(t.f.Name()); err != nil {
		panic(err)
	}
	t.closed = true
	return nil
}

func (t *txImpl) Remove() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ErrTransactionClosed
	}
	t.removed = true
	if !t.initialized {
		return nil
	}
	if err := t.f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(t.f.Name()); err != nil {
		panic(err)
	}
	return nil
}

func (t *txImpl) lazyInitialize() error {
	if t.initialized {
		return nil
	}
	clone, err := t.cloner()
	if err != nil {
		return errors.Wrap(err, "error initializing transaction")
	}
	t.f = clone
	t.initialized = true
	return nil
}
