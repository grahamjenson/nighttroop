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
