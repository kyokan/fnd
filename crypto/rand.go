package crypto

import "crypto/rand"

func Rand32() [32]byte {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(err)
	}
	return buf
}
