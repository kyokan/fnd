package blob

import (
	"crypto/rand"
	"fmt"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"
	"io"
	"os"
	"sync"
	"testing"
)

func TestBlob_Transaction_IO(t *testing.T) {
	f, done := newTempBlobFile(t)
	defer done()

	blob := newFromFile("whatever", f)
	h, _ := blake2b.New256(nil)
	tx, err := blob.Transaction()
	require.NoError(t, err)
	_, err = io.CopyN(NewWriter(tx), io.TeeReader(rand.Reader, h), Size)
	require.NoError(t, err)
	expHash := h.Sum(nil)

	h.Reset()
	require.NoError(t, tx.Commit())
	_, err = io.Copy(h, NewReader(blob))
	require.NoError(t, err)
	actHash := h.Sum(nil)

	require.Equal(t, expHash, actHash)
}

func TestBlob_Transaction_Truncation_Initialized(t *testing.T) {
	f, done := newTempBlobFile(t)
	defer done()

	blob := newFromFile("whatever", f)
	tx, err := blob.Transaction()
	require.NoError(t, err)
	_, err = io.CopyN(NewWriter(tx), rand.Reader, Size)
	require.NoError(t, err)
	sector, err := tx.ReadSector(0)
	require.NoError(t, err)
	require.NotEqual(t, ZeroSector, sector)
	require.NoError(t, tx.Truncate())
	require.NoError(t, err)
	sector, err = tx.ReadSector(0)
	require.NoError(t, err)
	require.Equal(t, ZeroSector, sector)

	require.NoError(t, tx.Commit())
	sector, err = blob.ReadSector(0)
	require.NoError(t, err)
	require.Equal(t, ZeroSector, sector)
}

func TestBlob_Transaction_Truncation_Uninitialized(t *testing.T) {
	f, done := newTempBlobFile(t)
	defer done()

	_, err := io.CopyN(f, rand.Reader, Size)
	require.NoError(t, err)

	blob := newFromFile("whatever", f)
	tx, err := blob.Transaction()
	require.NoError(t, err)
	require.NoError(t, tx.Truncate())
	sector, err := tx.ReadSector(0)
	require.NoError(t, err)
	require.Equal(t, ZeroSector, sector)

	require.NoError(t, tx.Commit())
	sector, err = blob.ReadSector(0)
	require.NoError(t, err)
	require.Equal(t, ZeroSector, sector)
}

func TestBlob_Transaction_Remove(t *testing.T) {
	// source file gets removed by tx.Commit(), so no need
	// to defer done()
	f, _ := newTempBlobFile(t)
	_, err := io.CopyN(f, rand.Reader, Size)
	require.NoError(t, err)

	blob := newFromFile("whatever", f)
	tx, err := blob.Transaction()
	require.NoError(t, err)
	require.NoError(t, tx.Remove())

	notAllowed := []func() error{
		func() error {
			_, err := tx.ReadSector(0)
			return err
		},
		func() error {
			_, err := tx.ReadAt(make([]byte, 8, 8), 0)
			return err
		},
		func() error {
			return tx.WriteSector(0, ZeroSector)
		},
		func() error {
			_, err := tx.WriteAt(make([]byte, 8, 8), 0)
			return err
		},
	}
	for _, tt := range notAllowed {
		require.Equal(t, ErrTransactionRemoved, tt())
	}

	require.NoError(t, tx.Commit())
	_, err = os.Stat(f.Name())
	require.True(t, os.IsNotExist(err))
	requireTxMethodsClosed(t, tx)
}

func TestBlob_Rollback_Initialized(t *testing.T) {
	f, done := newTempBlobFile(t)
	defer done()

	blob := newFromFile("whatever", f)
	tx, err := blob.Transaction()
	require.NoError(t, err)
	_, err = io.CopyN(NewWriter(tx), rand.Reader, Size)
	require.NoError(t, err)
	sector, err := tx.ReadSector(0)
	require.NoError(t, err)
	require.NotEqual(t, ZeroSector, sector)

	require.NoError(t, tx.Rollback())
	sector, err = blob.ReadSector(0)
	require.NoError(t, err)
	require.Equal(t, ZeroSector, sector)
	requireTxMethodsClosed(t, tx)
}

func TestBlob_Rollback_Uninitialized(t *testing.T) {
	f, done := newTempBlobFile(t)
	defer done()

	blob := newFromFile("whatever", f)
	tx, err := blob.Transaction()
	require.NoError(t, err)
	sector, err := tx.ReadSector(0)
	require.NoError(t, err)
	require.Equal(t, ZeroSector, sector)

	require.NoError(t, tx.Rollback())
	sector, err = blob.ReadSector(0)
	require.NoError(t, err)
	require.Equal(t, ZeroSector, sector)
	requireTxMethodsClosed(t, tx)
}

func TestBlob_Transaction_Race(t *testing.T) {
	f, done := newTempBlobFile(t)
	defer done()

	_, err := io.CopyN(f, rand.Reader, Size)
	require.NoError(t, err)

	blob := newFromFile("whatever", f)
	tx, err := blob.Transaction()
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 255; i++ {
		wg.Add(1)
		go func(id uint8) {
			_, _ = tx.ReadSector(id)
			wg.Done()
		}(uint8(i))
	}
	for i := 0; i < 255; i++ {
		wg.Add(1)
		go func(id uint8) {
			_ = tx.WriteSector(id, ZeroSector)
			wg.Done()
		}(uint8(i))
	}
	wg.Add(1)
	go func() {
		_ = tx.Commit()
		wg.Done()
	}()
	wg.Wait()
}

func requireTxMethodsClosed(t *testing.T, tx Transaction) {
	closedFuncs := []struct {
		name string
		fn   func() error
	}{
		{
			"ReadSector",
			func() error {
				_, err := tx.ReadSector(0)
				return err
			},
		},
		{
			"ReadAt",
			func() error {
				_, err := tx.ReadAt(make([]byte, 8, 8), 0)
				return err
			},
		},
		{
			"WriteSector",
			func() error {
				return tx.WriteSector(0, ZeroSector)
			},
		},
		{
			"WriteAt",
			func() error {
				_, err := tx.WriteAt(make([]byte, 8, 8), 0)
				return err
			},
		},
		{
			"Commit",
			func() error {
				return tx.Commit()
			},
		},
		{
			"Rollback",
			func() error {
				return tx.Rollback()
			},
		},
	}
	for _, tt := range closedFuncs {
		t.Run(fmt.Sprintf("method %s should return ErrTransactionClosed", tt.name), func(t *testing.T) {
			require.Equal(t, ErrTransactionClosed, tt.fn())
		})
	}
}
