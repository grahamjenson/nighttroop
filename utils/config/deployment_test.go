package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/consts"
)

// A server is the development deployment unless named, e.g. one of several staging servers, whatever its mode
func TestDeployment(t *testing.T) {
	cfg, err := readServerConfig(consts.Development, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, "development", cfg.Deployment)

	cfg, err = readServerConfig(consts.Staging, map[string]string{"DEPLOYMENT": "nighttroop-staging"})
	require.NoError(t, err)
	require.Equal(t, "nighttroop-staging", cfg.Deployment)
}
