package static_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server/assets/static"
)

func router() http.Handler {
	r := chi.NewRouter()
	static.AddStaticRoutes(r)
	return r
}

func get(handler http.Handler, path string, acceptEncoding string, ifNoneMatch string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if acceptEncoding != "" {
		req.Header.Set("Accept-Encoding", acceptEncoding)
	}
	if ifNoneMatch != "" {
		req.Header.Set("If-None-Match", ifNoneMatch)
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

// Static files are sent as they are, even to a client that accepts gzip, as the proxy in front compresses them
func TestStaticFilesAreNotCompressed(t *testing.T) {
	handler := router()
	tests := []struct {
		name string
		path string
		body []byte
	}{
		{"a script", "/echarts.min.js", static.StaticEchartsMinJs},
		{"a style", "/style_gen.css", static.StaticStyleGenCss},
		{"an svg", "/logo.svg", static.StaticLogoSvg},
		{"a png", "/logo.png", static.StaticLogoPng},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := get(handler, test.path, "gzip, deflate, br", "")
			require.Equal(t, http.StatusOK, res.Code)
			require.Empty(t, res.Header().Get("Content-Encoding"))
			require.Equal(t, test.body, res.Body.Bytes())

			// A browser that has the file is told it is unchanged
			res = get(handler, test.path, "gzip", res.Header().Get("ETag"))
			require.Equal(t, http.StatusNotModified, res.Code)
		})
	}
}

// BenchmarkGetEcharts is a browser's first download of the largest static file, a 1MB script every site page loads
func BenchmarkGetEcharts(b *testing.B) {
	handler := router()
	for b.Loop() {
		res := get(handler, "/echarts.min.js", "gzip", "")
		if res.Code != http.StatusOK {
			b.Fatalf("status %d", res.Code)
		}
	}
}
