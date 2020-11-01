package mstream

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type NopReadWriter struct{}

func (n *NopReadWriter) Read(p []byte) (int, error) {
	return len(p), nil
}

func (n *NopReadWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func BenchmarkUint8Encoding(b *testing.B) {
	var i uint8
	rw := new(NopReadWriter)
	for n := 0; n < b.N; n++ {
		require.NoError(b, EncodeField(rw, i))
	}
}

func BenchmarkUint16Encoding(b *testing.B) {
	var i uint16
	rw := new(NopReadWriter)
	for n := 0; n < b.N; n++ {
		require.NoError(b, EncodeField(rw, i))
	}
}

func BenchmarkUint32Encoding(b *testing.B) {
	var i uint32
	rw := new(NopReadWriter)
	for n := 0; n < b.N; n++ {
		require.NoError(b, EncodeField(rw, i))
	}
}

func BenchmarkUint64Encoding(b *testing.B) {
	var i uint64
	rw := new(NopReadWriter)
	for n := 0; n < b.N; n++ {
		require.NoError(b, EncodeField(rw, i))
	}
}

func BenchmarkByteSliceEncoding1024(b *testing.B) {
	bytes := make([]byte, 1024, 1024)
	rw := new(NopReadWriter)
	for n := 0; n < b.N; n++ {
		require.NoError(b, EncodeField(rw, bytes))
	}
}

func BenchmarkStringEncoding1024(b *testing.B) {
	bytes := make([]byte, 1024, 1024)
	str := string(bytes)
	rw := new(NopReadWriter)
	for n := 0; n < b.N; n++ {
		require.NoError(b, EncodeField(rw, str))
	}
}

func BenchmarkByteArrayEncoding32(b *testing.B) {
	bytes := make([]byte, 32, 32)
	rw := new(NopReadWriter)
	for n := 0; n < b.N; n++ {
		require.NoError(b, EncodeField(rw, bytes))
	}
}

func BenchmarkByteArrayEncodingReflect(b *testing.B) {
	var bytes [1024]byte
	rw := new(NopReadWriter)
	for n := 0; n < b.N; n++ {
		require.NoError(b, EncodeField(rw, bytes))
	}
}

func BenchmarkUint16ArrayEncodingReflect(b *testing.B) {
	var uints [1024]uint16
	rw := new(NopReadWriter)
	for n := 0; n < b.N; n++ {
		require.NoError(b, EncodeField(rw, uints))
	}
}

func BenchmarkStringArrayEncodingReflect(b *testing.B) {
	var strings [1024]string
	rw := new(NopReadWriter)
	for n := 0; n < b.N; n++ {
		require.NoError(b, EncodeField(rw, strings))
	}
}

func BenchmarkWellKnownTimeEncoding(b *testing.B) {
	ts := time.Now()
	rw := new(NopReadWriter)
	for n := 0; n < b.N; n++ {
		require.NoError(b, EncodeField(rw, ts))
	}
}
