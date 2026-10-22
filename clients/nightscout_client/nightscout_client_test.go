package nightscout_client

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseResultRefusesLargeBodies(t *testing.T) {
	// A few MB of spaces inside the object inflate well from a small gzip body
	big := `{"result":[` + strings.Repeat(" ", 3<<20) + `]}`
	var gz bytes.Buffer
	w := gzip.NewWriter(&gz)
	_, err := w.Write([]byte(big))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	tests := []struct {
		name     string
		body     []byte
		encoding string
	}{
		{"plain", []byte(big), ""},
		{"gzip", gz.Bytes(), "gzip"},
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

			result, err := parseResult(resp.Body)
			require.Error(t, err)
			require.Nil(t, result)
		})
	}
}

func TestEntriesFromResultRefusesTooManyEntries(t *testing.T) {
	result := make([]any, 0, maxResultItems+1)
	for i := 0; i < maxResultItems; i++ {
		result = append(result, map[string]any{"sgv": 100})
	}
	entries, err := entriesFromResult(result)
	require.NoError(t, err)
	require.Len(t, entries, maxResultItems)

	entries, err = entriesFromResult(append(result, map[string]any{"sgv": 100}))
	require.Error(t, err)
	require.Nil(t, entries)
}
