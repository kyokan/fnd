package util

import (
	"sync"
)

type refCounter struct {
	readers int64
	writers int64
}

type MultiLocker interface {
	TryLock(interface{}) bool
	TryRLock(interface{}) bool
	Unlock(interface{})
	RUnlock(interface{})
}

type lock struct {
	inUse map[interface{}]*refCounter
	mtx   sync.Mutex
}

func NewMultiLocker() MultiLocker {
	return &lock{
		inUse: make(map[interface{}]*refCounter),
	}
}

func (l *lock) TryLock(key interface{}) bool {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	m := l.getLocker(key)
	if m.readers > 0 || m.writers > 0 {
		return false
	}
	m.writers++
	return true
}

func (l *lock) TryRLock(key interface{}) bool {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	m := l.getLocker(key)
	if m.writers > 0 {
		return false
	}
	m.readers++
	return true
}

func (l *lock) Unlock(key interface{}) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	m := l.getLocker(key)
	if m.writers != 1 {
		panic("unlocking unlocked multi locker")
	}
	m.writers--
	if m.writers == 0 {
		delete(l.inUse, key)
	}
}

func (l *lock) RUnlock(key interface{}) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	m := l.getLocker(key)
	if m.readers < 1 {
		panic("unlocking unlocked multi locker")
	}
	m.readers--
	if m.readers == 0 {
		delete(l.inUse, key)
	}
}

func (l *lock) getLocker(key interface{}) *refCounter {
	res, ok := l.inUse[key]
	if !ok {
		res = &refCounter{}
		l.inUse[key] = res
	}

	return res
}
