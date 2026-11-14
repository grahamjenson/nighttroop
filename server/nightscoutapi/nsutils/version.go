package nsutils

// The versions every endpoint reports, so a client sees one Nightscout, as Nightscout reports its package.json
// version and its API v3 version
const (
	// Version is the Nightscout version: AAPS 3.4 warns of a Nightscout under 15.0.0, and xDrip+ reads it too
	Version = "15.0.9"
	// VersionNum is Version as Nightscout's socket status writes it: 10000 for each major, 100 for each minor
	VersionNum = 150009
	// APIv3Version is the version of API v3, not of Nightscout
	APIv3Version = "3.0.4"
)
