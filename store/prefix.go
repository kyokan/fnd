package store

import (
	"strings"
)

func Prefixer(prefix string) func(k ...string) []byte {
	return func(parts ...string) []byte {
		k := strings.Join(append([]string{prefix}, parts...), "/")
		return []byte(k)
	}
}
