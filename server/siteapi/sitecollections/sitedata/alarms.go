package sitedata

import (
	"context"
	"time"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/alarms"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// AlarmsCollection is the signed in identity's active alarms, which the alarm checker keeps in the live DB. Deleting
// one snoozes it for the identity alone: it stays active, but stops alerting.
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

func (s *AlarmsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*livedb.Alarm, error) {
	return collectionTx.Live().GetIdentityAlarm(ctx, livedb.GetIdentityAlarmParams{SiteID: siteContext.SiteID, IdentityID: siteContext.IdentityModel.ID, Type: alarms.Type(id)})
}

// ToModel is the alarm, not alerting once it is snoozed, before the checker next checks it
func (s *AlarmsCollection) ToModel(row *livedb.Alarm, _ *httputils.SiteContext) *sitemodels.SiteAlarm {
	return &sitemodels.SiteAlarm{ID: string(row.Type), Urgent: row.Type.Level() == alarms.Urgent, Alerting: row.Alerting && time.Now().UnixMilli() >= row.SnoozedUntil}
}

func (s *AlarmsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*livedb.Alarm, error) {
	return collectionTx.Live().IdentityAlarms(ctx, livedb.IdentityAlarmsParams{SiteID: siteContext.SiteID, IdentityID: siteContext.IdentityModel.ID})
}

func (s *AlarmsCollection) SearchHTMXResponse(models []*sitemodels.SiteAlarm, query *apiutils.Query) templ.Component {
	return site_templ.AlarmBanner(models)
}

// DeleteRow snoozes the alarm for 30 minutes, as a soft delete: it stops alerting at once, and the checker stops it
// alerting to the identity's other clients on its next check
func (s *AlarmsCollection) DeleteRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) error {
	return collectionTx.Live().SnoozeAlarm(ctx, livedb.SnoozeAlarmParams{SiteID: siteContext.SiteID, IdentityID: siteContext.IdentityModel.ID,
		Type: alarms.Type(id), Until: time.Now().Add(snoozeFor).UnixMilli()})
}

// DeleteHTMXResponse is the alarm's row in the banner, snoozed
func (s *AlarmsCollection) DeleteHTMXResponse(model *sitemodels.SiteAlarm) templ.Component {
	return site_templ.AlarmRow(&sitemodels.SiteAlarm{ID: model.ID, Urgent: model.Urgent})
}
