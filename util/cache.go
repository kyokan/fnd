package util

import (
	"math"
	"sync"
	"time"
)

const (
	DefaultReapInterval = time.Second
)

type Cache struct {
	ReaperFunc   OnRemove
	ReapInterval time.Duration
	entries      map[string]*entry
	entriesMtx   sync.Mutex
	reapMtx      sync.Mutex
	reapTicker   *time.Ticker
	reapQuitCh   chan struct{}
}

type entry struct {
	val    interface{}
	expiry int64
}

type OnRemove func(key string, val interface{})

var noopOnReap = func(key string, val interface{}) {}

func NewCache() *Cache {
	return &Cache{
		ReaperFunc:   noopOnReap,
		ReapInterval: DefaultReapInterval,
		entries:      make(map[string]*entry),
		reapQuitCh:   make(chan struct{}),
	}
}

func (l *Cache) Get(key string) interface{} {
	l.entriesMtx.Lock()
	defer l.entriesMtx.Unlock()
	entry := l.entries[key]
	if entry == nil {
		return nil
	}
	if time.Now().UnixNano() > entry.expiry {
		delete(l.entries, key)
		l.ReaperFunc(key, entry.val)
		return nil
	}
	return entry.val
}

func (l *Cache) Set(key string, val interface{}, expMS int64) {
	if val == nil {
		panic("cache values cannot be nil")
	}

	var expiry int64
	if expMS == 0 {
		expiry = math.MaxInt16
	} else {
		expiry = time.Now().Add(time.Duration(expMS) * time.Millisecond).UnixNano()
	}

	l.entriesMtx.Lock()
	defer l.entriesMtx.Unlock()
	l.entries[key] = &entry{
		val:    val,
		expiry: expiry,
	}
	l.touchReaper()
}

func (l *Cache) Has(key string) bool {
	return l.Get(key) != nil
}

func (l *Cache) Del(key string) {
	l.entriesMtx.Lock()
	defer l.entriesMtx.Unlock()
	delete(l.entries, key)
}

func (l *Cache) touchReaper() {
	l.reapMtx.Lock()
	defer l.reapMtx.Unlock()
	if l.reapTicker != nil {
		return
	}

	l.reapTicker = time.NewTicker(100 * time.Millisecond)
	go func() {
		for {
			<-l.reapTicker.C
			l.entriesMtx.Lock()
			l.reap()
			if len(l.entries) == 0 {
				l.reapMtx.Lock()
				l.reapTicker = nil
				l.reapMtx.Unlock()
				l.entriesMtx.Unlock()
				return
			}
			l.entriesMtx.Unlock()
		}
	}()
}

func (l *Cache) reap() {
	var toDelete []string
	for k, entry := range l.entries {
		if time.Now().Unix() < entry.expiry {
			continue
		}
		toDelete = append(toDelete, k)
	}

	for _, k := range toDelete {
		l.ReaperFunc(k, l.entries[k].val)
		delete(l.entries, k)
	}
}
