package controlapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// An address shows only the first letter of its name and of its domain, and what follows its last dot
func TestMaskEmail(t *testing.T) {
	for address, want := range map[string]string{
		"sam@example.com":          "s•••@e•••.com",
		"a@b.nz":                   "a•••@b•••.nz",
		"pat.lee@mail.example.org": "p•••@m•••.org",
		"sam@localhost":            "s•••@l•••",
		"not an address":           "•••",
		"@example.com":             "•••",
		"sam@":                     "•••",
		"":                         "•••",
	} {
		require.Equal(t, want, maskEmail(address), address)
	}
}
