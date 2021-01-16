package dwire

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type cafeEncodeDecoder struct {
	data []byte
}

func (c *cafeEncodeDecoder) Decode(r io.Reader) error {
	buf := make([]byte, 2, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	if !bytes.Equal(buf, []byte{0xca, 0xfe}) {
		return errors.New("invalid cafe decode")
	}
	c.data = buf
	return nil
}

func (c *cafeEncodeDecoder) Encode(w io.Writer) error {
	_, err := w.Write([]byte{0xca, 0xfe})
	return err
}

func TestDecodeFields(t *testing.T) {
	var cafe cafeEncodeDecoder

	type testStruct struct {
		f0  cafeEncodeDecoder
		f1  uint8
		f2  uint16
		f3  uint32
		f4  uint64
		f5  []byte
		f6  string
		f7  [32]byte
		f8  [2]uint8
		f9  []uint8
		f10 []string
		f11 time.Time
		f12 [2]string
		f13 []*cafeEncodeDecoder
	}
	exp := &testStruct{
		f0: cafe,
		f1: 1,
		f2: 2,
		f3: 3,
		f4: 4,
		f5: []byte{0xff, 0x00},
		f6: "testing",
		f7: [32]byte{},
		f8: [2]uint8{
			1,
			2,
		},
		f9: []uint8{
			3,
			4,
		},
		f10: []string{
			"testing",
			"testing",
		},
		f11: time.Unix(1, 0),
		f12: [2]string{
			"testing",
			"testing",
		},
		f13: []*cafeEncodeDecoder{
			&cafe,
			&cafe,
		},
	}
	exp.f7[0] = 0x11

	var actual testStruct
	inputBytes, err := hex.DecodeString(
		"cafe" +
			"01" +
			"0002" +
			"00000003" +
			"0000000000000004" +
			"02ff00" +
			"0774657374696e67" +
			"1100000000000000000000000000000000000000000000000000000000000000" +
			"0102" +
			"020304" +
			"020774657374696e670774657374696e67" +
			"00000001" +
			"0774657374696e670774657374696e67" +
			"02cafecafe",
	)
	require.NoError(t, err)
	require.NoError(t, DecodeFields(
		bytes.NewReader(inputBytes),
		&actual.f0,
		&actual.f1,
		&actual.f2,
		&actual.f3,
		&actual.f4,
		&actual.f5,
		&actual.f6,
		&actual.f7,
		&actual.f8,
		&actual.f9,
		&actual.f10,
		&actual.f11,
		&actual.f12,
		&actual.f13,
	))
	require.EqualValues(t, exp.f0.data, exp.f0.data)
	require.EqualValues(t, exp.f1, actual.f1)
	require.EqualValues(t, exp.f2, actual.f2)
	require.EqualValues(t, exp.f3, actual.f3)
	require.EqualValues(t, exp.f4, actual.f4)
	require.EqualValues(t, exp.f5, actual.f5)
	require.EqualValues(t, exp.f6, actual.f6)
	require.EqualValues(t, exp.f7, actual.f7)
	require.EqualValues(t, exp.f8, actual.f8)
	require.EqualValues(t, exp.f9, actual.f9)
	require.EqualValues(t, exp.f10, actual.f10)
	require.EqualValues(t, exp.f11, actual.f11)
	require.EqualValues(t, exp.f12, actual.f12)
	require.Equal(t, len(exp.f13), len(actual.f13))
}

func TestDecode_Errors(t *testing.T) {
	var boolVal bool
	err := DecodeField(bytes.NewReader([]byte{0x02}), &boolVal)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid boolean value")

	err = DecodeField(bytes.NewReader([]byte{}), uint64(0))
	require.Error(t, err)
	require.Contains(t, err.Error(), "can only decode into pointer types")

	var buf bytes.Buffer
	require.NoError(t, writeUvarint(&buf, DefaultMaxByteFieldLen+DefaultMaxVariableArrayLen+1))
	var byteArrVal []byte
	err = DecodeField(bytes.NewReader(buf.Bytes()), &byteArrVal)
	require.Error(t, err)
	require.Contains(t, err.Error(), "byte-assignable field length too large to decode")

	var strVal string
	// zero out err since the err message is the same
	err = nil
	err = DecodeField(bytes.NewReader(buf.Bytes()), &strVal)
	require.Error(t, err)
	require.Contains(t, err.Error(), "byte-assignable field length too large to decode")

	var strArrVal []string
	err = DecodeField(bytes.NewReader(buf.Bytes()), &strArrVal)
	require.Error(t, err)
	require.Contains(t, err.Error(), "variable array field length too large to decode")
}
