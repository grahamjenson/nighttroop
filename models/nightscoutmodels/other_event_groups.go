package nightscoutmodels

import (
	"crypto/sha256"
	"encoding/hex"

	datatypes "nighttroop.t1d.nz/data/types"
)

// TreatmentEventGroupOther is what only some apps send of a treatment, kept in its group's other attributes: the app's
// own id for it (xDrip+'s uuid), AAPS's running mode, whether the loop forced it and why, and how long it was to
// last, and Loop's indefinite duration, with a hash of the JSON it was sent as. It is written out, not generated, as a generated converter cannot be called from this package's own code,
// which goverter compiles first.
func TreatmentEventGroupOther(source Treatment) datatypes.EventGroupOther {
	return datatypes.EventGroupOther{Uuid: source.UUID.Stringp(), Mode: source.Mode.Stringp(),
		AutoForced: source.AutoForced, Reasons: source.Reasons.Stringp(), OriginalDurationMs: source.OriginalDurationMs.Int64p(),
		DurationType: source.DurationType.Stringp(), SentHash: SentHash(source.SentJSON)}
}

// SentHash is a short hash of the JSON a treatment was sent as, nil when it was not sent as JSON
func SentHash(sent string) *string {
	if sent == "" {
		return nil
	}
	sum := sha256.Sum256([]byte(sent))
	hash := hex.EncodeToString(sum[:16])
	return &hash
}
