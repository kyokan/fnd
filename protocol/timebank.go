package protocol

import (
	"time"
)

type TimebankParams struct {
	TimebankDuration     time.Duration
	MinUpdateInterval    time.Duration
	FullUpdatesPerPeriod int
}

func CheckTimebank(params *TimebankParams, prevUpdateTime time.Time, prevTimebank int, sectorsNeeded int) int {
	if sectorsNeeded == 0 {
		return -1
	}
	if sectorsNeeded > 256 {
		return -1
	}

	now := time.Now()
	if prevUpdateTime.After(now.Add(-1 * params.MinUpdateInterval)) {
		return -1
	}

	sectorUpdatesPerPeriod := params.FullUpdatesPerPeriod * 256
	secondsSince := int(time.Since(prevUpdateTime) / time.Second)
	secondsPerSector := int(params.TimebankDuration/time.Second) / sectorUpdatesPerPeriod
	sectorsAvailable := prevTimebank + (secondsSince / secondsPerSector)
	if sectorsAvailable > sectorUpdatesPerPeriod {
		sectorsAvailable = sectorUpdatesPerPeriod
	}

	if sectorsNeeded > sectorsAvailable {
		return -1
	}

	return sectorsAvailable - sectorsNeeded
}
