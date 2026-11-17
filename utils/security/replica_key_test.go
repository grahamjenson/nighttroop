package security_test

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tink-crypto/tink-go/v2/insecurecleartextkeyset"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"nighttroop.t1d.nz/utils/testutils"
)

// A replica's SSE-C key is the same each time for its DB, 32 bytes, different for each DB, and not a key of the key
// sets, as it is sent to S3
func TestReplicaKey(t *testing.T) {
	sec := testutils.GetTestServerConfig().Security
	keysets := bytes.Buffer{}
	for _, handle := range []*keyset.Handle{sec.EncryptionKeySet, sec.StreamEncryptionKeySet, sec.JWTKeySet, sec.MACKeySet} {
		require.NoError(t, insecurecleartextkeyset.Write(handle, keyset.NewBinaryWriter(&keysets)))
	}
	keys := map[string]bool{}
	for _, name := range []string{"control", "site/aaaaaaaa", "site/aaaaaaab"} {
		key, err := sec.ReplicaKey(name)
		require.NoError(t, err)
		again, err := sec.ReplicaKey(name)
		require.NoError(t, err)
		require.Equal(t, key, again)
		raw, err := base64.StdEncoding.DecodeString(key)
		require.NoError(t, err)
		require.Len(t, raw, 32)
		require.False(t, bytes.Contains(keysets.Bytes(), raw), "not a key of the key sets")
		keys[key] = true
	}
	require.Len(t, keys, 3)
}
