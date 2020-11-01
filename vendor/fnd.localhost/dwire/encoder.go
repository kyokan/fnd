package dwire

import "io"

const (
	DefaultMaxVariableArrayLen = 1024
	DefaultMaxByteFieldLen     = 256 * 1024
)

// Encoder is an interface that allows arbitrary types to be
// encoded. Types implementing the Encoder interface can be
// encoded using EncodeField or EncodeFields.
type Encoder interface {
	Encode(w io.Writer) error
}

// Decoder is an interface that allows arbitrary types to be
// decoded. Types implementing the Decoder interface can be
// decoded using DecodeField or DecodeFields.
type Decoder interface {
	Decode(r io.Reader) error
}

type EncodeDecoder interface {
	Encoder
	Decoder
}

type ConfiguredEncoder struct {
	// MaxVariableArrayLen is the maximum length of a variable-length array dwire
	// will decode before stopping early.
	MaxVariableArrayLen int

	// MaxByteFieldLen is the maximum length of a variable-length byte array field
	// dwire will decode before stopping early.
	MaxByteFieldLen     uint64
}

var defaultEncoder = &ConfiguredEncoder{
	MaxVariableArrayLen: DefaultMaxVariableArrayLen,
	MaxByteFieldLen:     DefaultMaxByteFieldLen,
}
