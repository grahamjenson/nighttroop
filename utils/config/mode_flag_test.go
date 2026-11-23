package config_test

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// Starting with -MODE=prod used to fail, as MODE was left over as an unknown .env variable
func TestModeFlagIsNotAnUnknownVariable(t *testing.T) {
	args := os.Args
	t.Cleanup(func() { os.Args = args })
	os.Args = []string{args[0], "-MODE=test"}

	serverConfig, err := config.NewServerConfig()
	require.NoError(t, err)
	require.Equal(t, consts.Test, serverConfig.Mode)
}

// A server with no KEY_SETS, or with one cut short, does not start and says so as a config error: it used to panic
func TestKeySetsThatCannotBeReadAreAConfigError(t *testing.T) {
	args := os.Args
	t.Cleanup(func() { os.Args = args })
	os.Args = []string{args[0], "-MODE=production"}
	t.Setenv("HOST_PROTOCOL", "https")
	t.Setenv("OPERATOR_NAME", "Example Health")
	t.Setenv("SUPPORT_EMAIL", "help@example.org")

	for _, keySets := range []string{"", "one,two", "a,b,c,d"} {
		t.Setenv("KEY_SETS", keySets)
		// Each start defines its flags, which one process does one time
		flag.CommandLine = flag.NewFlagSet(args[0], flag.ContinueOnError)
		_, err := config.NewServerConfig()
		require.ErrorContains(t, err, "KEY_SETS cannot be read", keySets)
	}
}
