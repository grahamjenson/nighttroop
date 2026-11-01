package sitecollections

import (
	"net/http"
	"time"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitedata"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/apiutils"
)

// eventChoicesWindow is how far either side of a group's time the events to pick from are
const eventChoicesWindow = time.Hour

// eventChoicesParams are the time to pick events near, typed in the form, and the group being edited, if it is
type eventChoicesParams struct {
	OccurredAt int64  `json:"occurred_at"`
	GroupID    string `json:"group_id"`
}

// EventGroupChoicesHandler is the events that can be grouped, within an hour of occurred_at (or now), as a select of
// each kind for the create and edit forms. With group_id it is the edit form's, with the group's own events picked.
func EventGroupChoicesHandler(w http.ResponseWriter, r *http.Request) {
	siteContext := SiteContextFor(w, r, (&sitedata.EventGroupsCollection{}).GetName(), siteapiutils.ActionCreate, siteapiutils.ActionUpdate)
	if siteContext == nil {
		return
	}
	var params eventChoicesParams
	if err := utils.ConvertViaJSON(apiutils.FormToMap[eventChoicesParams](r.URL.Query(), siteContext.GetSettings().Location()), &params); err != nil {
		writeError(w, r, siteContext.AcceptType, apiutils.ErrBadRequest)
		return
	}
	at := time.Now()
	if params.OccurredAt != 0 {
		at = time.UnixMilli(params.OccurredAt)
	}

	q := siteContext.SiteDB.Queries
	picked := map[string]*string{}
	if params.GroupID != "" {
		group, err := q.GetEventGroupWithoutJoins(r.Context(), sitedb.GetEventGroupWithoutJoinsParams{ID: params.GroupID})
		if herr := answer(err, "event_groups"); herr != nil {
			writeError(w, r, siteContext.AcceptType, herr)
			return
		}
		picked = map[string]*string{"glucose_reading_id": group.GlucoseReadingID, "insulin_dose_id": group.InsulinDoseID,
			"food_intake_id": group.FoodIntakeID, "temp_basal_id": group.TempBasalID, "temp_target_id": group.TempTargetID,
			"profile_switch_id": group.ProfileSwitchID, "component_event_id": group.ComponentEventID,
			"activity_id": group.ActivityID, "notification_event_id": group.NotificationEventID}
	}
	from, to := at.Add(-eventChoicesWindow).UnixMilli(), at.Add(eventChoicesWindow).UnixMilli()
	rows, err := q.EventsToGroup(r.Context(), sitedb.EventsToGroupParams{GroupID: params.GroupID, From: from, To: to})
	if herr := answer(err, "event_groups"); herr != nil {
		writeError(w, r, siteContext.AcceptType, herr)
		return
	}

	choices := make([]*sitemodels.SiteEventChoice, 0, len(rows))
	for _, row := range rows {
		choice := &sitemodels.SiteEventChoice{Kind: row.Kind, ID: row.ID, OccurredAt: row.OccurredAt, Amount: row.Amount,
			Minutes: row.Minutes, Detail: row.Detail}
		choice.Picked = picked[row.Kind] != nil && *picked[row.Kind] == row.ID
		if (row.Kind == "glucose_reading_id" || row.Kind == "temp_target_id") && row.Amount != nil {
			choice.Amount = new(sitemodels.MgdlToSettingsUnits(*row.Amount, siteContext))
		}
		choices = append(choices, choice)
	}
	WriteResponse(w, r, siteContext, http.StatusOK, choices, site_data_templ.EventGroupChoices(params.GroupID != "", from, to, choices), false)
}
