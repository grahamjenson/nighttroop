package static_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server/assets/static"
)

// router serves the static files behind the same compression as the site's routers
func router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	static.AddStaticRoutes(r)
	return r
}

func get(handler http.Handler, path string, acceptEncoding string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if acceptEncoding != "" {
		req.Header.Set("Accept-Encoding", acceptEncoding)
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

func gunzip(t *testing.T, body []byte) []byte {
	reader, err := gzip.NewReader(bytes.NewReader(body))
	require.NoError(t, err)
	unzipped, err := io.ReadAll(reader)
	require.NoError(t, err)
	return unzipped
}

func TestStaticFilesAreGzipped(t *testing.T) {
	handler := router()
	tests := []struct {
		name           string
		path           string
		acceptEncoding string
		body           []byte
		gzipped        bool
	}{
		{"scripts are gzipped", "/echarts.min.js", "gzip, deflate, br", static.StaticEchartsMinJs, true},
		{"styles are gzipped", "/style_gen.css", "gzip", static.StaticStyleGenCss, true},
		{"svgs are gzipped", "/logo.svg", "gzip", static.StaticLogoSvg, true},
		{"not gzipped for a client that does not accept it", "/echarts.min.js", "", static.StaticEchartsMinJs, false},
		{"pngs are already compressed", "/logo.png", "gzip", static.StaticLogoPng, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := get(handler, test.path, test.acceptEncoding)
			require.Equal(t, http.StatusOK, res.Code)
			body := res.Body.Bytes()
			if test.gzipped {
				require.Equal(t, "gzip", res.Header().Get("Content-Encoding"))
				require.Equal(t, "Accept-Encoding", res.Header().Get("Vary"))
				body = gunzip(t, body)
			} else {
				require.Empty(t, res.Header().Get("Content-Encoding"))
			}
			require.Equal(t, test.body, body)
		})
	}
}

// BenchmarkGetEcharts is a browser's first download of the largest static file, a 1MB script every site page loads
func BenchmarkGetEcharts(b *testing.B) {
	handler := router()
	for b.Loop() {
		res := get(handler, "/echarts.min.js", "gzip")
		if res.Code != http.StatusOK {
			b.Fatalf("status %d", res.Code)
		}
	}
}
