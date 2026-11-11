package sitelogs

import (
	"context"
	"strconv"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_logs_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// EventsCollection, ErrorsCollection and LogsCollection show a site's admin its own rows of the logs DB, newest first,
// with the site taken from the site context. Events also include its batch's, as SiteEventFromDB shows them.

// rowID is a row's number, or false for anything else, which is not a row
func rowID(id string) (int64, bool) {
	n, err := strconv.ParseInt(id, 10, 64)
	return n, err == nil
}

type EventsCollection struct{}

func (s *EventsCollection) GetName() string  { return "site_events" }
func (s *EventsCollection) GetRoute() string { return "events" }

func (s *EventsCollection) ToModel(row *logsdb.Event, _ *httputils.SiteContext) *sitemodels.SiteEvent {
	return sitemodels.SiteEventFromDB(row)
}

func (s *EventsCollection) SearchRows(ctx context.Context, _ *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*logsdb.Event, error) {
	return siteContext.ControlDB.Logs.Queries.SearchSiteEvents(ctx, logsdb.SearchSiteEventsParams{SiteID: &siteContext.SiteID, Batch: &siteContext.SiteModel.Batch, Type: query.Type, Search: query.Search, Skip: query.Skip, Limit: query.Limit})
}

func (s *EventsCollection) SearchHTMXResponse(models []*sitemodels.SiteEvent, query *apiutils.Query) templ.Component {
	return site_logs_templ.EventsTable(models, query.Skip, query.Limit, query.Type, query.Search)
}

func (s *EventsCollection) GetRow(ctx context.Context, _ *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*logsdb.Event, error) {
	n, ok := rowID(id)
	if !ok {
		return nil, nil
	}
	return siteContext.ControlDB.Logs.Queries.GetSiteEvent(ctx, logsdb.GetSiteEventParams{ID: n, SiteID: &siteContext.SiteID, Batch: &siteContext.SiteModel.Batch})
}

func (s *EventsCollection) GetHTMXResponse(model *sitemodels.SiteEvent) templ.Component {
	return site_logs_templ.EventModal(model)
}

type ErrorsCollection struct{}

func (s *ErrorsCollection) GetName() string  { return "site_errors" }
func (s *ErrorsCollection) GetRoute() string { return "errors" }

func (s *ErrorsCollection) ToModel(row *logsdb.Error, _ *httputils.SiteContext) *sitemodels.SiteError {
	return sitemodels.SiteErrorFromDB(row)
}

func (s *ErrorsCollection) SearchRows(ctx context.Context, _ *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*logsdb.Error, error) {
	return siteContext.ControlDB.Logs.Queries.SearchSiteErrors(ctx, logsdb.SearchSiteErrorsParams{SiteID: &siteContext.SiteID, Type: query.Type, Search: query.Search, Skip: query.Skip, Limit: query.Limit})
}

func (s *ErrorsCollection) SearchHTMXResponse(models []*sitemodels.SiteError, query *apiutils.Query) templ.Component {
	return site_logs_templ.ErrorsTable(models, query.Skip, query.Limit, query.Type, query.Search)
}

func (s *ErrorsCollection) GetRow(ctx context.Context, _ *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*logsdb.Error, error) {
	n, ok := rowID(id)
	if !ok {
		return nil, nil
	}
	return siteContext.ControlDB.Logs.Queries.GetSiteError(ctx, logsdb.GetSiteErrorParams{ID: n, SiteID: &siteContext.SiteID})
}

func (s *ErrorsCollection) GetHTMXResponse(model *sitemodels.SiteError) templ.Component {
	return site_logs_templ.ErrorModal(model)
}

type LogsCollection struct{}

func (s *LogsCollection) GetName() string  { return "site_logs" }
func (s *LogsCollection) GetRoute() string { return "logs" }

func (s *LogsCollection) ToModel(row *logsdb.Log, _ *httputils.SiteContext) *sitemodels.SiteLog {
	return sitemodels.SiteLogFromDB(row)
}

func (s *LogsCollection) SearchRows(ctx context.Context, _ *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*logsdb.Log, error) {
	return siteContext.ControlDB.Logs.Queries.SearchSiteLogs(ctx, logsdb.SearchSiteLogsParams{SiteID: &siteContext.SiteID, Search: query.Search, Skip: query.Skip, Limit: query.Limit})
}

func (s *LogsCollection) SearchHTMXResponse(models []*sitemodels.SiteLog, query *apiutils.Query) templ.Component {
	return site_logs_templ.LogsTable(models, query.Skip, query.Limit, query.Search)
}

func (s *LogsCollection) GetRow(ctx context.Context, _ *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*logsdb.Log, error) {
	n, ok := rowID(id)
	if !ok {
		return nil, nil
	}
	return siteContext.ControlDB.Logs.Queries.GetSiteLog(ctx, logsdb.GetSiteLogParams{ID: n, SiteID: &siteContext.SiteID})
}

func (s *LogsCollection) GetHTMXResponse(model *sitemodels.SiteLog) templ.Component {
	return site_logs_templ.LogModal(model)
}
