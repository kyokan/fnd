package mstream

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

var (
	trueWire  = []byte{0x01}
	falseWire = []byte{0x00}
)

func EncodeFields(w io.Writer, items ...interface{}) error {
	return defaultEncoder.EncodeFields(w, items...)
}

func EncodeField(w io.Writer, item interface{}) error {
	return defaultEncoder.EncodeField(w, item)
}

func (c *ConfiguredEncoder) EncodeFields(w io.Writer, items ...interface{}) error {
	for _, item := range items {
		if err := c.EncodeField(w, item); err != nil {
			return err
		}
	}

	return nil
}

func (c *ConfiguredEncoder) EncodeField(w io.Writer, item interface{}) error {
	var err error
	switch it := item.(type) {
	case Encoder:
		err = it.Encode(w)
	case bool:
		val := falseWire
		if it {
			val = trueWire
		}
		_, err = w.Write(val)
	case uint8:
		_, err = w.Write([]byte{it})
	case uint16:
		b := make([]byte, 2, 2)
		binary.BigEndian.PutUint16(b, it)
		_, err = w.Write(b)
	case uint32:
		b := make([]byte, 4, 4)
		binary.BigEndian.PutUint32(b, it)
		_, err = w.Write(b)
	case uint64:
		b := make([]byte, 8, 8)
		binary.BigEndian.PutUint64(b, it)
		_, err = w.Write(b)
	case []byte:
		if uint64(len(it)) > c.MaxByteFieldLen {
			return errors.New("byte-assignable field length too large to encode")
		}
		if err := writeUvarint(w, len(it)); err != nil {
			return err
		}
		_, err = w.Write(it)
	case string:
		err = c.EncodeField(w, []byte(item.(string)))
	case [32]byte:
		_, err = w.Write(it[:])
	default:
		err = c.encodeReflect(w, item)
	}

	return err
}

func (c *ConfiguredEncoder) encodeReflect(w io.Writer, item interface{}) error {
	t := reflect.TypeOf(item)

	canonicalized := canonicalizeWellKnown(t)
	if wellKnownEncoders[canonicalized] != nil {
		return wellKnownEncoders[canonicalized](w, item)
	}

	if t.Kind() == reflect.Array {
		itemVal := reflect.ValueOf(item)
		if t.Elem().Kind() == reflect.Uint8 {
			itemPtr := reflect.New(t)
			itemPtr.Elem().Set(itemVal)
			_, err := w.Write(itemPtr.Elem().Slice(0, itemVal.Len()).Bytes())
			return err
		}

		for i := 0; i < itemVal.Len(); i++ {
			if err := c.EncodeField(w, itemVal.Index(i).Interface()); err != nil {
				return err
			}
		}
		return nil
	}

	if t.Kind() == reflect.Slice {
		val := reflect.ValueOf(item)
		if val.Len() > c.MaxVariableArrayLen {
			return errors.New("variable array field length too large to encode")
		}

		if err := writeUvarint(w, val.Len()); err != nil {
			return err
		}
		for i := 0; i < val.Len(); i++ {
			if err := c.EncodeField(w, val.Index(i).Interface()); err != nil {
				return err
			}
		}
		return nil
	}

	return errors.New(fmt.Sprintf("type %s cannot be encoded", t.String()))
}

func writeUvarint(w io.Writer, n int) error {
	lenBuf := make([]byte, binary.MaxVarintLen64, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(lenBuf, uint64(n))
	_, err := w.Write(lenBuf[:bytesWritten])
	return err
}
