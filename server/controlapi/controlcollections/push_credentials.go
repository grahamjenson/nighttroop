package controlcollections

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/controlmodels"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/control_templ/control_push_templ"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/i18n"
)

// PushCredentialsCollection lets control identities see every site's push credentials, and those for every site,
// never with their secrets, and admins add one to a site and delete any. One for every site cannot be added until the
// server uses it only for tokens its app registered itself, requires its bundle IDs and limits pushes across the
// server, as any site could otherwise push to any token with it.
type PushCredentialsCollection struct{}

func (s *PushCredentialsCollection) GetName() string {
	return "push_credentials"
}

// GetRoute is where the collection is in its URL, /push-credentials
func (s *PushCredentialsCollection) GetRoute() string {
	return "push-credentials"
}

func (s *PushCredentialsCollection) ToModel(row *controldb.PushCredential) *controlmodels.PushCredential {
	return &controlmodels.PushCredential{PushCredential: sitemodels.PushCredentialFromDB(row), SiteID: row.SiteID}
}

// SearchRows is every live credential, newest first
func (s *PushCredentialsCollection) SearchRows(ctx context.Context, query *apiutils.Query, controlContext *httputils.ControlContext) ([]*controldb.PushCredential, error) {
	return controlContext.ControlDB.Queries.SearchAllPushCredentials(ctx, controldb.SearchAllPushCredentialsParams{Skip: query.Skip, Limit: query.Limit})
}

func (s *PushCredentialsCollection) SearchHTMXResponse(models []*controlmodels.PushCredential, query *apiutils.Query) templ.Component {
	return control_push_templ.PushCredentials(models)
}

func (s *PushCredentialsCollection) GetRow(ctx context.Context, id string, controlContext *httputils.ControlContext) (*controldb.PushCredential, error) {
	credentialID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, apiutils.ErrNotFound
	}
	return controlContext.ControlDB.Queries.GetAnyPushCredential(ctx, controldb.GetAnyPushCredentialParams{ID: credentialID})
}

func (s *PushCredentialsCollection) GetHTMXResponse(model *controlmodels.PushCredential) templ.Component {
	return control_push_templ.PushCredentialRow(model)
}

// NewHandler is the modal to add a credential to a site in
func (s *PushCredentialsCollection) NewHandler(w http.ResponseWriter, r *http.Request) {
	WriteResponse(w, r, httputils.GetAcceptType(r), http.StatusOK, nil, control_push_templ.PushCredentialNew())
}

// CreateHandler adds a credential to a live site, as its admin would, its secret encrypted. Only admins can.
func (s *PushCredentialsCollection) CreateHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	model, herr := apiutils.ExtractModelFromRequest[controlmodels.PushCredential](httputils.GetContentType(r), r)
	if herr == nil && (model.SiteID == nil || *model.SiteID == "") {
		herr = apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.control.server_push_credential"))
	} else if herr == nil && model.PushCredential == nil {
		herr = apiutils.ErrMissingRequiredFieldWithHint("errors.missing_field.push_credential")
	}
	var created *controlmodels.PushCredential
	if herr == nil {
		var err error
		created, err = s.create(r.Context(), controlContext, model)
		herr = apiutils.ErrorFor(err, s.GetName())
	}
	if herr != nil {
		herr.Write(acceptType, w)
		return
	}
	slog.InfoContext(r.Context(), "Control admin added a push credential", "identity", controlContext.IdentityModel.Name,
		"site_id", *created.SiteID, "id", created.ID)
	WriteResponse(w, r, acceptType, http.StatusCreated, created, control_push_templ.PushCredentialRow(created))
}

// create keeps the credential for its site, if it is live, within the site's limit, in one transaction
func (s *PushCredentialsCollection) create(ctx context.Context, controlContext *httputils.ControlContext, model *controlmodels.PushCredential) (*controlmodels.PushCredential, error) {
	controlDB := controlContext.ControlDB
	tx, err := controlDB.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	q := controlDB.Queries.WithTx(tx)
	if _, err := q.GetSite(ctx, controldb.GetSiteParams{ID: *model.SiteID}); errors.Is(err, sql.ErrNoRows) {
		return nil, apiutils.ErrConstraintCheckFailedWithHint(i18n.NewError("errors.control.no_live_site"))
	} else if err != nil {
		return nil, err
	}
	credential := model.PushCredential
	params := controldb.CreatePushCredentialParams{SiteID: model.SiteID, Name: credential.Name, Provider: credential.Provider,
		TeamOrProjectID: credential.TeamOrProjectID, KeyID: credential.KeyID, AppsJson: sitemodels.StringsJSON(credential.Apps),
		BundleIdsJson: sitemodels.StringsJSONPtr(credential.BundleIDs), CreatedByIdentityName: controlContext.IdentityModel.Name}
	created, err := sitesettings.CreatePushCredential(ctx, q, controlDB.ServerConfig.Security, credential, params)
	if err != nil {
		return nil, err
	}
	return &controlmodels.PushCredential{PushCredential: created, SiteID: model.SiteID}, tx.Commit()
}

// DeleteHandler deletes any credential, removing its secret, so pushes no longer use it. Only admins can.
func (s *PushCredentialsCollection) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	controlContext, acceptType := httputils.GetControlContext(r.Context()), httputils.GetAcceptType(r)
	row, err := s.GetRow(r.Context(), chi.URLParam(r, "id"), controlContext)
	if err == nil {
		_, err = controlContext.ControlDB.Queries.DeletePushCredential(r.Context(), controldb.DeletePushCredentialParams{ID: row.ID, SiteID: row.SiteID})
	}
	if herr := apiutils.ErrorFor(err, s.GetName()); herr != nil {
		herr.Write(acceptType, w)
		return
	}
	slog.InfoContext(r.Context(), "Control admin deleted a push credential", "identity", controlContext.IdentityModel.Name, "id", row.ID)
	WriteResponse(w, r, acceptType, http.StatusOK, s.ToModel(row), control_push_templ.PushCredentialDeleted())
}
