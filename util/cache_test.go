package util

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestCache_Race(t *testing.T) {
	cache := NewCache()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			for j := 0; j < 10; j++ {
				cache.Set(strconv.Itoa(i), "bonobo", 250)
				cache.Get("bonobo")
				time.Sleep(10 * time.Millisecond)
			}
			wg.Done()
		}(i)
	}

	time.Sleep(1500)
	wg.Wait()
}

func TestCache_GetSetHas_LazyExpiry(t *testing.T) {
	cache := NewCache()
	var reaped atomic.Int32
	cache.ReaperFunc = func(key string, val interface{}) {
		reaped.Store(int32(val.(int)))
	}
	assert.False(t, cache.Has("test"))
	assert.Nil(t, cache.Get("test"))
	cache.Set("test", 123, 10) // expiry is less than reap interval
	assert.True(t, cache.Has("test"))
	assert.EqualValues(t, 123, cache.Get("test"))
	assert.Panics(t, func() {
		cache.Set("test3", nil, 0)
	})

	time.Sleep(50 * time.Millisecond)
	assert.Nil(t, cache.Get("test"))
	assert.False(t, cache.Has("test"))
	assert.EqualValues(t, 123, reaped.Load())
}

func TestCache_GetSetHas_ProactiveExpiry(t *testing.T) {
	cache := NewCache()
	var reaped atomic.Int32
	cache.ReaperFunc = func(key string, val interface{}) {
		reaped.Store(int32(val.(int)))
	}
	cache.ReapInterval = 50 * time.Millisecond
	cache.Set("test", 123, 10)
	time.Sleep(80 * time.Millisecond)
	assert.False(t, cache.Has("test"))
	assert.Nil(t, cache.Get("test"))
	assert.EqualValues(t, 123, reaped.Load())
}

func TestCache_Del(t *testing.T) {
	cache := NewCache()
	cache.Set("test", 123, 0)
	cache.Del("test")
	assert.False(t, cache.Has("test"))
	assert.Nil(t, cache.Get("test"))
}
