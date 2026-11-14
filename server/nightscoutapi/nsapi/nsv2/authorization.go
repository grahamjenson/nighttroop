package nsv2

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// role is a Nightscout v2 authorization role
type role struct {
	ID          string   `json:"_id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	Notes       string   `json:"notes"`
}

// subject is a Nightscout v2 authorization subject, an identity of the site
type subject struct {
	ID          string   `json:"_id"`
	Name        string   `json:"name"`
	Roles       []string `json:"roles"`
	Notes       string   `json:"notes"`
	AccessToken *string  `json:"accessToken,omitempty"`
}

// subjectNameLimit is what a site identity's name takes, as its table checks
var subjectNameLimit, _ = controldb.ColumnLimitOf("site_identities", "Name")

// aapsClient is the role AAPS gives a client it pairs, a device of the site that reads and writes
const aapsClient = "aaps-client"

// roles are the roles a subject can have, fixed, as each is one of the site's identity roles; a role sent is not kept
var roles = []role{
	{ID: "admin", Name: "admin", Permissions: []string{"*"}},
	{ID: "readable", Name: "readable", Permissions: []string{"*:*:read"}},
	{ID: aapsClient, Name: aapsClient, Permissions: []string{"*:*:read", "api:settings:admin", "api:settings:create", "api:settings:update", "notifications:*:ack"}},
}

// newTokens are the Nightscout tokens of the subjects made here in the last minutes, by site and identity, as AAPS
// reads a new subject's accessToken from the subjects after it makes it; a token is kept only as a digest after that
var newTokens = expirable.NewLRU[string, string](1000, nil, 10*time.Minute)

// pwdPhoneOnly lets only the site's PWD phone, the AAPS that pairs its clients, add, list and delete subjects
func pwdPhoneOnly(next func(http.ResponseWriter, *http.Request, *httputils.SiteContext)) http.HandlerFunc {
	return nsutils.Require("api:*:read", false, func(w http.ResponseWriter, r *http.Request, _ *nsutils.Identity) {
		siteContext := httputils.GetSiteContext(r.Context())
		if siteContext.IdentityModel.Role != consts.IdentityRoleDevicePWDPhone {
			nsutils.WriteError(w, nsutils.ErrMissingPermission("api:subjects:admin"))
			return
		}
		next(w, r, siteContext)
	})
}

// registerAuthorizationRoutes are Nightscout's roles and subjects, which AAPS pairs a client with by QR code
func registerAuthorizationRoutes(r *chi.Mux) {
	r.Get("/authorization/roles", nsutils.Require("api:*:read", false, func(w http.ResponseWriter, _ *http.Request, _ *nsutils.Identity) {
		apiutils.WriteJSON(w, http.StatusOK, roles)
	}))
	r.Post("/authorization/roles", pwdPhoneOnly(echoRole(http.StatusCreated)))
	r.Put("/authorization/roles", pwdPhoneOnly(echoRole(http.StatusOK)))
	r.Get("/authorization/subjects", pwdPhoneOnly(listSubjects))
	r.Post("/authorization/subjects", pwdPhoneOnly(createSubject))
	r.Delete("/authorization/subjects/{id}", pwdPhoneOnly(deleteSubject))
}

// echoRole gives back the fixed role of the name sent, or the role as sent, without keeping it
func echoRole(status int) func(http.ResponseWriter, *http.Request, *httputils.SiteContext) {
	return func(w http.ResponseWriter, r *http.Request, _ *httputils.SiteContext) {
		var sent role
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&sent); err != nil || sent.Name == "" {
			nsutils.WriteError(w, nsutils.ErrBadRequest)
			return
		}
		for _, known := range roles {
			if known.Name == sent.Name {
				sent = known
			}
		}
		sent.ID = sent.Name
		apiutils.WriteJSON(w, status, sent)
	}
}

func listSubjects(w http.ResponseWriter, r *http.Request, siteContext *httputils.SiteContext) {
	tx := siteapiutils.NewCollectionTx(r.Context(), siteContext)
	defer tx.Rollback()
	identities, err := (&sitesettings.IdentitiesCollection{}).SearchRows(r.Context(), tx, nil, siteContext)
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	subjects := make([]subject, 0, len(identities))
	for _, identity := range identities {
		subjects = append(subjects, toSubject(siteContext, identity))
	}
	apiutils.WriteJSON(w, http.StatusOK, subjects)
}

// createSubject adds a device that reads and writes, for an AAPS client, as the site's identities are added
func createSubject(w http.ResponseWriter, r *http.Request, siteContext *httputils.SiteContext) {
	var sent subject
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&sent); err != nil || sent.Name == "" {
		nsutils.WriteError(w, nsutils.ErrBadRequest)
		return
	}
	if len(sent.Name) > subjectNameLimit.MaxLength {
		nsutils.WriteError(w, nsutils.BadDataBodyErrorWithHint("name must be at most "+strconv.Itoa(subjectNameLimit.MaxLength)+" characters"))
		return
	}
	tx := siteapiutils.NewCollectionTx(r.Context(), siteContext)
	defer tx.Rollback()
	created, err := (&sitesettings.IdentitiesCollection{}).CreateRow(r.Context(), tx, nil, siteContext,
		&sitemodels.UserIdentity{Name: sent.Name, Role: string(consts.IdentityRoleDeviceRead), Client: sitemodels.AAPSClient})
	if err == nil {
		err = finish(tx)
	}
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	if created.NSToken != nil {
		newTokens.Add(tokenKey(siteContext, created.ID), *created.NSToken)
	}
	apiutils.WriteJSON(w, http.StatusCreated, toSubject(siteContext, created))
}

// deleteSubject deletes the site's identity, as the site does, so its token stops working
func deleteSubject(w http.ResponseWriter, r *http.Request, siteContext *httputils.SiteContext) {
	id := chi.URLParam(r, "id")
	tx := siteapiutils.NewCollectionTx(r.Context(), siteContext)
	defer tx.Rollback()
	err := (&sitesettings.IdentitiesCollection{}).DeleteRow(r.Context(), tx, id, siteContext)
	if err == nil {
		err = finish(tx)
	}
	if err != nil {
		writeIdentityError(w, err)
		return
	}
	newTokens.Remove(siteContext.SiteID + ":" + id) // as tokenKey writes it
	apiutils.WriteJSON(w, http.StatusOK, map[string]any{"status": http.StatusOK})
}

func toSubject(siteContext *httputils.SiteContext, identity *sitemodels.UserIdentity) subject {
	s := subject{ID: strconv.FormatInt(identity.ID, 10), Name: identity.Name, Roles: []string{"readable"}}
	ns, err := nsutils.NighttroopIdentityToNightscoutIdentity(identity.Name, consts.IdentityRole(identity.Role), &identity.Client)
	switch {
	case identity.Client == sitemodels.AAPSClient:
		s.Roles = []string{aapsClient}
	case consts.IdentityRole(identity.Role) == consts.IdentityRoleAdmin, err == nil && nsutils.HasPermission(*ns, "api:treatments:create"):
		s.Roles = []string{"admin"}
	}
	if token, ok := newTokens.Get(tokenKey(siteContext, identity.ID)); ok {
		s.AccessToken = &token
	}
	return s
}

func tokenKey(siteContext *httputils.SiteContext, identityID int64) string {
	return siteContext.SiteID + ":" + strconv.FormatInt(identityID, 10)
}

// finish commits the identity's change
func finish(tx *siteapiutils.CollectionTx) error {
	if herr := tx.Finish(nil); herr != nil {
		return herr
	}
	return nil
}

// writeIdentityError writes the site's error for an identity, as Nightscout's error
func writeIdentityError(w http.ResponseWriter, err error) {
	var herr *apiutils.HTTPError
	switch {
	case errors.As(err, &herr):
		nsutils.WriteError(w, &nsutils.HTTPError{Status: herr.Status, Message: herr.Message})
	case errors.Is(err, sql.ErrNoRows):
		nsutils.WriteError(w, nsutils.ErrNotFound)
	default:
		nsutils.WriteError(w, nsutils.ErrBadRequest)
	}
}
