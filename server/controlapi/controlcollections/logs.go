package controlcollections

import (
	"context"
	"strconv"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_logs_templ"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
)

// LogsCollection, EventsCollection and ErrorsCollection let control identities see the logs DB, for every site and
// the server itself, newest first

type LogsCollection struct{}

func (s *LogsCollection) GetName() string  { return "logs" }
func (s *LogsCollection) GetRoute() string { return "logs" }

func (s *LogsCollection) ToModel(row *logsdb.Log) *controlmodels.Log {
	return controlmodels.LogFromDB(row)
}

func (s *LogsCollection) SearchRows(ctx context.Context, query *apiutils.Query, controlContext *httputils.ControlContext) ([]*logsdb.Log, error) {
	return controlContext.ControlDB.Logs.Queries.SearchLogs(ctx, controlmodels.QueryToSearchLogs(query))
}

func (s *LogsCollection) SearchHTMXResponse(models []*controlmodels.Log, query *apiutils.Query) templ.Component {
	return control_logs_templ.LogsTable(models, query.Skip, query.Limit)
}

func (s *LogsCollection) GetRow(ctx context.Context, id string, controlContext *httputils.ControlContext) (*logsdb.Log, error) {
	rowID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, nil
	}
	return controlContext.ControlDB.Logs.Queries.GetLog(ctx, logsdb.GetLogParams{ID: rowID})
}

func (s *LogsCollection) GetHTMXResponse(model *controlmodels.Log) templ.Component {
	return control_logs_templ.LogModal(model)
}

type EventsCollection struct{}

func (s *EventsCollection) GetName() string  { return "events" }
func (s *EventsCollection) GetRoute() string { return "events" }

func (s *EventsCollection) ToModel(row *logsdb.Event) *controlmodels.Event {
	return controlmodels.EventFromDB(row)
}

func (s *EventsCollection) SearchRows(ctx context.Context, query *apiutils.Query, controlContext *httputils.ControlContext) ([]*logsdb.Event, error) {
	return controlContext.ControlDB.Logs.Queries.SearchEvents(ctx, controlmodels.QueryToSearchEvents(query))
}

func (s *EventsCollection) SearchHTMXResponse(models []*controlmodels.Event, query *apiutils.Query) templ.Component {
	return control_logs_templ.EventsTable(models, query.Skip, query.Limit, query.Type)
}

func (s *EventsCollection) GetRow(ctx context.Context, id string, controlContext *httputils.ControlContext) (*logsdb.Event, error) {
	rowID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, nil
	}
	return controlContext.ControlDB.Logs.Queries.GetEvent(ctx, logsdb.GetEventParams{ID: rowID})
}

func (s *EventsCollection) GetHTMXResponse(model *controlmodels.Event) templ.Component {
	return control_logs_templ.EventModal(model)
}

type ErrorsCollection struct{}

func (s *ErrorsCollection) GetName() string  { return "errors" }
func (s *ErrorsCollection) GetRoute() string { return "errors" }

func (s *ErrorsCollection) ToModel(row *logsdb.Error) *controlmodels.Error {
	return controlmodels.ErrorFromDB(row)
}

func (s *ErrorsCollection) SearchRows(ctx context.Context, query *apiutils.Query, controlContext *httputils.ControlContext) ([]*logsdb.Error, error) {
	return controlContext.ControlDB.Logs.Queries.SearchErrors(ctx, controlmodels.QueryToSearchErrors(query))
}

func (s *ErrorsCollection) SearchHTMXResponse(models []*controlmodels.Error, query *apiutils.Query) templ.Component {
	return control_logs_templ.ErrorsTable(models, query.Skip, query.Limit, query.Type)
}

func (s *ErrorsCollection) GetRow(ctx context.Context, id string, controlContext *httputils.ControlContext) (*logsdb.Error, error) {
	rowID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, nil
	}
	return controlContext.ControlDB.Logs.Queries.GetError(ctx, logsdb.GetErrorParams{ID: rowID})
}

func (s *ErrorsCollection) GetHTMXResponse(model *controlmodels.Error) templ.Component {
	return control_logs_templ.ErrorModal(model)
}
