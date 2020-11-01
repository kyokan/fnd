package mstream

import "io"

const (
	DefaultMaxVariableArrayLen = 1024
	DefaultMaxByteFieldLen     = 256 * 1024
)

type Encoder interface {
	Encode(w io.Writer) error
}

type Decoder interface {
	Decode(r io.Reader) error
}

type EncodeDecoder interface {
	Encoder
	Decoder
}

type ConfiguredEncoder struct {
	MaxVariableArrayLen int
	MaxByteFieldLen     uint64
}

var defaultEncoder = &ConfiguredEncoder{
	MaxVariableArrayLen: DefaultMaxVariableArrayLen,
	MaxByteFieldLen:     DefaultMaxByteFieldLen,
}
