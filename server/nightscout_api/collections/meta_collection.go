package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type MetaCreateCollection[NS, DM any] interface {
	Name() string
	Convert(*DM) nsutils.Record
	CreateDBFunc(*httputils.SiteContext, context.Context, NS, nightscout.CreateContext) (*DM, error)
}

type MetaUpdateCollection[NS, DM any] interface {
	Name() string
	Convert(*DM) nsutils.Record
	UpdateSBFunc(*httputils.SiteContext, context.Context, NS, nightscout.UpdateContext) (*DM, error)
}

type MetaGetCollection[DM any] interface {
	Name() string
	Convert(*DM) nsutils.Record
	GetDBFunc(*httputils.SiteContext, context.Context, string) (*DM, error)
}

type MetaDeleteCollection[NS, DM any] interface {
	Name() string
	DeleteDBFunc(*httputils.SiteContext, context.Context, string) error
}

type MetaDeleteTypeCollection[NS, DM any] interface {
	Name() string
	DeleteDBFunc(*httputils.SiteContext, context.Context, string) error
}

type MetaReplaceCollection[NS, DM any] interface {
	Name() string
	Convert(*DM) nsutils.Record
	ReplaceParams(NS, nightscout.UpdateContext) DM
	ReplaceDBFunc(*httputils.SiteContext, context.Context, DM) (*DM, error)
}

func MetaCreate[NS, DM any](
	m MetaCreateCollection[NS, DM],
	ctx context.Context,
	siteContext *httputils.SiteContext,
	data map[string]any,

) (*nsutils.Record, *nsutils.HTTPError) {
	var nsModel NS
	err := utils.FromMap(data, &nsModel)

	if err != nil {
		slog.Error("failed to map data", "error", err, "collection", m.Name())
		return nil, nsutils.BadDataBodyError
	}

	cc := nightscout.CreateContext{CreatedByIdentityID: siteContext.IdentityModel.ID, CreatedByIdentityName: siteContext.IdentityModel.Name}
	dbModel, err := m.CreateDBFunc(siteContext, ctx, nsModel, cc)
	if err != nil {
		slog.Error("failed to create in database", "error", err, "collection", m.Name())
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		slog.Error("failed to create in database, dbModel is nil", "collection", m.Name())
		return nil, nsutils.ErrInternal
	}
	return ptr(m.Convert(dbModel)), nil
}

func MetaGet[DM any](
	m MetaGetCollection[DM],
	ctx context.Context,
	siteContext *httputils.SiteContext,
	identifier string,
) (*nsutils.Record, *nsutils.HTTPError) {
	dbModel, err := m.GetDBFunc(siteContext, ctx, identifier)
	if err == sql.ErrNoRows {
		return nil, nsutils.ErrNotFound
	}
	if err != nil {
		slog.Error("failed to get "+m.Name()+" by ID", "error", err)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		return nil, nsutils.ErrNotFound
	}
	return ptr(m.Convert(dbModel)), nil
}
