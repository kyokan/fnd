package util

import "math/rand"

func SampleIndices(sliceLen int, count int) []int {
	var out []int
	if sliceLen == 0 {
		return out
	}

	if count > sliceLen {
		for i := 0; i < sliceLen; i++ {
			out = append(out, i)
		}
		return out
	}

	usedMap := make(map[int]bool)
	for len(out) < count {
		for i := 0; i <= 100; i++ {
			if i == 100 {
				panic("failed to find random sample in time")
			}

			j := rand.Intn(sliceLen)
			if used := usedMap[j]; !used {
				usedMap[j] = true
				out = append(out, j)
				break
			}
		}
	}

	return out
}
