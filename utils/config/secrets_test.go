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
		{"production with S3 over HTTP", consts.Production, map[string]string{"S3_URL": "http://s3.example.com", "S3_ACCESS_KEY": "an access key", "S3_SECRET_KEY": "a secret key"}, "HTTPS"},
		{"production without S3", consts.Production, map[string]string{}, ""},
		{"production with S3 and no access key", consts.Production, map[string]string{"S3_URL": "https://s3.example.com", "S3_SECRET_KEY": "a secret key"}, "S3_ACCESS_KEY"},
		{"production with S3 and no secret key", consts.Production, map[string]string{"S3_URL": "https://s3.example.com", "S3_ACCESS_KEY": "an access key"}, "S3_SECRET_KEY"},
		{"staging with S3 and minioadmin", consts.Staging, map[string]string{"S3_URL": "https://s3.example.com", "S3_ACCESS_KEY": "minioadmin", "S3_SECRET_KEY": "minioadmin"}, "S3_ACCESS_KEY"},
		{"production with S3 and a minioadmin secret", consts.Production, map[string]string{"S3_URL": "https://s3.example.com", "S3_ACCESS_KEY": "an access key", "S3_SECRET_KEY": "minioadmin"}, "S3_SECRET_KEY"},
		{"production with S3 and an empty access key", consts.Production, map[string]string{"S3_URL": "https://s3.example.com", "S3_ACCESS_KEY": "", "S3_SECRET_KEY": "a secret key"}, "S3_ACCESS_KEY"},
		{"production with S3 and an empty bucket", consts.Production, map[string]string{"S3_URL": "https://s3.example.com", "S3_ACCESS_KEY": "an access key", "S3_SECRET_KEY": "a secret key", "S3_BUCKET": ""}, "S3_BUCKET"},
		{"development with S3 and an empty bucket", consts.Development, map[string]string{"S3_URL": "http://localhost:9000", "S3_BUCKET": ""}, ""},
		{"development with S3 and the minioadmin defaults", consts.Development, map[string]string{"S3_URL": "http://localhost:9000"}, ""},
		{"test with S3 and the minioadmin defaults", consts.Test, map[string]string{"S3_URL": "http://localhost:9000"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dotenv := map[string]string{"HOST_PROTOCOL": "https", "OPERATOR_NAME": "Example Health", "SUPPORT_EMAIL": "help@example.org",
				"STAGING_ADMIN_PASSWORD": "a long staging password", "STAGING_ADMIN_TOTP_SECRET": "JBSWY3DPEHPK3PXP"}
			// Development is only for local use, over http
			if tt.mode == consts.Development {
				dotenv["HOST_PROTOCOL"] = "http"
			}
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

// A server in any mode does not start with a time or a retention under 1: 0 minutes deletes each new site at once
func TestValidateTimesAreAtLeastOne(t *testing.T) {
	for _, name := range []string{"DELETE_AFTER_INACTIVITY_MINS", "SITE_BACKUP_INTERVAL_IN_SECONDS", "SITE_BACKUP_RETENTION_DAYS", "CONTROL_BACKUP_RETENTION_DAYS"} {
		for value, ok := range map[string]bool{"-1": false, "0": false, "1": true} {
			t.Run(name+"="+value, func(t *testing.T) {
				cfg, err := readServerConfig(consts.Test, map[string]string{"OPERATOR_NAME": "Example Health", "SUPPORT_EMAIL": "help@example.org", name: value})
				require.NoError(t, err)
				if err = cfg.ValidateClean(); ok {
					require.NoError(t, err)
				} else {
					require.ErrorContains(t, err, name+" must be at least 1, not "+value)
				}
			})
		}
	}
}
