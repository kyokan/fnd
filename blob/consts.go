package blob

const (
	HeaderLen      = 1 + 1 + 63 + 8 + 32 + 65
	SectorLen      = 4096
	SectorCount    = 256
	Size           = SectorLen * SectorCount
	CurrentVersion = 1
)
