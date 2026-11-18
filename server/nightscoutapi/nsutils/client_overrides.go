package nsutils

import "nighttroop.t1d.nz/data/sql/controldb"

// ClientOverrides are where the Nightscout API answers one client differently from every other client.
//
// WARNING: each override here goes against how the API answers every other client, and against what REQUIREMENTS.md
// says the API gives. Add one only when a client cannot work with what every other client gets, and the difference
// cannot be fixed in the client. Say which client needs it, why, and what it would break for the others.
//
// An override is chosen by the client the reading identity was added for (its site identity's client), so it never
// reaches another client's reads. It never applies to the Socket.IO broadcasts, which every reader of a site gets
// alike.
type ClientOverrides struct {
	// FractionalDurations: a treatment's duration is given as the minutes of its milliseconds, with their fraction,
	// e.g. 1.6666666666666667, not in whole minutes as every other client gets it on API v1.
	//
	// BAD: Trio, xDrip4iOS and AAPS read duration as a whole number and drop the whole page if one has a fraction,
	// so this must never be given to them. Only Loop: it sends a dose's or override's minutes with their fraction
	// and its NightscoutKit reads them back as sent, as upstream Nightscout gives them.
	FractionalDurations bool
}

// clientOverrides are the overrides of each client a device can be added for that needs them
var clientOverrides = map[string]ClientOverrides{
	"loop": {FractionalDurations: true},
}

// OverridesFor are the overrides of the client an identity was added for, none for any other identity
func OverridesFor(identity *controldb.SiteIdentity) ClientOverrides {
	if identity == nil || identity.Client == nil {
		return ClientOverrides{}
	}
	return clientOverrides[*identity.Client]
}
