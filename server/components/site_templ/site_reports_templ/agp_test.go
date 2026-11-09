package site_reports_templ

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
)

func agpOption(t *testing.T, p *sitemodels.SiteSummaryPeriod) map[string]any {
	chart := agp(p, 4, 10)
	chart.Validate()
	var option map[string]any
	require.NoError(t, json.Unmarshal([]byte(chart.JSONNotEscaped()), &option))
	return option
}

func agpSeries(option map[string]any, name string) map[string]any {
	for _, s := range option["series"].([]any) {
		if s.(map[string]any)["name"] == name {
			return s.(map[string]any)
		}
	}
	return nil
}

// values are a series' values at each time
func values(s map[string]any) []any {
	values := []any{}
	for _, point := range s["data"].([]any) {
		value := point.(map[string]any)["value"]
		if v, ok := value.([]any); ok {
			value = v[1]
		}
		values = append(values, value)
	}
	return values
}

func TestAGP(t *testing.T) {
	mmol := consts.UnitsMMOLL
	// The first 5 minutes have percentiles, the next none
	week := &sitemodels.SiteSummaryPeriod{Period: "week", Units: &mmol, AGP: [][]float64{{3, 4, 5, 6, 18}, {}}}
	option := agpOption(t, week)

	// Each band is stacked on the percentile under it, so it is shaded between them, and is broken where there is none
	require.Equal(t, []any{3.0, "-"}, values(agpSeries(option, "5th")))
	require.Equal(t, []any{15.0, "-"}, values(agpSeries(option, "5–95%")))
	require.Equal(t, agpSeries(option, "5th")["stack"], agpSeries(option, "5–95%")["stack"])
	require.Equal(t, []any{4.0, "-"}, values(agpSeries(option, "25th")))
	require.Equal(t, []any{2.0, "-"}, values(agpSeries(option, "25–75%")))
	require.Equal(t, agpSeries(option, "25th")["stack"], agpSeries(option, "25–75%")["stack"])
	require.Equal(t, []any{5.0, "-"}, values(agpSeries(option, "Median")))
	// The second slot is 5 minutes after midnight
	require.EqualValues(t, 5*60_000, agpSeries(option, "Median")["data"].([]any)[1].(map[string]any)["value"].([]any)[0])

	// The axis goes up to the highest 95th percentile, and a week has no glucose line
	require.EqualValues(t, 18, option["yAxis"].([]any)[0].(map[string]any)["max"])
	require.Nil(t, agpSeries(option, "Glucose"))

	// A day's glucose is drawn every 5 minutes over its AGP
	glucose := 7.0
	day := &sitemodels.SiteSummaryPeriod{Period: "day", Units: &mmol, AGP: week.AGP, Glucose: []*float64{&glucose, nil}}
	require.Equal(t, []any{7.0, "-"}, values(agpSeries(agpOption(t, day), "Glucose")))
	require.True(t, hasGlucose(day))
	require.False(t, hasGlucose(&sitemodels.SiteSummaryPeriod{AGP: [][]float64{{}, {}}}))
}
