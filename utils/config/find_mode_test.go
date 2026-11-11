package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/consts"
)

// A MODE that is unset, unknown, or test outside go test stops the server starting, rather than falling back to
// development, whose passwords, TOTP and keys are known
func TestFindMode(t *testing.T) {
	tests := []struct {
		name        string
		flag        *string
		env         *string
		underGoTest bool
		mode        consts.ServerMode
		err         string
	}{
		{name: "unset outside go test", err: "MODE is not set"},
		{name: "unset under go test", underGoTest: true, mode: consts.Test},
		{name: "production flag", flag: new("production"), mode: consts.Production},
		{name: "development flag", flag: new("development"), mode: consts.Development},
		{name: "staging env", env: new("staging"), mode: consts.Staging},
		{name: "flag over env", flag: new("prod"), env: new("dev"), mode: consts.Production},
		{name: "misspelt flag", flag: new("producton"), err: "invalid MODE"},
		{name: "misspelt env", env: new("producton"), err: "invalid MODE"},
		{name: "misspelt flag under go test", flag: new("producton"), underGoTest: true, err: "invalid MODE"},
		{name: "empty env", env: new(""), err: "invalid MODE"},
		{name: "test flag outside go test", flag: new("test"), err: "only for go test"},
		{name: "test env outside go test", env: new("test"), err: "only for go test"},
		{name: "test env under go test", env: new("test"), underGoTest: true, mode: consts.Test},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MODE", "") // restores MODE after the test
			if tt.env != nil {
				t.Setenv("MODE", *tt.env)
			} else {
				require.NoError(t, os.Unsetenv("MODE"))
			}
			flags := map[string]string{}
			if tt.flag != nil {
				flags["MODE"] = *tt.flag
			}

			mode, err := findMode(flags, tt.underGoTest)
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.mode, mode)
		})
	}
}

// Development has known passwords and TOTP, so it is only for local use, over http
func TestDevelopmentOnlyOverHTTP(t *testing.T) {
	tests := []struct {
		mode     consts.ServerMode
		protocol string
		err      string
	}{
		{consts.Development, "http", ""},
		{consts.Development, "https", "development mode is only for local use"},
		{consts.Test, "http", ""},
		{consts.Test, "https", ""},
		{consts.Production, "https", ""},
		{consts.Production, "http", "http is only allowed"},
	}
	for _, tt := range tests {
		t.Run(string(tt.mode)+" "+tt.protocol, func(t *testing.T) {
			cfg, err := readServerConfig(tt.mode, map[string]string{"HOST_PROTOCOL": tt.protocol})
			require.NoError(t, err)
			err = cfg.ValidateClean()
			if tt.err == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.err)
			}
		})
	}
}
