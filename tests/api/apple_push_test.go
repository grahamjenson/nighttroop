package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"nighttroop.t1d.nz/utils/push"
)

// fakeApplePush sends the server's Apple pushes to a server that takes every one, so no test reaches Apple, as a
// Loop remote command or a test push does once its site has a credential for the app
func fakeApplePush(tb testing.TB) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	server.EnableHTTP2 = true
	server.StartTLS()
	tb.Cleanup(server.Close)
	apns := push.DefaultAPNs
	push.DefaultAPNs = &push.APNs{Client: server.Client(), Host: server.URL}
	tb.Cleanup(func() { push.DefaultAPNs = apns })
}
