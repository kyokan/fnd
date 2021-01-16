package blob

const (
	HeaderLen      = 1 + 1 + 63 + 8 + 32 + 65
	SectorLen      = 256
	SectorCount    = 4096
	Size           = SectorLen * SectorCount
	CurrentVersion = 1
)
