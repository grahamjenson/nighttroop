package httputils

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// A response is gzipped only when it is compressMinSize bytes or more, or streamed, and keeps its status either way
func TestCompressSkipsSmallResponses(t *testing.T) {
	small := `{"message":"ok"}`
	large := "[" + strings.Repeat(`{"sgv":120,"direction":"Flat"},`, 60) + `{"sgv":120}]`
	tests := []struct {
		name    string
		status  int
		writes  []string
		flush   bool
		gzipped bool
	}{
		{"a small reply is sent as it is", http.StatusCreated, []string{small}, false, false},
		{"a reply written in small parts that adds up to the minimum is gzipped", http.StatusOK, []string{large[:600], large[600:]}, false, true},
		{"a large reply is gzipped", http.StatusOK, []string{large}, false, true},
		{"a small reply that is flushed is streamed, gzipped", http.StatusOK, []string{small}, true, true},
		{"an empty reply keeps its status", http.StatusNoContent, nil, false, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := compress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.status)
				for _, write := range test.writes {
					_, err := w.Write([]byte(write))
					require.NoError(t, err)
				}
				if test.flush {
					require.NoError(t, http.NewResponseController(w).Flush())
				}
			}))
			r := httptest.NewRequest(http.MethodGet, "/api/v1/entries", nil)
			r.Header.Set("Accept-Encoding", "gzip")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)

			require.Equal(t, test.status, w.Code)
			body := w.Body.String()
			if test.gzipped {
				require.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
				reader, err := gzip.NewReader(w.Body)
				require.NoError(t, err)
				unzipped, err := io.ReadAll(reader)
				require.NoError(t, err)
				body = string(unzipped)
			} else {
				require.Empty(t, w.Header().Get("Content-Encoding"))
			}
			require.Equal(t, strings.Join(test.writes, ""), body)
		})
	}
}

// A small reply written before a panic is still sent, as it was before small replies were held
func TestCompressSendsASmallReplyWrittenBeforeAPanic(t *testing.T) {
	handler := compress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
		panic("after the reply")
	}))
	w := httptest.NewRecorder()
	require.Panics(t, func() { handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil)) })
	require.Equal(t, http.StatusCreated, w.Code)
	require.Equal(t, "ok", w.Body.String())
}

// A WebSocket's upgrade can still take over the connection through the compression
func TestCompressKeepsHijacking(t *testing.T) {
	hijacked := false
	server := httptest.NewServer(compress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hijacker, ok := w.(http.Hijacker)
		require.True(t, ok)
		conn, _, err := hijacker.Hijack()
		require.NoError(t, err)
		hijacked = true
		_, _ = conn.Write([]byte("HTTP/1.1 204 No Content\r\n\r\n"))
		conn.Close()
	})))
	defer server.Close()
	res, err := http.Get(server.URL)
	require.NoError(t, err)
	res.Body.Close()
	require.True(t, hijacked)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
}
