package nightscoutmodels

import (
	"crypto/sha256"
	"encoding/hex"
)

// SentHash is a short hash of the JSON a treatment was sent as, nil when it was not sent as JSON
func SentHash(sent string) *string {
	if sent == "" {
		return nil
	}
	sum := sha256.Sum256([]byte(sent))
	hash := hex.EncodeToString(sum[:16])
	return &hash
}
