package sitemodels

import (
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
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

func ptr[T any](v T) *T {
	return &v
}
