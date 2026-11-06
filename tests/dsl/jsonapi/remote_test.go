package jsonapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type recordTransport struct{ urls []string }

func (r *recordTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.urls = append(r.urls, req.URL.String()+" host="+req.Host)
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Request: req}, nil
}

// A script's localhost URLs go to the same names under the remote host, over its scheme, and any other as it is
func TestRemoteClientMapsLocalhost(t *testing.T) {
	recorded := &recordTransport{}
	client := RemoteClient("https", "staging.example.com")
	client.Transport.(*remote).next = recorded
	for _, url := range []string{
		"http://localhost/site",
		"http://abcdefgh.localhost:1234/login?x=1",
		"http://abcdefgh.ns.localhost/api/v1/entries",
		"https://example.org/other",
	} {
		_, err := client.Get(url)
		require.NoError(t, err)
	}
	require.Equal(t, []string{
		"https://staging.example.com/site host=",
		"https://abcdefgh.staging.example.com/login?x=1 host=",
		"https://abcdefgh.ns.staging.example.com/api/v1/entries host=",
		"https://example.org/other host=example.org",
	}, recorded.urls)
}
