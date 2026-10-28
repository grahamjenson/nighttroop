package nightscoutapi

import (
	"os"
	"sync/atomic"
	"testing"

	"nighttroop.t1d.nz/utils/config"
)

// testConfig allows 3 GETs a site, then 1 a minute, and 2 unauthenticated requests an IP; a config can only be made
// once per process
var testConfig *config.ServerConfig

func TestMain(m *testing.M) {
	os.Setenv("SITE_GET_PER_MINUTE_LIMIT", "1")
	os.Setenv("SITE_GET_PER_MINUTE_BURST", "3")
	os.Setenv("SITE_UNAUTH_PER_MINUTE_LIMIT", "1")
	os.Setenv("SITE_UNAUTH_PER_MINUTE_BURST", "2")
	cfg, err := config.NewServerConfig()
	if err != nil {
		panic(err)
	}
	testConfig = cfg
	os.Exit(m.Run())
}

// runs makes each test's site new, so -count does not share their limits
var runs atomic.Int64
