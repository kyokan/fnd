package blob

const (
	HeaderLen      = 1 + 1 + 63 + 8 + 32 + 65
	SectorBytes      = 256
	MaxSectors    = 4096
	Size           = SectorBytes * MaxSectors
	CurrentVersion = 1
)
