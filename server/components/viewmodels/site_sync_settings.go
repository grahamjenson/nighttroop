package viewmodels

type SiteSyncSettings struct {
	NSURL              string
	NSPasswordGarbage  string
	NSEnabled          bool
	NSID               *int64
	DexUsername        string
	DexPasswordGarbage string
	DexUSARegion       bool
	DexEnabled         bool
	DexID              *int64
}
