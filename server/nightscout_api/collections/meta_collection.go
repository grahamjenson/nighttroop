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

type MetaPatchCollection[NS, DM any] interface {
	Name() string
	Convert(*DM) nsutils.Record
	PatchDBFunc(*httputils.SiteContext, context.Context, NS, string, nightscout.UpdateContext) (*DM, error)
}

type MetaGetCollection[DM any] interface {
	Name() string
	Convert(*DM) nsutils.Record
	GetDBFunc(*httputils.SiteContext, context.Context, string) (*DM, error)
}

type MetaDeleteCollection interface {
	Name() string
	HardDeleteDBFunc(*httputils.SiteContext, context.Context, string) (sql.Result, error)
	SoftDeleteDBFunc(*httputils.SiteContext, context.Context, string) (sql.Result, error)
}

type MetaDeleteTypeCollection interface {
	Name() string
	HardDeleteTypeDBFunc(*httputils.SiteContext, context.Context, string) (sql.Result, error)
	SoftDeleteTypeDBFunc(*httputils.SiteContext, context.Context, string) (sql.Result, error)
}

type MetaReplaceCollection[NS, DM any] interface {
	Name() string
	Convert(*DM) nsutils.Record

	ReplaceDBFunc(*httputils.SiteContext, context.Context, DM) (*DM, error)
}

// //
// Fns
// //
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

func MetaPatch[NS, DM any](
	m MetaPatchCollection[NS, DM],
	ctx context.Context,
	siteContext *httputils.SiteContext,
	identifier string,
	data map[string]any,
) (*nsutils.Record, *nsutils.HTTPError) {
	var nsModel NS
	err := utils.FromMap(data, &nsModel)
	if err != nil {
		slog.Error("failed to map data to model", "error", err, "collection", m.Name())
		return nil, nsutils.BadDataBodyError
	}

	uc := nightscout.UpdateContext{UpdatedByIdentityID: siteContext.IdentityModel.ID, UpdatedByIdentityName: siteContext.IdentityModel.Name}
	dbModel, err := m.PatchDBFunc(siteContext, ctx, nsModel, identifier, uc)
	if err != nil {
		slog.Error("failed to patch in database", "error", err, "collection", m.Name(), "identifier", identifier, "params", data)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		slog.Error("failed to patch in database, dbModel is nil", "collection", m.Name(), "identifier", identifier, "params", data)
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

func MetaDelete(
	m MetaDeleteCollection,
	ctx context.Context,
	siteContext *httputils.SiteContext,
	identifier string,
	permanent bool,
) *nsutils.HTTPError {

	if permanent {
		res, err := m.HardDeleteDBFunc(siteContext, ctx, identifier)
		if err != nil {
			slog.Error("failed to permanently delete by ID", "error", err, "collection", m.Name())
			return nsutils.ErrInternal
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for delete", "error", err, "collection", m.Name())
			return nsutils.ErrInternal
		}
		if rowsAffected == 0 {
			return nsutils.ErrNotFound
		}
		slog.Info("Hard Deleted", "eid", identifier, "rowsAffected", rowsAffected, "collection", m.Name())

	} else {
		res, err := m.SoftDeleteDBFunc(siteContext, ctx, identifier)
		if err != nil {
			slog.Error("failed to soft delete by ID", "error", err, "collection", m.Name())
			return nsutils.ErrInternal
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for delete", "error", err, "collection", m.Name())
			return nsutils.ErrInternal
		}
		if rowsAffected == 0 {
			return nsutils.ErrNotFound
		}
		slog.Info("Soft Deleted", "eid", identifier, "rowsAffected", rowsAffected, "collection", m.Name())

	}
	return nil
}

func MetaDeleteType(
	m MetaDeleteTypeCollection,
	ctx context.Context,
	siteContext *httputils.SiteContext,
	typeString string,
	permanent bool,
) (int, *nsutils.HTTPError) {

	if permanent {
		res, err := m.HardDeleteTypeDBFunc(siteContext, ctx, typeString)
		if err != nil {
			slog.Error("failed to permanently delete  by type", "error", err, "collection", m.Name())
			return 0, nsutils.ErrInternal
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for  delete", "error", err, "collection", m.Name())
			return 0, nsutils.ErrInternal
		}
		return int(rowsAffected), nil
	} else {
		res, err := m.SoftDeleteTypeDBFunc(siteContext, ctx, typeString)
		if err != nil {
			slog.Error("failed to soft delete  by IDs", "error", err, "collection", m.Name())
			return 0, nsutils.ErrInternal
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for  delete", "error", err, "collection", m.Name())
			return 0, nsutils.ErrInternal
		}
		return int(rowsAffected), nil
	}
}
