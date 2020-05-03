package util

import (
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestObservable(t *testing.T) {
	obs := NewObservable()

	var m sync.Map
	ch := make(chan struct{})
	offStr := obs.On("strEvent", func(arg string) {
		m.Store("strEvent", arg)
		ch <- struct{}{}
	})
	obs.On("intEvent", func(arg int) {
		m.Store("intEvent", arg)
		ch <- struct{}{}
	})

	obs.Emit("strEvent", "hello")
	obs.Emit("intEvent", 1)
	<-ch
	<-ch

	strVal, _ := m.Load("strEvent")
	require.Equal(t, "hello", strVal.(string))
	intVal, _ := m.Load("intEvent")
	require.Equal(t, 1, intVal.(int))

	offStr()
	obs.Emit("strEvent", "sup")
	strVal, _ = m.Load("strEvent")
	require.Equal(t, "hello", strVal.(string))
	require.Panics(t, func() {
		offStr()
	})
}
