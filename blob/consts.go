package blob

const (
	HeaderLen      = 1 + 1 + 63 + 8 + 32 + 65
	SectorLen      = 65536
	SectorCount    = 256
	Size           = SectorLen * SectorCount
	CurrentVersion = 1
)
