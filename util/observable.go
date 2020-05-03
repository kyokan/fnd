package util

import (
	"reflect"
	"sync"
	"sync/atomic"
)

type Unsubscriber func()

type Observable struct {
	listeners map[string]map[int]func(...interface{})
	mtx       sync.RWMutex
	hdl       int
}

func NewObservable() *Observable {
	return &Observable{
		listeners: make(map[string]map[int]func(...interface{})),
	}
}

func (o *Observable) On(evt string, cb interface{}) func() {
	val := reflect.ValueOf(cb)
	if val.Kind() != reflect.Func {
		panic("callback must be of type reflect.Func")
	}

	var closed int32
	wrapped := func(args ...interface{}) {
		if atomic.LoadInt32(&closed) == 1 {
			return
		}
		preparedArgs := make([]reflect.Value, len(args), len(args))
		for i, arg := range args {
			if arg == nil {
				preparedArgs[i] = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())
			} else {
				preparedArgs[i] = reflect.ValueOf(arg)
			}
		}
		val.Call(preparedArgs)
	}

	o.mtx.Lock()
	defer o.mtx.Unlock()
	if o.listeners[evt] == nil {
		o.listeners[evt] = make(map[int]func(...interface{}))
	}
	hdl := o.hdl
	o.hdl++
	o.listeners[evt][hdl] = wrapped
	return func() {
		atomic.StoreInt32(&closed, 1)
		o.off(evt, hdl)
	}
}

func (o *Observable) off(evt string, hdl int) {
	o.mtx.Lock()
	defer o.mtx.Unlock()
	if o.listeners[evt] == nil || o.listeners[evt][hdl] == nil {
		panic("listener not found")
	}
	delete(o.listeners[evt], hdl)
	if len(o.listeners[evt]) == 0 {
		delete(o.listeners, evt)
	}
}

func (o *Observable) Emit(evt string, args ...interface{}) {
	o.mtx.RLock()
	defer o.mtx.RUnlock()
	listeners := o.listeners[evt]
	if len(listeners) == 0 {
		return
	}
	for _, lis := range listeners {
		go lis(args...)
	}
}
