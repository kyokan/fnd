package util

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestMultiLocker(t *testing.T) {
	locker := NewMultiLocker()
	n1 := "name1"
	n2 := "name2"

	assert.True(t, locker.TryLock(n1))
	assert.False(t, locker.TryLock(n1))
	assert.False(t, locker.TryRLock(n1))
	assert.True(t, locker.TryLock(n2))
	locker.Unlock(n1)
	locker.Unlock(n2)

	assert.Panics(t, func() {
		locker.Unlock(n1)
	})
	assert.Panics(t, func() {
		locker.RUnlock(n1)
	})

	var wg sync.WaitGroup
	for i := 0; i < 7; i++ {
		wg.Add(1)
		go func() {
			assert.True(t, locker.TryRLock(n1))
			wg.Done()
		}()
	}
	wg.Wait()

	assert.Panics(t, func() {
		locker.Unlock(n1)
	})

	assert.False(t, locker.TryLock(n1))
	assert.True(t, locker.TryRLock(n1))

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			locker.RUnlock(n1)
			wg.Done()
		}()
	}
	wg.Wait()

	assert.Panics(t, func() {
		locker.RUnlock(n1)
	})

	assert.True(t, locker.TryLock(n1))
}
