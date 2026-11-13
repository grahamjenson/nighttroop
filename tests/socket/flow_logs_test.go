package socket_test

import (
	"nighttroop.t1d.nz/tests/dsl/flow"
	"nighttroop.t1d.nz/utils/aslog"
)

// The test servers' logs are kept, and shown only with a flow that fails
func init() { aslog.TestOutput = flow.ServerLogs }
