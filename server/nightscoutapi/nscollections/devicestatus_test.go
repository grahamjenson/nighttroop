package nscollections

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/nightscoutmodels"
)

// loopColumns are the device report's columns and other attributes a loop's status sets: its name, the status as sent,
// its IOB and COB
type loopColumns struct {
	Name, Status                     *string
	IOB, IOBBasal, IOBTreatment, COB *float64
}

// A device status keeps its OpenAPS or Loop status as sent, predictions and all, and its IOB and COB, Loop's over
// OpenAPS's when it has both
func TestNightscoutToDBCreateDeviceReportLoopColumns(t *testing.T) {
	openAPS := `{"iob":{"iob":2.5,"basaliob":1},"suggested":{"COB":20,"timestamp":"2024-01-01T00:00:00Z","predBGs":{"IOB":[100,110],"ZT":[90]}}}`
	loop := `{"iob":{"iob":3},"cob":{"cob":15},"predicted":{"startDate":"2024-01-01T00:05:00Z","values":[120,130]}}`
	tests := []struct {
		name   string
		status string
		want   loopColumns
	}{
		{"none", `{}`, loopColumns{}},
		{"openaps", `{"openaps":` + openAPS + `}`, loopColumns{
			Name: ptr("openaps"), Status: ptr(openAPS),
			IOB: ptr(2.5), IOBBasal: ptr(1.0), IOBTreatment: ptr(1.5), COB: ptr(20.0),
		}},
		{"openaps without iob or suggested", `{"openaps":{}}`, loopColumns{Name: ptr("openaps"), Status: ptr("{}")}},
		{"openaps without basal iob or predictions", `{"openaps":{"iob":{"iob":2},"suggested":{"predBGs":{}}}}`, loopColumns{
			Name: ptr("openaps"), Status: ptr(`{"iob":{"iob":2},"suggested":{"predBGs":{}}}`), IOB: ptr(2.0),
		}},
		{"loop", `{"loop":` + loop + `}`, loopColumns{
			Name: ptr("loop"), Status: ptr(loop),
			IOB: ptr(3.0), COB: ptr(15.0),
		}},
		{"loop without predictions", `{"loop":{"predicted":{}}}`, loopColumns{Name: ptr("loop"), Status: ptr(`{"predicted":{}}`)}},
		{"both", `{"openaps":` + openAPS + `,"loop":{"iob":{"iob":3}}}`, loopColumns{
			Name: ptr("loop"), Status: ptr(`{"iob":{"iob":3}}`),
			IOB: ptr(3.0), IOBBasal: ptr(1.0), IOBTreatment: ptr(1.5), COB: ptr(20.0),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status nightscoutmodels.DeviceStatus
			require.NoError(t, json.Unmarshal([]byte(tt.status), &status))
			got := NightscoutToDBCreateDeviceReport(status, nightscoutmodels.ConverterContext{})
			other := got.OtherAttributes
			require.Equal(t, tt.want, loopColumns{
				Name: other.LoopName, Status: got.LoopStatusJson, IOB: got.LoopIob, IOBBasal: got.LoopIobBasal,
				IOBTreatment: got.LoopIobTreatment, COB: got.LoopCob,
			})
		})
	}
}
