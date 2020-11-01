package primitives

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateBlind(t *testing.T) {
	tests := []struct {
		value    uint64
		nonceHex string
		blindHex string
	}{
		{
			410050,
			"fa3858ba5b513da9bc7276820d0d0e678da26ef47632e3d325caf2c20b877453",
			"7694e4a9c3598d64de5f834c5b0bce3d0bd31cdc9dbeb11571aeef4eda853108",
		},
	}
	for _, tt := range tests {
		nonceB, err := hex.DecodeString(tt.nonceHex)
		require.NoError(t, err)
		blindB, err := CreateBlind(tt.value, nonceB)
		require.Equal(t, tt.blindHex, hex.EncodeToString(blindB))
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		errStr string
	}{
		{
			"cannot have zero length",
			"",
			"must have nonzero length",
		},
		{
			"cannot be over MaxNameLen characters",
			"longlonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong",
			"over maximum length",
		},
		{
			"cannot start with a hyphen",
			"-startswithhyphen",
			"cannot start with a hyphen",
		},
		{
			"cannot end with a hyphen",
			"endwithhyphen-",
			"cannot end with a hyphen",
		},
		{
			"cannot contain unicode",
			"我叫鸣字",
			"invalid character",
		},
		{
			"can only contain certain characters",
			"hello!",
			"invalid character",
		},
		{
			"can only contain certain characters",
			"hello.",
			"invalid character",
		},
		{
			"works with valid names",
			"heres-a-valid-name",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.errStr == "" {
				require.Nil(t, ValidateName(tt.input))
				return
			}

			err := ValidateName(tt.input)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errStr)
		})
	}
}
