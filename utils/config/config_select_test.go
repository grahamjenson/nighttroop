package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigSelectUsesDefaultForBadValues(t *testing.T) {
	dotenv := map[string]string{
		"TEST_GOOD_INT": "5", "TEST_BAD_INT": "lots",
		"TEST_GOOD_FLOAT": "1.5", "TEST_BAD_FLOAT": "lots",
	}
	require.Equal(t, 5, configIntSelect(dotenv, "TEST_GOOD_INT", 7))
	require.Equal(t, 7, configIntSelect(dotenv, "TEST_BAD_INT", 7))
	require.Equal(t, 1.5, configFloatSelect(dotenv, "TEST_GOOD_FLOAT", 2.5))
	require.Equal(t, 2.5, configFloatSelect(dotenv, "TEST_BAD_FLOAT", 2.5))
}

func TestConfigSelectPrefersTheEnvironmentAndRemovesWhatItReads(t *testing.T) {
	t.Setenv("TEST_ENV_INT", "9")
	dotenv := map[string]string{"TEST_ENV_INT": "5", "TEST_DOTENV_STRING": "dotenv", "TEST_UNKNOWN": "x"}
	require.Equal(t, 9, configIntSelect(dotenv, "TEST_ENV_INT", 7))
	require.Equal(t, "dotenv", configStringSelect(dotenv, "TEST_DOTENV_STRING", "default"))
	require.Equal(t, 0.1, configFloatSelect(dotenv, "TEST_MISSING_FLOAT", 0.1))
	require.Equal(t, map[string]string{"TEST_UNKNOWN": "x"}, dotenv)
	require.Error(t, rejectUnknownVariables(dotenv))
	require.NoError(t, rejectUnknownVariables(map[string]string{}))
}
