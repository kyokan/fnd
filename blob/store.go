package blob

import (
	"errors"
	"os"
	"path"
	"path/filepath"
)

type Store interface {
	Open(name string) (Blob, error)
	Exists(name string) (bool, error)
	Reset(name string) error
}

type storeImpl struct {
	blobsPath string
	pool      *Pool
}

func NewStore(blobsPath string) *storeImpl {
	s := &storeImpl{
		blobsPath: blobsPath,
		pool: NewPool(func(name string) (Blob, error) {
			return NewInStorePath(blobsPath, name)
		}),
	}
	return s
}

func (s *storeImpl) Open(name string) (Blob, error) {
	blob, err := s.pool.Get(name)
	if err != nil {
		return nil, err
	}
	return &wrappedBlob{
		pool: s.pool,
		blob: blob,
	}, nil
}

func (s *storeImpl) Exists(name string) (bool, error) {
	return fileExists(path.Join(s.blobsPath, PathifyName(name)))
}

func (s *storeImpl) Reset(name string) error {
	blobSubpath := PathifyName(name)
	blobFile := path.Join(s.blobsPath, blobSubpath)
	exists, err := fileExists(blobFile)
	var f *os.File
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	err = os.Remove(path.Join(s.blobsPath, PathifyName(name)))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(blobFile), 0700); err != nil {
		return err
	}
	f, err = os.OpenFile(blobFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	if err := f.Truncate(Size); err != nil {
		return err
	}
	return nil
}

func NewInStorePath(blobsPath string, name string) (Blob, error) {
	blobSubpath := PathifyName(name)
	blobFile := path.Join(blobsPath, blobSubpath)
	exists, err := fileExists(blobFile)
	var f *os.File
	if err != nil {
		return nil, err
	}
	// TODO: Open in append-only mode (O_APPEND)
	if exists {
		f, err = os.OpenFile(blobFile, os.O_RDWR, 0666)
	} else {
		if err := os.MkdirAll(filepath.Dir(blobFile), 0700); err != nil {
			return nil, err
		}
		f, err = os.OpenFile(blobFile, os.O_RDWR|os.O_CREATE, 0666)
	}
	if err != nil {
		return nil, err
	}
	if err := f.Truncate(Size); err != nil {
		return nil, err
	}

	return newFromFile(name, f), nil
}

func fileExists(f string) (bool, error) {
	info, err := os.Stat(f)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return false, errors.New("is a directory")
	}
	return true, nil
}

type wrappedBlob struct {
	pool *Pool
	blob Blob
}

func (w *wrappedBlob) Name() string {
	return w.blob.Name()
}

func (w *wrappedBlob) ReadAt(p []byte, off int64) (n int, err error) {
	return w.blob.ReadAt(p, off)
}

func (w *wrappedBlob) ReadSector(id uint8) (Sector, error) {
	return w.blob.ReadSector(id)
}

func (w *wrappedBlob) Transaction() (Transaction, error) {
	return w.blob.Transaction()
}

func (w *wrappedBlob) Close() error {
	return w.pool.Put(w.blob)
}
