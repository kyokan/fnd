package protocol

import (
	"time"

	"github.com/mslipper/handshake/primitives"
)

// 2020 Jan 1 00:00 UTC
var epochDate = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

const (
	secondsPerHour = 60 * 60
	hoursPerWeek   = 7 * 24
	weekDuration   = time.Duration(hoursPerWeek * time.Hour)
)

func modBuffer(b []byte, n int) int {
	p := 256 % n
	acc := 0

	for i := 0; i < len(b); i++ {
		acc = (p*acc + int(b[i])) % n
	}
	return acc
}

func CurrentEpoch(name string) uint16 {
	hash := primitives.HashName(name)
	mod := modBuffer(hash, hoursPerWeek)
	offset := mod * secondsPerHour
	startDate := epochDate.Add(time.Duration(offset) * time.Second)
	return uint16(int(time.Now().Sub(startDate).Seconds()) / int(weekDuration.Seconds()))
}
