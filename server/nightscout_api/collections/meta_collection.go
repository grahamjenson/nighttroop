package collections

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type MetaCreateCollection[NS, DM any] interface {
	Name() string
	Convert(*DM) nsutils.Record
	CreateDBFunc(*site.Queries, context.Context, NS, nightscout.CreateContext, *DM, *string) (*DM, error)
}

type MetaPatchCollection[NS, DM any] interface {
	Name() string
	Convert(*DM) nsutils.Record
	PatchDBFunc(*site.Queries, context.Context, NS, string, nightscout.UpdateContext) (*DM, error)
}

type MetaGetCollection[DM any] interface {
	Name() string
	Convert(*DM) nsutils.Record
	GetDBFunc(*site.Queries, context.Context, string) (*DM, error)
}

type MetaDeleteCollection interface {
	Name() string
	HardDeleteDBFunc(*site.Queries, context.Context, string) (sql.Result, error)
	SoftDeleteDBFunc(*site.Queries, context.Context, string) (sql.Result, error)
}

type MetaDeleteTypeCollection interface {
	Name() string
	HardDeleteTypeDBFunc(*site.Queries, context.Context, string) (sql.Result, error)
	SoftDeleteTypeDBFunc(*site.Queries, context.Context, string) (sql.Result, error)
}

type MetaReplaceCollection[NS, DM any] interface {
	Name() string
	Convert(*DM) nsutils.Record
	GetDBFunc(*site.Queries, context.Context, string) (*DM, error)
	HardDeleteDBFunc(*site.Queries, context.Context, string) (sql.Result, error)
	CreateDBFunc(*site.Queries, context.Context, NS, nightscout.CreateContext, *DM, *string) (*DM, error)
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

	tx, err := siteContext.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, nsutils.ErrInternal
	}
	defer tx.Rollback()

	q := siteContext.SiteDB.Queries.WithTx(tx)

	dbModel, err := m.CreateDBFunc(
		q,
		ctx,
		nsModel,
		nightscout.CreateContext{CreatedByIdentityID: siteContext.IdentityModel.ID, CreatedByIdentityName: siteContext.IdentityModel.Name},
		nil,
		nil,
	)

	if err != nil {
		slog.Error("failed to create in database", "error", err, "collection", m.Name())
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		slog.Error("failed to create in database, dbModel is nil", "collection", m.Name())
		return nil, nsutils.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit create in database", "error", err, "dbModel", dbModel, "collection", m.Name())
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

	tx, err := siteContext.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, nsutils.ErrInternal
	}
	defer tx.Rollback()

	q := siteContext.SiteDB.Queries.WithTx(tx)

	dbModel, err := m.PatchDBFunc(
		q,
		ctx,
		nsModel,
		identifier,
		nightscout.UpdateContext{UpdatedByIdentityID: siteContext.IdentityModel.ID, UpdatedByIdentityName: siteContext.IdentityModel.Name},
	)

	if err != nil {
		slog.Error("failed to patch in database", "error", err, "collection", m.Name(), "identifier", identifier, "params", data)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		slog.Error("failed to patch in database, dbModel is nil", "collection", m.Name(), "identifier", identifier, "params", data)
		return nil, nsutils.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit patch in database", "error", err, "dbModel", dbModel, "collection", m.Name())
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
	dbModel, err := m.GetDBFunc(siteContext.SiteDB.Queries, ctx, identifier)
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
	tx, err := siteContext.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return nsutils.ErrInternal
	}
	defer tx.Rollback()
	q := siteContext.SiteDB.Queries.WithTx(tx)

	if permanent {
		res, err := m.HardDeleteDBFunc(q, ctx, identifier)
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
		res, err := m.SoftDeleteDBFunc(q, ctx, identifier)
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

	if err := tx.Commit(); err != nil {
		slog.Error("failed to delete in database", "error", err, "collection", m.Name())
		return nsutils.ErrInternal
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
	rowsAffected := int64(0)
	tx, err := siteContext.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return 0, nsutils.ErrInternal
	}
	defer tx.Rollback()
	q := siteContext.SiteDB.Queries.WithTx(tx)

	if permanent {
		res, err := m.HardDeleteTypeDBFunc(q, ctx, typeString)
		if err != nil {
			slog.Error("failed to permanently delete  by type", "error", err, "collection", m.Name())
			return 0, nsutils.ErrInternal
		}
		rowsAffected, err = res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for  delete", "error", err, "collection", m.Name())
			return 0, nsutils.ErrInternal
		}

	} else {
		res, err := m.SoftDeleteTypeDBFunc(q, ctx, typeString)
		if err != nil {
			slog.Error("failed to soft delete  by IDs", "error", err, "collection", m.Name())
			return 0, nsutils.ErrInternal
		}
		rowsAffected, err = res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for  delete", "error", err, "collection", m.Name())
			return 0, nsutils.ErrInternal
		}

	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to delete in database", "error", err, "collection", m.Name())
		return 0, nsutils.ErrInternal
	}

	return int(rowsAffected), nil
}

func MetaReplace[NS interface {
	GetDevice() *string
	GetApp() *string
	GetSubject() *string
	GetDate() *int64
}, DM interface {
	GetOccurredAt() int64
	GetCreatedByAppName() *string
	GetCreatedByDeviceName() *string
	GetCreatedByIdentityName() string
}](
	m MetaReplaceCollection[NS, DM],
	ctx context.Context,
	siteContext *httputils.SiteContext,
	identifier string,
	data map[string]any,
	canCreate bool,
) (*nsutils.Record, bool, *nsutils.HTTPError) {
	var nsModel NS
	err := utils.FromMap(data, &nsModel)
	if err != nil {
		slog.Error("failed to map data to model", "error", err, "collection", m.Name())
		return nil, false, nsutils.BadDataBodyError
	}

	tx, err := siteContext.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, nsutils.ErrInternal
	}
	defer tx.Rollback()
	q := siteContext.SiteDB.Queries.WithTx(tx)

	cc := nightscout.CreateContext{CreatedByIdentityID: siteContext.IdentityModel.ID, CreatedByIdentityName: siteContext.IdentityModel.Name}

	creatingNewRecord := false

	// GET
	replacingModel, err := m.GetDBFunc(q, ctx, identifier)
	if err == sql.ErrNoRows {
		// Check no Record exists
		if !canCreate {
			return nil, false, nsutils.ErrNotFound
		}
		creatingNewRecord = true
		replacingModel = nil
	} else if err != nil {
		slog.Error("failed to get by ID", "error", err, "collection", m.Name(), "identifier", identifier)
		return nil, false, nsutils.ErrInternal
	}

	// DELETE record if we are not creating a new record
	if !creatingNewRecord {
		_, err := m.HardDeleteDBFunc(q, ctx, identifier)
		if err != nil {
			slog.Error("failed to hard delete in database", "error", err, "collection", m.Name(), "identifier", identifier)
			return nil, false, nsutils.ErrInternal
		}
	}

	if err := ValidateNotChangingSharedFields(replacingModel, nsModel); err != nil {
		slog.Error("validation failed for replacing model", "error", err, "collection", m.Name())
		return nil, false, err
	}

	// CREATE
	dbCreateModel, err := m.CreateDBFunc(q, ctx, nsModel, cc, replacingModel, &identifier)
	if err != nil {
		slog.Error("failed to create in database", "error", err, "collection", m.Name())
		return nil, false, nsutils.ErrInternal
	}

	if dbCreateModel == nil {
		slog.Error("failed to create in database, dbCreateModel is nil", "collection", m.Name())
		return nil, false, nsutils.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit create in database", "error", err, "collection", m.Name())
		return nil, false, nsutils.ErrInternal
	}

	return ptr(m.Convert(dbCreateModel)), creatingNewRecord, nil

}

func ValidateNotChangingSharedFields[NS interface {
	GetDevice() *string
	GetApp() *string
	GetSubject() *string
	GetDate() *int64
}, DM interface {
	GetOccurredAt() int64
	GetCreatedByAppName() *string
	GetCreatedByDeviceName() *string
	GetCreatedByIdentityName() string
}](replacingModelPtr *DM, nsModel NS) *nsutils.HTTPError {
	if replacingModelPtr == nil {
		return nil
	}
	replacingModel := *replacingModelPtr

	fmt.Printf("Validating not changing shared fields for replacing model: %+v and new model: %+v\n", *replacingModelPtr, nsModel)
	fmt.Printf("Diff in nsModel Model Device: %v, App: %v, Subject: %v, Date: %v\n Replacing Device: %v, App: %v, Subject: %v, Date: %v\n", nsModel.GetDevice(), nsModel.GetApp(), nsModel.GetSubject(), nsModel.GetDate(), replacingModel.GetCreatedByDeviceName(), replacingModel.GetCreatedByAppName(), replacingModel.GetCreatedByIdentityName(), replacingModel.GetOccurredAt())

	// Validate that the new model has the same device, app, subject, and date as the replacing model
	if nsModel.GetDevice() != nil && replacingModel.GetCreatedByDeviceName() != nil && *nsModel.GetDevice() != *replacingModel.GetCreatedByDeviceName() {
		return nsutils.ErrCannotChangeDevice
	}
	if nsModel.GetApp() != nil && replacingModel.GetCreatedByAppName() != nil && *nsModel.GetApp() != *replacingModel.GetCreatedByAppName() {
		return nsutils.ErrCannotChangeApp
	}
	if nsModel.GetSubject() != nil && replacingModel.GetCreatedByIdentityName() != "" && *nsModel.GetSubject() != replacingModel.GetCreatedByIdentityName() {
		return nsutils.ErrCannotChangeSubject
	}
	if nsModel.GetDate() != nil && replacingModel.GetOccurredAt() != 0 && *nsModel.GetDate() != replacingModel.GetOccurredAt() {
		return nsutils.ErrCannotChangeDate
	}

	return nil
}
