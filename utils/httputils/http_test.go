package httputils

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadLimitedBodyRefusesLargeBodies(t *testing.T) {
	// A JSON object padded with spaces compresses well, like a gzip bomb would
	big := `{"a":` + strings.Repeat(" ", 3<<20) + `1}`
	var gz bytes.Buffer
	w := gzip.NewWriter(&gz)
	_, err := w.Write([]byte(big))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	require.Less(t, gz.Len(), 1<<16, "the gzip body is small on the wire")

	tests := []struct {
		name     string
		body     []byte
		encoding string
		max      int64
		wantErr  bool
	}{
		{"under the limit", []byte(`{"a":1}`), "", 1 << 20, false},
		{"plain over the limit", []byte(big), "", 1 << 20, true},
		{"gzip inflating over the limit", gz.Bytes(), "gzip", 1 << 20, true},
		{"gzip inflating under the limit", gz.Bytes(), "gzip", 4 << 20, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.encoding != "" {
					w.Header().Set("Content-Encoding", tt.encoding)
				}
				_, _ = w.Write(tt.body)
			}))
			defer server.Close()
			resp, err := http.Get(server.URL)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := ReadLimitedBody(resp.Body, tt.max)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, body)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, body)
		})
	}
}
