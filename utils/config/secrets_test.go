package config

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tink-crypto/tink-go/v2/aead"
	"github.com/tink-crypto/tink-go/v2/insecurecleartextkeyset"
	"github.com/tink-crypto/tink-go/v2/jwt"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/mac"
	tinkpb "github.com/tink-crypto/tink-go/v2/proto/tink_go_proto"
	"github.com/tink-crypto/tink-go/v2/streamingaead"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

// newKeySets are four new key sets, as cmd/keysets/create_keysets.go makes them
func newKeySets(t *testing.T) []string {
	keySets := []string{}
	for _, template := range []*tinkpb.KeyTemplate{aead.AES256GCMKeyTemplate(), streamingaead.AES256GCMHKDF4KBKeyTemplate(),
		jwt.HS256Template(), mac.HMACSHA256Tag256KeyTemplate()} {
		handle, err := keyset.NewHandle(template)
		require.NoError(t, err)
		buf := new(bytes.Buffer)
		require.NoError(t, insecurecleartextkeyset.Write(handle, keyset.NewBinaryWriter(buf)))
		keySets = append(keySets, base64.RawStdEncoding.EncodeToString(buf.Bytes()))
	}
	return keySets
}

// Outside development and test, the server will not start with the public dev and test keys, or the dev MinIO's
// minioadmin S3 credentials
func TestValidateSecrets(t *testing.T) {
	fresh := newKeySets(t)
	freshKeySets := strings.Join(fresh, ",")
	devKeySets := strings.Split(security.DevKeySets, ",")
	// A new key set for each but the JWT key set, which is the public one that signs control logins
	oneDevKeySet := strings.Join([]string{fresh[0], fresh[1], devKeySets[2], fresh[3]}, ",")
	s3 := map[string]string{"S3_URL": "https://s3.example.com", "S3_ACCESS_KEY": "an access key", "S3_SECRET_KEY": "a secret key"}

	tests := []struct {
		name    string
		mode    consts.ServerMode
		dotenv  map[string]string
		wantErr string
	}{
		{"production with the dev keys", consts.Production, map[string]string{"KEY_SETS": security.DevKeySets}, "KEY_SETS"},
		{"staging with the dev keys", consts.Staging, map[string]string{"KEY_SETS": security.DevKeySets}, "KEY_SETS"},
		{"production with one dev key set", consts.Production, map[string]string{"KEY_SETS": oneDevKeySet}, "KEY_SETS"},
		{"production with new keys", consts.Production, map[string]string{"KEY_SETS": freshKeySets}, ""},
		{"development with the dev keys", consts.Development, map[string]string{"KEY_SETS": security.DevKeySets}, ""},
		{"test with the dev keys", consts.Test, map[string]string{"KEY_SETS": security.DevKeySets}, ""},

		{"production with S3 keys", consts.Production, s3, ""},
		{"production without S3", consts.Production, map[string]string{}, ""},
		{"production with S3 and no access key", consts.Production, map[string]string{"S3_URL": "https://s3.example.com", "S3_SECRET_KEY": "a secret key"}, "S3_ACCESS_KEY"},
		{"production with S3 and no secret key", consts.Production, map[string]string{"S3_URL": "https://s3.example.com", "S3_ACCESS_KEY": "an access key"}, "S3_SECRET_KEY"},
		{"staging with S3 and minioadmin", consts.Staging, map[string]string{"S3_URL": "https://s3.example.com", "S3_ACCESS_KEY": "minioadmin", "S3_SECRET_KEY": "minioadmin"}, "S3_ACCESS_KEY"},
		{"production with S3 and a minioadmin secret", consts.Production, map[string]string{"S3_URL": "https://s3.example.com", "S3_ACCESS_KEY": "an access key", "S3_SECRET_KEY": "minioadmin"}, "S3_SECRET_KEY"},
		{"production with S3 and an empty access key", consts.Production, map[string]string{"S3_URL": "https://s3.example.com", "S3_ACCESS_KEY": "", "S3_SECRET_KEY": "a secret key"}, "S3_ACCESS_KEY"},
		{"development with S3 and the minioadmin defaults", consts.Development, map[string]string{"S3_URL": "http://localhost:9000"}, ""},
		{"test with S3 and the minioadmin defaults", consts.Test, map[string]string{"S3_URL": "http://localhost:9000"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dotenv := map[string]string{"HOST_PROTOCOL": "https",
				"STAGING_ADMIN_PASSWORD": "a long staging password", "STAGING_ADMIN_TOTP_SECRET": "JBSWY3DPEHPK3PXP"}
			for key, value := range tt.dotenv {
				dotenv[key] = value
			}
			cfg, err := readServerConfig(tt.mode, dotenv)
			require.NoError(t, err)
			err = cfg.ValidateClean()
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}
