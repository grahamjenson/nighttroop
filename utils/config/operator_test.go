package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/consts"
)

// Only development and test default to the project's own operator and support email; production and staging will not
// start until they name who runs them and how to contact them
func TestOperatorConfig(t *testing.T) {
	stagingAdmin := map[string]string{"STAGING_ADMIN_PASSWORD": "a long staging password", "STAGING_ADMIN_TOTP_SECRET": "JBSWY3DPEHPK3PXP"}
	tests := []struct {
		name                    string
		mode                    consts.ServerMode
		operator, email         *string
		wantOperator, wantEmail string
		ok                      bool
	}{
		{"development defaults to the project's", consts.Development, nil, nil, "T1DNZ", "nighttroop@t1d.nz", true},
		{"test defaults to the project's", consts.Test, nil, nil, "T1DNZ", "nighttroop@t1d.nz", true},
		{"development can name its own", consts.Development, new("Example Health"), new("help@example.org"), "Example Health", "help@example.org", true},
		{"production with neither", consts.Production, nil, nil, "", "", false},
		{"production with no operator", consts.Production, nil, new("help@example.org"), "", "help@example.org", false},
		{"production with no support email", consts.Production, new("Example Health"), nil, "Example Health", "", false},
		{"production with a blank operator", consts.Production, new("  "), new("help@example.org"), "  ", "help@example.org", false},
		{"production with both", consts.Production, new("Example Health"), new("help@example.org"), "Example Health", "help@example.org", true},
		{"staging with neither", consts.Staging, nil, nil, "", "", false},
		{"staging with both", consts.Staging, new("T1DNZ"), new("nighttroop@t1d.nz"), "T1DNZ", "nighttroop@t1d.nz", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dotenv := map[string]string{"HOST_PROTOCOL": "https"}
			// Development is only for local use, over http
			if tt.mode == consts.Development {
				dotenv["HOST_PROTOCOL"] = "http"
			}
			if tt.mode == consts.Staging {
				dotenv = MergeMaps(dotenv, stagingAdmin)
			}
			if tt.operator != nil {
				dotenv["OPERATOR_NAME"] = *tt.operator
			}
			if tt.email != nil {
				dotenv["SUPPORT_EMAIL"] = *tt.email
			}
			cfg, err := readServerConfig(tt.mode, dotenv)
			require.NoError(t, err)
			require.NoError(t, rejectUnknownVariables(dotenv))
			require.Equal(t, tt.wantOperator, cfg.OperatorName)
			require.Equal(t, tt.wantEmail, cfg.SupportEmail)
			if err := cfg.ValidateClean(); tt.ok {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, "OPERATOR_NAME and SUPPORT_EMAIL must be set")
			}
		})
	}
}
