package security_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tink-crypto/tink-go/v2/aead"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

// A ciphertext decrypts only for the purpose and site it was encrypted for, so one site's Dexcom password cannot be
// copied in as another site's, or as a TOTP secret; one encrypted before ciphertexts were bound still decrypts
func TestDecryptForOnlyItsPurposeAndSite(t *testing.T) {
	cfg := security.NewSecurityConfig("", consts.Test, "localhost", false)
	ciphertext, err := cfg.EncryptFor(security.PurposeConnector, "sitea", "dexcom password")
	require.NoError(t, err)

	tests := []struct {
		name    string
		purpose security.Purpose
		siteID  string
		ok      bool
	}{
		{"its purpose and site", security.PurposeConnector, "sitea", true},
		{"another site", security.PurposeConnector, "siteb", false},
		{"no site", security.PurposeConnector, "", false},
		{"another purpose", security.PurposeTOTP, "sitea", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plaintext, err := cfg.DecryptFor(tt.purpose, tt.siteID, ciphertext)
			if tt.ok {
				require.NoError(t, err)
				require.Equal(t, "dexcom password", plaintext)
			} else {
				require.Error(t, err)
			}
		})
	}

	unbound, err := aead.New(cfg.EncryptionKeySet)
	require.NoError(t, err)
	older, err := unbound.Encrypt([]byte("older password"), nil)
	require.NoError(t, err)
	plaintext, err := cfg.DecryptFor(security.PurposeConnector, "sitea", base64.RawStdEncoding.EncodeToString(older))
	require.NoError(t, err)
	require.Equal(t, "older password", plaintext)

	_, err = cfg.DecryptFor(security.PurposeConnector, "sitea", "not a ciphertext")
	require.Error(t, err)
}
