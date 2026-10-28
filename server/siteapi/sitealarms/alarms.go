// Package sitealarms is the signed in identity's alarms on the site's pages: a banner of its alerting alarms, kept
// up to date by a stream of its alarm events, and a snooze of them
package sitealarms

import (
	"net/http"
	"time"

	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/server/components/site_templ"
	"nighttroop.t1d.nz/server/siteapi/sitecollections"
	"nighttroop.t1d.nz/utils/alarms"
	"nighttroop.t1d.nz/utils/httputils"
)

// snoozeFor is how long the banner's snooze quiets the alarms, Nightscout's shortest
const snoozeFor = 30 * time.Minute

type Alarms struct {
	now *alarms.Now
}

// New keeps each identity's alarms from the bus's alarm events, for the banner to show as a page opens
func New(bus *events.SiteBus) *Alarms {
	return &Alarms{now: alarms.WatchNow(bus)}
}

// Banner is the identity's alerting alarms
func (a *Alarms) Banner(w http.ResponseWriter, r *http.Request) {
	siteContext := httputils.GetSiteContext(r.Context())
	site_templ.AlarmBanner(a.now.Of(siteContext.SiteID, siteContext.IdentityModel.ID)).Render(r.Context(), w)
}

// Events streams an alarm event each time the identity's alarms change, for its banner to be loaded again
func (a *Alarms) Events(w http.ResponseWriter, r *http.Request) {
	siteContext := httputils.GetSiteContext(r.Context())
	sitecollections.StreamEvents(w, r, siteContext, func(event events.Event) bool {
		checked, ok := event.Attributes.(alarms.Checked)
		return ok && event.Name == events.Alarm && checked.IdentityID == siteContext.IdentityModel.ID
	}, func(events.Event) (string, any) { return "alarm", struct{}{} })
}

// Snooze snoozes the identity's alerting alarms, its own and no one else's, sent to the alarm checker as an event,
// and shows its banner without them
func (a *Alarms) Snooze(w http.ResponseWriter, r *http.Request) {
	siteContext := httputils.GetSiteContext(r.Context())
	checked := a.now.Of(siteContext.SiteID, siteContext.IdentityModel.ID)
	snoozed := alarms.Snoozed{IdentityID: siteContext.IdentityModel.ID, Types: checked.Alerting, Until: time.Now().Add(snoozeFor).UnixMilli()}
	siteContext.Config.Events.Publish(siteContext.SiteID, events.Event{Name: events.AlarmSnoozed, SiteID: siteContext.SiteID, Attributes: snoozed, StartedAt: time.Now()})
	site_templ.AlarmBanner(alarms.Checked{Active: checked.Active}).Render(r.Context(), w)
}
