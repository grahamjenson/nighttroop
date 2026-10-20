package sitemodels

import (
	"strconv"
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/security"
)

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:output:format function
// goverter:output:file ./z_site_converters.go
// goverter:useZeroValueOnPointerInconsistency
type SiteConverter interface {

	///////
	// SYNC SOURCE
	///////

	// goverter:map Password | SetPasswordGarbage
	// goverter:map UrlOrUsernameEnc UrlOrUsername | Decrypt
	// goverter:context security
	SyncSourceFromDB(source *controldb.SiteSyncSource, security *security.SecurityConfig) (*SyncSource, error)

	// goverter:map Password PasswordEnc
	// goverter:map UrlOrUsername UrlOrUsernameEnc
	SyncSourceToUpdate(source *SyncSource) controldb.UpdateSyncSourceParams

	// goverter:map Password PasswordEnc
	// goverter:map UrlOrUsername UrlOrUsernameEnc
	SyncSourceToCreate(source *SyncSource) controldb.CreateSyncSourceParams

	///////
	// SITE IDENTITY
	///////

	// goverter:context siteContext
	// goverter:map ID | IdentityIDToString
	// goverter:map CreatedAt Created | FormatIdentityDate
	// goverter:map LastSeenAt LastSeen | FormatOptionalIdentityDate
	// goverter:map LastWriteAt LastWrite | FormatOptionalIdentityDate
	// goverter:map Name CurrentIdentity | IsCurrentIdentity
	// goverter:map TotpSecretEnc HasTotpSecret | HasTotpSecret
	// goverter:ignore Type Password TotpSecret NSToken SiteURL
	IdentityFromDB(source *controldb.SiteIdentity, siteContext *httputils.SiteContext) *UserIdentity

	// goverter:context siteContext
	// goverter:map ID | IdentityIDFromString
	// goverter:map SiteID | ContextSiteID
	// goverter:map UpdatedByIdentityID | ContextIdentityID
	// goverter:map UpdatedByIdentityName | ContextIdentityName
	IdentityToUpdate(source *UserIdentity, siteContext *httputils.SiteContext) (controldb.UpdateSiteIdentityEnabledParams, error)

	// goverter:context siteContext
	// goverter:map ID | IdentityIDFromString
	// goverter:map SiteID | ContextSiteID
	IdentityToDelete(source *UserIdentity, siteContext *httputils.SiteContext) (controldb.DeleteSiteIdentityParams, error)
}

func SetPasswordGarbage() *string {
	return ptr(consts.PASSWORD_GARBAGE_VALUE)
}

// goverter:context security
func Decrypt(encString *string, security *security.SecurityConfig) (*string, error) {
	if encString != nil {
		decString, err := security.Decrypt(*encString)
		if err != nil {
			return nil, err
		}
		encString = &decString
	}
	if encString != nil {
		return encString, nil
	}
	return nil, nil
}

func IdentityIDToString(id int64) string {
	return strconv.FormatInt(id, 10)
}

func IdentityIDFromString(id string) (int64, error) {
	return strconv.ParseInt(id, 10, 64)
}

func FormatIdentityDate(ms int64) string {
	return time.UnixMilli(ms).Format("2006-01-02 15:04:05")
}

func FormatOptionalIdentityDate(ms *int64) string {
	if ms == nil || *ms == 0 {
		return "Never"
	}
	return FormatIdentityDate(*ms)
}

// goverter:context siteContext
func IsCurrentIdentity(name string, siteContext *httputils.SiteContext) bool {
	return name == siteContext.IdentityModel.Name
}

// goverter:context siteContext
func ContextSiteID(siteContext *httputils.SiteContext) string {
	return siteContext.Subdomain
}

// goverter:context siteContext
func ContextIdentityID(siteContext *httputils.SiteContext) *int64 {
	return &siteContext.IdentityModel.ID
}

// goverter:context siteContext
func ContextIdentityName(siteContext *httputils.SiteContext) *string {
	return &siteContext.IdentityModel.Name
}

func HasTotpSecret(totpSecretEnc string) bool {
	return totpSecretEnc != ""
}

func ptr[T any](v T) *T {
	return &v
}
