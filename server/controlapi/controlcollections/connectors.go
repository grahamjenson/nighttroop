package controlcollections

import (
	"context"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_connectors_templ"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// ConnectorsCollection lets control identities see every site's connectors and how their syncs are going
type ConnectorsCollection struct{}

func (s *ConnectorsCollection) GetName() string {
	return "connectors"
}

// GetRoute is where the collection is in its URL, /connectors
func (s *ConnectorsCollection) GetRoute() string {
	return "connectors"
}

func (s *ConnectorsCollection) ToModel(row *controldb.SiteConnector) *controlmodels.Connector {
	return controlmodels.ConnectorFromDB(row)
}

func (s *ConnectorsCollection) SearchRows(ctx context.Context, query *apiutils.Query, controlContext *httputils.ControlContext) ([]*controldb.SiteConnector, error) {
	return controlContext.ControlDB.Queries.SearchConnectors(ctx, controlmodels.QueryToSearchConnectors(query))
}

func (s *ConnectorsCollection) SearchHTMXResponse(models []*controlmodels.Connector, query *apiutils.Query) templ.Component {
	return control_connectors_templ.ConnectorsTable(models, query.Skip, query.Limit)
}

// GetRow finds a connector by its site::type
func (s *ConnectorsCollection) GetRow(ctx context.Context, id string, controlContext *httputils.ControlContext) (*controldb.SiteConnector, error) {
	siteID, sourceType, ok := strings.Cut(id, "::")
	if !ok {
		return nil, nil
	}
	return controlContext.ControlDB.Queries.GetConnector(ctx, controldb.GetConnectorParams{SiteID: siteID, Type: consts.ConnectorType(sourceType)})
}

// SharedHostsHandler is the Nightscout hosts more than one connector syncs from, the most used first, as many sites
// syncing from one host may be Nighttroop used as a proxy
func (s *ConnectorsCollection) SharedHostsHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	rows, err := controlContext.ControlDB.Queries.SharedNightscoutHosts(r.Context())
	if herr := apiutils.ErrorFor(err, s.GetName()); herr != nil {
		herr.Write(acceptType, w)
		return
	}
	hosts := make([]*controlmodels.SharedHost, 0, len(rows))
	for _, row := range rows {
		hosts = append(hosts, controlmodels.SharedHostFromDB(row))
	}
	WriteResponse(w, r, acceptType, http.StatusOK, hosts, control_connectors_templ.SharedHosts(hosts))
}

// GetHTMXResponse is the connector's reports and recent errors, shown in a modal
func (s *ConnectorsCollection) GetHTMXResponse(model *controlmodels.Connector) templ.Component {
	return control_connectors_templ.ConnectorModal(model)
}
