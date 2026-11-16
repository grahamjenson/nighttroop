package static_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/utils/templui/utils"
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

// A file's versioned URL is kept by the browser for a year; its plain URL, or one with another version, is checked
// against the ETag each time
func TestVersionedURLsAreKept(t *testing.T) {
	handler := router()
	url := static.URL("/style_gen.css")
	require.Regexp(t, `^/style_gen\.css\?v=[0-9a-f]{16}$`, url)
	tests := []struct {
		name         string
		path         string
		cacheControl string
	}{
		{"its version", url, "public, max-age=31536000, immutable"},
		{"no version", "/style_gen.css", "no-cache"},
		{"another version", "/style_gen.css?v=0123456789abcdef", "no-cache"},
		{"an empty version", "/style_gen.css?v=", "no-cache"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := get(handler, test.path, "", "")
			require.Equal(t, http.StatusOK, res.Code)
			require.Equal(t, test.cacheControl, res.Header().Get("Cache-Control"))
			require.Equal(t, static.StaticStyleGenCss, res.Body.Bytes())
		})
	}
}

// A component's script is linked with its file's version, so the browser keeps it too
func TestComponentScriptsAreKept(t *testing.T) {
	url := utils.ScriptURL("/toast.min.js")
	require.Equal(t, static.URL("/toast.min.js"), url)
	require.Regexp(t, `^/toast\.min\.js\?v=[0-9a-f]{16}$`, url)
	res := get(router(), url, "", "")
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "public, max-age=31536000, immutable", res.Header().Get("Cache-Control"))
}
