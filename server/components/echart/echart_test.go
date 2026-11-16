package echart

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// A chart's script runs in withECharts, so it waits for ECharts, and its element is in the box with the given id
func TestChartRunsItsScriptOnceEChartsLoads(t *testing.T) {
	tests := []struct {
		name   string
		script string
		want   string
	}{
		{"go-echarts' script", "<script type=\"text/javascript\">\n    echarts.init(x);\n</script>", "<script>withECharts(() => {\n    echarts.init(x);\n})</script>"},
		{"space around it", "\n <script type=\"text/javascript\">draw()</script>\n", "<script>withECharts(() => {draw()})</script>"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, whenLoaded(test.script))

			var html bytes.Buffer
			require.NoError(t, Chart("box", "chart1", test.script).Render(context.Background(), &html))
			require.Contains(t, html.String(), `<div id="box"`)
			require.Contains(t, html.String(), `<div id="chart1"`)
			require.Contains(t, html.String(), test.want)
		})
	}
}

// The loader names ECharts' versioned URL and does not load it itself
func TestLoaderDoesNotLoadECharts(t *testing.T) {
	var html bytes.Buffer
	require.NoError(t, Loader().Render(context.Background(), &html))
	require.Regexp(t, `data-src="/echarts\.min\.js\?v=[0-9a-f]{16}"`, html.String())
	require.NotContains(t, html.String(), `<script src=`)
}
