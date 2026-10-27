package jsonapi

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestIncludeRunsAnotherFilesRequests(t *testing.T) {
	fsys := fstest.MapFS{
		"steps/site.steps": {Data: []byte("NAME create_site\nPOST http://localhost/site\nASSERT_STATUS 200\nSET token {{ .create_site.token }}\n")},
		"api/list.http":    {Data: []byte("INCLUDE steps/site.steps\nSET site {{ .create_site.subdomain }}\n\nNAME list\nGET http://localhost/list?token={{ token }}\nASSERT_STATUS 200\n")},
		"api/missing.http": {Data: []byte("NAME list\nGET http://localhost/list\nINCLUDE steps/missing.steps\n")},
	}

	file, err := ParseRequestFile(fsys, "api/list.http")
	require.NoError(t, err)
	require.Equal(t, "list", file.Name)
	require.Len(t, file.Requests, 2)
	require.Equal(t, "create_site", file.Requests[0].Name)
	// A SET after the INCLUDE goes on the included file's last request
	require.Len(t, file.Requests[0].Sets, 2)
	require.Equal(t, "list", file.Requests[1].Name)

	_, err = ParseRequestFile(fsys, "api/missing.http")
	require.ErrorContains(t, err, "api/missing.http:3: INCLUDE steps/missing.steps")
}
