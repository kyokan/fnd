package mockapp

import (
	"io"
	"io/ioutil"
)

type reader struct{}

func (reader) Read(b []byte) (int, error) {
	for i := range b {
		b[i] = 0
	}
	return len(b), nil
}

var NullReader io.Reader

func Read(b []byte) (n int, err error) {
	return NullReader.Read(b)
}

var NullWriter io.Writer

func Write(p []byte) (n int, err error) {
	return NullWriter.Write(p)
}

func init() {
	NullReader = new(reader)
	NullWriter = ioutil.Discard
}
