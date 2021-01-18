package blob

import (
	"bytes"
	"encoding/hex"
	"fnd.localhost/handshake/primitives"
	"path"
)

func PathifyName(n string) string {
	hash := primitives.HashName(n)
	r := bytes.NewReader(hash)
	var segments []string
	for len(segments) < 16 {
		buf := make([]byte, 2, 2)
		_, err := r.Read(buf)
		// should not be possible
		if err != nil {
			panic(err)
		}

		segments = append(segments, hex.EncodeToString(buf))
	}

	if len(segments) != 16 {
		panic("invalid number of path segments - should never happen")
	}

	return path.Join(path.Join(segments[0:15]...), segments[15]+"_blob")
}
