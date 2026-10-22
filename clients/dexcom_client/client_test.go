package dexcom_client

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeEGVsRefusesLargeBodies(t *testing.T) {
	// A few MB of spaces inside the array inflate well from a small gzip body
	big := `[{"Value":100}` + strings.Repeat(" ", 3<<20) + `]`
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

			egvs, err := decodeEGVs(resp.Body)
			require.Error(t, err)
			require.Nil(t, egvs)
		})
	}
}

func TestDecodeEGVsRefusesTooManyEGVs(t *testing.T) {
	egv := `{"WT":"Date(1700000000000)","ST":"Date(1700000000000)","DT":"Date(1700000000000)","Value":100,"Trend":"Flat"}`
	ok := "[" + strings.Repeat(egv+",", maxEGVs-1) + egv + "]"
	egvs, err := decodeEGVs(strings.NewReader(ok))
	require.NoError(t, err)
	require.Len(t, egvs, maxEGVs)

	tooMany := "[" + strings.Repeat(egv+",", maxEGVs) + egv + "]"
	egvs, err = decodeEGVs(strings.NewReader(tooMany))
	require.Error(t, err)
	require.Nil(t, egvs)
}
