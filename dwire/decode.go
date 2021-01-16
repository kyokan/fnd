package dwire

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"

	"github.com/pkg/errors"
)

type byteReader struct {
	r   io.Reader
	buf []byte
}

func newByteReader(r io.Reader) *byteReader {
	return &byteReader{
		r:   r,
		buf: make([]byte, 1, 1),
	}
}

func (r *byteReader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func (r *byteReader) ReadByte() (byte, error) {
	_, err := io.ReadFull(r.r, r.buf)
	if err != nil {
		return 0, err
	}
	return r.buf[0], nil
}

// DecodeFields decodes each field in the variadic items argument from the
// Reader using the default Encoder. Items provided to DecodeFields
// must be pointer types.
func DecodeFields(r io.Reader, items ...interface{}) error {
	return defaultEncoder.DecodeFields(r, items...)
}

// DecodeFields decodes the field in the item argument from the Reader using
// the default Encoder. The item provided to DecodeField must be a pointer type.
func DecodeField(r io.Reader, item interface{}) error {
	return defaultEncoder.DecodeField(r, item)
}

// DecodeFields decodes each field in the variadic items argument
// from the Reader. Items provided to DecodeFields must be pointer types.
func (c *ConfiguredEncoder) DecodeFields(r io.Reader, items ...interface{}) error {
	for _, item := range items {
		if err := c.DecodeField(r, item); err != nil {
			return err
		}
	}

	return nil
}

// DecodeField decodes the field in the item argument from the Reader. The item
// provided to DecodeField must be a pointer type.
func (c *ConfiguredEncoder) DecodeField(r io.Reader, item interface{}) error {
	var err error
	switch it := item.(type) {
	case Decoder:
		err = it.Decode(r)
	case *bool:
		b := make([]byte, 1, 1)
		if _, err := io.ReadFull(r, b); err != nil {
			return err
		}
		if b[0] == 0x00 {
			*it = false
		} else if b[0] == 0x01 {
			*it = true
		} else {
			return errors.Errorf("invalid boolean value: %x", b[0])
		}
	case *uint8:
		b := make([]byte, 1, 1)
		if _, err := io.ReadFull(r, b); err != nil {
			return err
		}
		*it = b[0]
	case *uint16:
		b := make([]byte, 2, 2)
		if _, err := io.ReadFull(r, b); err != nil {
			return err
		}
		*it = binary.BigEndian.Uint16(b)
	case *uint32:
		b := make([]byte, 4, 4)
		if _, err := io.ReadFull(r, b); err != nil {
			return err
		}
		*it = binary.BigEndian.Uint32(b)
	case *uint64:
		b := make([]byte, 8, 8)
		if _, err := io.ReadFull(r, b); err != nil {
			return err
		}
		*it = binary.BigEndian.Uint64(b)
	case *[]byte:
		br := newByteReader(r)
		l, err := binary.ReadUvarint(br)
		if err != nil {
			return err
		}
		if l > c.MaxByteFieldLen {
			return errors.New("byte-assignable field length too large to decode")
		}
		buf := make([]byte, l, l)
		if _, err := io.ReadFull(r, buf); err != nil {
			return err
		}
		*it = buf
	case *string:
		var buf []byte
		if err := c.DecodeField(r, &buf); err != nil {
			return err
		}
		*it = string(buf)
	case *[32]byte:
		var buf [32]byte
		if _, err := io.ReadFull(r, buf[:]); err != nil {
			return err
		}
		*it = buf
	default:
		err = c.decodeReflect(r, item)
	}

	return err
}

func (c *ConfiguredEncoder) decodeReflect(r io.Reader, item interface{}) error {
	itemT := reflect.TypeOf(item)
	if itemT.Kind() != reflect.Ptr {
		return errors.New("can only decode into pointer types")
	}

	canonicalized := canonicalizeWellKnown(itemT.Elem())
	if wellKnownDecoders[canonicalized] != nil {
		return wellKnownDecoders[canonicalized](r, item)
	}

	elemKind := itemT.Elem().Kind()
	if itemT.Elem().Kind() == reflect.Array {
		return c.decodeArray(r, item)
	}

	if elemKind == reflect.Slice {
		return c.decodeSlice(r, item)
	}

	return errors.New(fmt.Sprintf("type %s cannot be decoded", itemT.String()))
}

func (c *ConfiguredEncoder) decodeArray(r io.Reader, item interface{}) error {
	itemVal := reflect.ValueOf(item)
	indirectVal := reflect.Indirect(itemVal)
	indirectT := indirectVal.Type()

	l := indirectT.Len()
	tmp := reflect.Zero(reflect.ArrayOf(l, indirectT.Elem()))
	tmpPtr := reflect.New(indirectT)
	tmpPtr.Elem().Set(tmp)
	if indirectT.Elem().Kind() == reflect.Uint8 {
		buf := make([]byte, l, l)
		if _, err := io.ReadFull(r, buf); err != nil {
			return err
		}
		reflect.Copy(tmpPtr.Elem().Slice(0, l), reflect.ValueOf(buf))
	} else {
		for i := 0; i < indirectVal.Len(); i++ {
			if err := c.DecodeField(r, tmpPtr.Elem().Index(i).Addr().Interface()); err != nil {
				return err
			}
		}
	}
	itemVal.Elem().Set(tmpPtr.Elem())
	return nil
}

func (c *ConfiguredEncoder) decodeSlice(r io.Reader, item interface{}) error {
	itemVal := reflect.ValueOf(item)
	indirectVal := reflect.Indirect(itemVal)
	indirectT := indirectVal.Type()

	tmp := reflect.Zero(reflect.SliceOf(indirectT.Elem()))
	tmpPtr := reflect.New(indirectT)
	tmpPtr.Elem().Set(tmp)

	br := newByteReader(r)
	l, err := binary.ReadUvarint(br)
	if err != nil {
		return err
	}
	if l > uint64(c.MaxVariableArrayLen) {
		return errors.New("variable array field length too large to decode")
	}

	if indirectT.Elem().Kind() == reflect.Ptr {
		for i := 0; i < int(l); i++ {
			sliceItem := reflect.Zero(indirectT.Elem().Elem())
			sliceItemPtr := reflect.New(sliceItem.Type())
			sliceItemPtr.Elem().Set(sliceItem)
			if err := c.DecodeField(r, sliceItemPtr.Interface()); err != nil {
				return err
			}
			tmpPtr.Elem().Set(reflect.Append(tmpPtr.Elem(), sliceItemPtr))
		}
	} else {
		for i := 0; i < int(l); i++ {
			sliceItem := reflect.Zero(indirectT.Elem())
			sliceItemPtr := reflect.New(indirectT.Elem())
			sliceItemPtr.Elem().Set(sliceItem)
			if err := c.DecodeField(r, sliceItemPtr.Interface()); err != nil {
				return err
			}
			tmpPtr.Elem().Set(reflect.Append(tmpPtr.Elem(), sliceItemPtr.Elem()))
		}
	}

	itemVal.Elem().Set(tmpPtr.Elem())
	return nil
}
