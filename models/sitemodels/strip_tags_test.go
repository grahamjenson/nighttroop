package sitemodels

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

// An app's notes, reasons and notifications are kept as sent, but the site gets them without HTML tags or scripts
func TestSiteModelsStripTags(t *testing.T) {
	sent := `xDrip note <b>bold</b> & "quoted" <script>alert(1)</script><img src=x onerror=alert(2)>`
	want := `xDrip note bold & "quoted" `

	dose := InsulinDoseFromDB(&sitedb.InsulinDose{Notes: &sent, Reason: &sent})
	require.Equal(t, want, *dose.Notes)
	require.Equal(t, want, *dose.Reason)

	food := FoodIntakeFromDB(&sitedb.FoodIntake{Notes: &sent})
	require.Equal(t, want, *food.Notes)

	notification := NotificationEventFromDB(&sitedb.NotificationEvent{Title: sent, Content: &sent, Notes: &sent})
	require.Equal(t, want, notification.Title)
	require.Equal(t, want, *notification.Content)
	require.Equal(t, want, *notification.Notes)

	group := EventGroupFromDB(&sitedb.EventGroupsBundle{EventGroup: sitedb.EventGroup{Notes: &sent, Reason: &sent}}, nil)
	require.Equal(t, want, *group.Notes)
	require.Equal(t, want, *group.Reason)

	// None stays none
	require.Nil(t, InsulinDoseFromDB(&sitedb.InsulinDose{}).Notes)
}
