package config_test

import (
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
