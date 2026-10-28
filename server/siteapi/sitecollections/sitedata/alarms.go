package sitedata

import (
	"context"
	"slices"
	"time"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/alarms"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// AlarmsCollection is the signed in identity's active alarms. It is ephemeral: no DB keeps them, they are held as the
// alarm checker last sent them, and deleting one snoozes it for the identity alone.
type AlarmsCollection struct{}

// snoozeFor is how long deleting an alarm snoozes it, Nightscout's shortest
const snoozeFor = 30 * time.Minute

func (s *AlarmsCollection) GetName() string {
	return "alarms"
}

// GetRoute is where the collection is in its URL
func (s *AlarmsCollection) GetRoute() string {
	return "alarms"
}

// identityAlarms are the identity's active alarms, the most urgent first
func identityAlarms(siteContext *httputils.SiteContext) []*sitemodels.SiteAlarm {
	checked := alarms.NowFor(siteContext.Config.Events).Of(siteContext.SiteID, siteContext.IdentityModel.ID)
	rows := make([]*sitemodels.SiteAlarm, len(checked.Active))
	for i, t := range checked.Active {
		rows[i] = &sitemodels.SiteAlarm{ID: string(t), Urgent: t.Level() == alarms.Urgent, Alerting: slices.Contains(checked.Alerting, t)}
	}
	return rows
}

func (s *AlarmsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteAlarm, error) {
	for _, alarm := range identityAlarms(siteContext) {
		if alarm.ID == id {
			return alarm, nil
		}
	}
	return nil, nil
}

func (s *AlarmsCollection) ToModel(row *sitemodels.SiteAlarm, _ *httputils.SiteContext) *sitemodels.SiteAlarm {
	return row
}

func (s *AlarmsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteAlarm, error) {
	return identityAlarms(siteContext), nil
}

func (s *AlarmsCollection) SearchHTMXResponse(models []*sitemodels.SiteAlarm, query *apiutils.Query) templ.Component {
	return site_templ.AlarmBanner(models)
}

// DeleteRow snoozes the alarm for 30 minutes: it stops alerting at once, and the snooze is sent to the alarm checker
// as an event
func (s *AlarmsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	identityID, alarm := siteContext.IdentityModel.ID, alarms.Type(id)
	alarms.NowFor(siteContext.Config.Events).Quiet(siteContext.SiteID, identityID, alarm)
	snoozed := alarms.Snoozed{IdentityID: identityID, Types: []alarms.Type{alarm}, Until: time.Now().Add(snoozeFor).UnixMilli()}
	siteContext.Config.Events.Publish(siteContext.SiteID, events.Event{Name: events.AlarmSnoozed, SiteID: siteContext.SiteID, Attributes: snoozed, StartedAt: time.Now()})
	return nil
}

// DeleteHTMXResponse is the alarm's row in the banner, snoozed
func (s *AlarmsCollection) DeleteHTMXResponse(model *sitemodels.SiteAlarm) templ.Component {
	return site_templ.AlarmRow(&sitemodels.SiteAlarm{ID: model.ID, Urgent: model.Urgent})
}

// Watches is the identity's own alarm events, which are what change its alarms
func (s *AlarmsCollection) Watches(event events.Event, siteContext *httputils.SiteContext) bool {
	checked, ok := event.Attributes.(alarms.Checked)
	return ok && event.Name == events.Alarm && checked.IdentityID == siteContext.IdentityModel.ID
}
