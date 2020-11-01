package primitives

import (
	"errors"
	"fnd.localhost/handshake/encoding"
	"golang.org/x/crypto/blake2b"
)

func HashName(name string) []byte {
	h, _ := blake2b.New256(nil)
	h.Write([]byte(name))
	return h.Sum(nil)
}

func CreateBlind(value uint64, nonce []byte) ([]byte, error) {
	if len(nonce) != 32 {
		return nil, errors.New("nonce must be 32 bytes long")
	}

	h, _ := blake2b.New256(nil)
	_ = encoding.WriteUint64(h, value)
	h.Write(nonce)
	return h.Sum(nil), nil
}

const (
	MaxNameLen = 63
)

var validCharset = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0,
	0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 0, 0, 0, 0, 4,
	0, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 0, 0, 0, 0, 0,
}

var blacklist = map[string]bool{
	"example":   true,
	"invalid":   true,
	"local":     true,
	"localhost": true,
	"test":      true,
}

func ValidateName(name string) error {
	if len(name) == 0 {
		return errors.New("name must have nonzero length")
	}

	if len(name) > MaxNameLen {
		return errors.New("name over maximum length")
	}

	for i := 0; i < len(name); i++ {
		ch := name[i]

		if int(ch) > len(validCharset) {
			return errors.New("invalid character")
		}

		charType := validCharset[ch]
		switch charType {
		case 0:
			return errors.New("invalid character")
		case 1:
			continue
		case 2:
			return errors.New("name cannot contain capital letters")
		case 3:
			continue
		case 4:
			if i == 0 {
				return errors.New("name cannot start with a hyphen")
			}
			if i == len(name)-1 {
				return errors.New("name cannot end with a hyphen")
			}
		}
	}

	if blacklist[name] {
		return errors.New("name is blacklisted")
	}

	return nil
}
