/*
Package dwire implements the Footnote message encoding scheme as defined in
PIP-1.

Fundamental types:

	- bool: Encoded as 0x00 or 0x01 if the value is true or false,
	  respectively.
	- uint8: Encoded as a single byte in the range 0x00-0xff.
	- uint16: Encoded as two big-endian bytes in the range 0x0000-0xffff.
	- uint32: Encoded as four big-endian bytes in the range
	  0x00000000-0xffffffff
	- uint64: Encoded as eight big-endian bytes in the range
	  0x0000000000000000-0xffffffffffffffff.
	- string: Encoded as a UTF-8 []byte.
	- [N]T: Encoded as the concatenation of the encoding of T.
	- []T: Encoded as a binary.Uvarint length prefix followed by the
	  concatenation of the encoding of T.

Well-known types:

	- time.Time: Encoded as uint32(time.Unix()).

The easiest way to use this library is to call the EncodeField/DecodeField family
of methods. To encode a value into a Writer:

	value1 := "this is my value"
	value2 := 2
	err := dwire.EncodeFields(w, value1, value2)

To decode a value from a reader:

	var value1 string
	var value2 string
	err := dwire.DecodeFields(r, &value1, &value2)

Note that values passed to DecodeField/DecodeFields MUST be pointers.

dwire exposes Encoder and Decoder interfaces, which allow arbitrary types to be
encoded and decoded by EncodeField/DecodeField. For example:

	type Foo struct {
		Value string
	}

	func (f *Foo) Encode(w io.Writer) error {
		return dwire.EncodeFields(w, f.Value)
	}

	func (f *Foo) Decode(r io.Reader) error {
		return dwire.DecodeFields(r, &f.Value)
	}
 */
package dwire
