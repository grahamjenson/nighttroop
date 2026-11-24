package site_data_templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/httputils"
)

// TestDetailsDeleteTheirRecord checks each kind's details have a Delete button that asks that kind's question, then
// deletes the record at its own address and removes its row
func TestDetailsDeleteTheirRecord(t *testing.T) {
	ctx := httputils.CreateSiteContext(context.Background(), nil, nil, nil, &controldb.Site{}, "site", "", 0, 0, "")
	ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{})
	tests := []struct {
		id, route, question string
		details             templ.Component
	}{
		{"site--glucose", "glucose", "Delete this glucose reading?", GlucoseReadingModal(&sitemodels.SiteGlucoseReading{ID: "r1"})},
		{"site--insulin", "insulin", "Delete this insulin dose?", InsulinDoseModal(&sitemodels.SiteInsulinDose{ID: "r1"})},
		{"site--food", "food", "Delete this food intake?", FoodIntakeModal(&sitemodels.SiteFoodIntake{ID: "r1"})},
		{"site--temp-basal", "temp-basals", "Delete this temp basal?", TempBasalModal(&sitemodels.SiteTempBasal{ID: "r1"})},
		{"site--temp-target", "temp-targets", "Delete this temp target?", TempTargetModal(&sitemodels.SiteTempTarget{ID: "r1"})},
		{"site--profile-switch", "profile-switches", "Delete this profile switch?", ProfileSwitchModal(&sitemodels.SiteProfileSwitch{ID: "r1"})},
		{"site--activity", "activities", "Delete this activity?", ActivityModal(&sitemodels.SiteActivity{ID: "r1"})},
		{"site--notification", "notifications", "Delete this notification?", NotificationEventModal(&sitemodels.SiteNotificationEvent{ID: "r1"})},
		{"site--event-group", "event-groups", "Delete this event group and all its events?", EventGroupModal(&sitemodels.SiteEventGroup{ID: "r1", GroupType: "Note"})},
	}
	for _, tt := range tests {
		t.Run(tt.route, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, tt.details.Render(ctx, &buf))
			html := buf.String()
			require.Contains(t, html, `<button id="`+tt.id+`--edit--delete"`)
			require.Contains(t, html, `hx-delete="/data/`+tt.route+`/r1"`)
			require.Contains(t, html, `hx-confirm="`+tt.question+`"`)
			// The row it removes is the one its table has for the record
			require.Contains(t, html, `data-row="`+tt.id+`--r1"`)
		})
	}
}
