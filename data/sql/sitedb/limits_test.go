package sitedb_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// Every way in writes through the site DB, so its limits refuse values far outside real use, such as a 1000 unit bolus
// or 10000 g of carbs, while still taking negative insulin and food, which take back too much
func TestSiteDataLimits(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	q, ctx, now := site.SiteDB.Queries(), context.Background(), time.Now().UnixMilli()
	tests := []struct {
		name    string
		write   func(id string, v float64) error
		ok      []float64
		refused []float64
	}{
		{"insulin dose", func(id string, v float64) error {
			_, err := q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: id, Dose: v, OccurredAt: now, CreatedByIdentityName: "test"})
			return err
		}, []float64{-10, 0.05, 500}, []float64{1000, -1000}},
		{"carbs", func(id string, v float64) error {
			_, err := q.CreateFoodIntake(ctx, sitedb.CreateFoodIntakeParams{ID: id, Carbs: &v, OccurredAt: now, CreatedByIdentityName: "test"})
			return err
		}, []float64{-50, 0, 1000}, []float64{10000, -1001}},
		{"glucose", func(id string, v float64) error {
			_, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: id, Mgdl: v, ReadingType: "sgv", OccurredAt: now, CreatedByIdentityName: "test"})
			return err
		}, []float64{39, 1000}, []float64{0, 1001}},
		{"temp basal rate", func(id string, v float64) error {
			_, err := q.CreateTempBasal(ctx, sitedb.CreateTempBasalParams{ID: id, DurationMinutes: 30, Rate: v, OccurredAt: now, CreatedByIdentityName: "test"})
			return err
		}, []float64{0, 35}, []float64{-1, 1000}},
		{"temp basal multiplier", func(id string, v float64) error {
			_, err := q.CreateTempBasal(ctx, sitedb.CreateTempBasalParams{ID: id, DurationMinutes: 30, Multiplier: &v, OccurredAt: now, CreatedByIdentityName: "test"})
			return err
		}, []float64{0, 11}, []float64{-0.1, 12}},
		{"temp target duration", func(id string, v float64) error {
			_, err := q.CreateTempTarget(ctx, sitedb.CreateTempTargetParams{ID: id, DurationMinutes: int64(v), HighTargetMgdl: 140, LowTargetMgdl: 120, OccurredAt: now, CreatedByIdentityName: "test"})
			return err
		}, []float64{0, 10080}, []float64{-1, 20000}},
		{"heart rate", func(id string, v float64) error {
			hr := int64(v)
			_, err := q.CreateActivity(ctx, sitedb.CreateActivityParams{ID: id, ActivityType: "exercise", HeartRate: &hr, OccurredAt: now, CreatedByIdentityName: "test"})
			return err
		}, []float64{60}, []float64{1000, -1}},
		{"loop IOB", func(id string, v float64) error {
			_, err := q.CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: id, LoopIob: &v, OccurredAt: now, CreatedByIdentityName: "test"})
			return err
		}, []float64{-3, 20}, []float64{5000}},
		{"profile switch percent", func(id string, v float64) error {
			_, err := q.CreateProfileSwitch(ctx, sitedb.CreateProfileSwitchParams{ID: id, Name: id, Percent: &v, OccurredAt: now, CreatedByIdentityName: "test"})
			return err
		}, []float64{100}, []float64{5000, -1}},
		{"occurred at", func(id string, v float64) error {
			_, err := q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: id, Dose: 1, OccurredAt: int64(v), CreatedByIdentityName: "test"})
			return err
		}, []float64{float64(now)}, []float64{-1, 1e15}},
		{"utc offset", func(id string, v float64) error {
			offset := int64(v)
			_, err := q.CreateInsulinDose(ctx, sitedb.CreateInsulinDoseParams{ID: id, Dose: 1, UtcOffsetMinutes: &offset, OccurredAt: now, CreatedByIdentityName: "test"})
			return err
		}, []float64{720, -600}, []float64{100000}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, v := range tt.ok {
				require.NoError(t, tt.write(fmt.Sprintf("%s ok %v", tt.name, v), v), v)
			}
			for _, v := range tt.refused {
				require.ErrorContains(t, tt.write(fmt.Sprintf("%s refused %v", tt.name, v), v), "CHECK constraint failed", v)
			}
		})
	}
}
