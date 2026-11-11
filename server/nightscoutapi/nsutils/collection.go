package nsutils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"nighttroop.t1d.nz/models/nightscoutmodels"
)

type Collection string

const (
	Entries      Collection = "entries"
	Treatments   Collection = "treatments"
	Devicestatus Collection = "devicestatus"
	Profile      Collection = "profile"
	Properties   Collection = "properties"
	Food         Collection = "food"
	Settings     Collection = "settings"
	Activity     Collection = "activity"
)

const DATE_FORMAT_STRING = "2006-01-02T15:04:05.000Z"

func ParseCollection(s string) (Collection, error) {

	// Validate and normalize
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "device_status":
		s = "devicestatus"
	case "profiles":
		s = "profile"
	case "foods":
		s = "food"
	}

	c := Collection(s)
	switch c {
	case Entries,
		Treatments,
		Devicestatus,
		Profile,
		Properties,
		Food,
		Settings,
		Activity:
		return c, nil
	default:
		return "", fmt.Errorf("unknown collection %q", s)
	}
}

func (c Collection) DefaultDateField() string {
	switch c {
	case Entries:
		return "date"
	default:
		return "created_at"
	}
}

func (c Collection) IsTypeString(t string) bool {
	switch c {
	case Entries:
		return t == string(nightscoutmodels.SGVType) || t == string(nightscoutmodels.MBGType)
	case Treatments:
		for _, et := range nightscoutmodels.TreatmentEventTypes {
			if t == string(et) {
				return true
			}
		}
	}

	return false
}

// /
// Field Processing
// //

type CollectionAction string

const (
	CollectionActionCreate CollectionAction = "create"
	CollectionActionRead   CollectionAction = "read"
	CollectionActionUpdate CollectionAction = "update"
	CollectionActionDelete CollectionAction = "delete"
)

func GetCollectionAndEnsurePermission(r *http.Request, action CollectionAction) (Collection, *Identity, *HTTPError) {
	return GetCollectionFromStringAndEnsurePermission(chi.URLParam(r, "collection"), r, action)
}

func GetCollectionFromStringAndEnsurePermission(collectionString string, r *http.Request, action CollectionAction) (Collection, *Identity, *HTTPError) {
	collection, err := ParseCollection(collectionString)
	if err != nil {
		return "", nil, ErrInvalidCollection
	}

	identity, errr := AuthenticateExplicit(r)
	if errr != nil {
		return "", nil, ErrMissingCredentials
	}

	permission := fmt.Sprintf("api:%s:%s", collection, action)
	if !HasPermission(*identity, permission) {
		return "", nil, ErrMissingPermission(permission)
	}

	return collection, identity, nil
}

func CollectionActionHasPermission(collection Collection, identity *Identity, action CollectionAction) bool {
	if identity == nil {
		return false
	}
	permission := fmt.Sprintf("api:%s:%s", collection, action)
	return HasPermission(*identity, permission)
}
