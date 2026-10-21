package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/events/sitemodels"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

// CollectionStore is what the Nightscout API does with a collection
type CollectionStore interface {
	Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError)
	Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (*nsutils.Record, *nsutils.HTTPError)
	Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError)
	Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError)
	Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, raw bool) (*nsutils.Record, *nsutils.HTTPError)
	Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError
	DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError)
}

// CollectionDB is a collection kept in the site's tables, like a siteapi store: NS is its Nightscout model and DM
// its database model. Collection does each Nightscout operation the same way with it.
type CollectionDB[NS, DM any] interface {
	Name() string
	Convert(*DM, *httputils.SiteContext, bool) nsutils.Record
	// Validate checks a model being created, when old is nil, or replacing old
	Validate(nsModel NS, old *DM) *nsutils.HTTPError
	GetModel(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*DM, error)
	// GetDuplicate is a model already stored that the new one repeats, or nil
	GetDuplicate(q *sitedb.Queries, ctx context.Context, nsModel NS) (*DM, error)
	// CreateModel stores the model, with the identifier given when it is replacing one
	CreateModel(q *sitedb.Queries, ctx context.Context, nsModel NS, cc nightscoutmodels.ConverterContext, identifier *string) (*DM, error)
	SearchModels(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]*DM, error)
	SoftDelete(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error)
	HardDelete(q *sitedb.Queries, ctx context.Context, identifier string) (sql.Result, error)
}

// CollectionPatcher is a CollectionDB that can change a model in place; without it a patch is not implemented
type CollectionPatcher[NS, DM any] interface {
	ValidatePatch(nsModel NS, old *DM) *nsutils.HTTPError
	UpdateModel(q *sitedb.Queries, ctx context.Context, nsModel NS, identifier string, old *DM, cc nightscoutmodels.ConverterContext) (*DM, error)
}

// CollectionTypeDeleter is a CollectionDB that can delete everything of one type; without it that is not implemented
type CollectionTypeDeleter interface {
	SoftDeleteType(q *sitedb.Queries, ctx context.Context, typeString string) (sql.Result, error)
	HardDeleteType(q *sitedb.Queries, ctx context.Context, typeString string) (sql.Result, error)
}

// All is each Nightscout collection's store
func All() map[nsutils.Collection]CollectionStore {
	return map[nsutils.Collection]CollectionStore{
		nsutils.Entries:      NewEntriesCollection(),
		nsutils.Treatments:   NewTreatmentCollection(),
		nsutils.Devicestatus: NewDeviceStatusCollection(),
		nsutils.Profile:      NewProfileCollection(),
		nsutils.Food:         NewFoodCollection(),
		nsutils.Settings:     NewSettingsCollection(),
		nsutils.Activity:     NewActivityCollection(),
	}
}

// Collection is the CollectionStore for a CollectionDB
type Collection[NS SharedNSFields, DM SharedDBFields] struct {
	db CollectionDB[NS, DM]
}

func NewCollection[NS SharedNSFields, DM SharedDBFields](db CollectionDB[NS, DM]) Collection[NS, DM] {
	return Collection[NS, DM]{db: db}
}

// SharedNSFields are the fields every Nightscout model has, which a replace or patch must not change
type SharedNSFields interface {
	GetDevice() *string
	GetApp() *string
	GetSubject() *string
	GetDate() *int64
	GetUtcOffsetMinutes() *int64
	GetIdentifier() *string
}

// SharedDBFields are the same fields on the database model
type SharedDBFields interface {
	GetOccurredAt() int64
	GetCreatedByAppName() *string
	GetCreatedByDeviceName() *string
	GetCreatedByIdentityName() string
	GetUtcOffsetMinutes() *int64
	GetID() string
}

// Record is a database model as this collection's record, and false when the model is not this collection's
func (c Collection[NS, DM]) Record(model any, siteContext *httputils.SiteContext) (nsutils.Record, bool) {
	dbModel, ok := model.(*DM)
	if !ok || dbModel == nil {
		return nsutils.Record{}, false
	}
	return c.db.Convert(dbModel, siteContext, false), true
}

func (c Collection[NS, DM]) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	m := c.db
	var nsModel NS
	err := utils.FromMap(data, &nsModel)
	if err != nil {
		slog.Error("failed to map data", "error", err, "collection", m.Name())
		return nil, false, nsutils.BadDataBodyErrorWithHint(err.Error())
	}

	if err := ValidateInputData(data); err != nil {
		return nil, false, err
	}

	if err := m.Validate(nsModel, nil); err != nil {
		return nil, false, err
	}

	tx, err := siteContext.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, nsutils.ErrInternal
	}
	defer tx.Rollback()

	q := siteContext.SiteDB.Queries.WithTx(tx)

	duplicateDBModel, err := m.GetDuplicate(q, ctx, nsModel)
	if err != nil {
		slog.Error("failed to check for duplicate in database", "error", err, "collection", m.Name())
		return nil, false, nsutils.ErrInternal
	}
	if duplicateDBModel != nil {
		return ptr(m.Convert(duplicateDBModel, siteContext, raw)), true, nil
	}

	dbModel, err := m.CreateModel(
		q,
		ctx,
		nsModel,
		nightscoutmodels.ConverterContext{IdentityID: siteContext.IdentityModel.ID, IdentityName: siteContext.IdentityModel.Name},
		nil,
	)

	if err != nil {
		sqlTranslatedError := nsutils.GetSQLTranslatedError(err)
		if sqlTranslatedError != nil {
			return nil, false, sqlTranslatedError
		}

		slog.Error("failed to create in database", "error", err, "collection", m.Name())
		return nil, false, nsutils.ErrInternal
	}
	if dbModel == nil {
		slog.Error("failed to create in database, dbModel is nil", "collection", m.Name())
		return nil, false, nsutils.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit create in database", "error", err, "dbModel", dbModel, "collection", m.Name())
		return nil, false, nsutils.ErrInternal
	}

	siteContext.PublishModelChange(sitemodels.Created, dbModel)
	return ptr(m.Convert(dbModel, siteContext, raw)), false, nil
}

func (c Collection[NS, DM]) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (*nsutils.Record, *nsutils.HTTPError) {
	m := c.db
	dbModel, err := m.GetModel(siteContext.SiteDB.Queries, ctx, identifier, query.IncludeDeleted)
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
	return ptr(m.Convert(dbModel, siteContext, query.Raw)), nil
}

func (c Collection[NS, DM]) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	models, err := c.db.SearchModels(siteContext.SiteDB.Queries, ctx, query)
	if err != nil {
		slog.Error("failed to search the database", "error", err, "collection", c.db.Name())
		return nil, nsutils.ErrInternal
	}
	return nsutils.CreateRecordSlice(models, c.db.Convert, siteContext, query.Raw), nil
}

func (c Collection[NS, DM]) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	m := c.db
	tx, err := siteContext.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return nsutils.ErrInternal
	}
	defer tx.Rollback()
	q := siteContext.SiteDB.Queries.WithTx(tx)

	// Read first, so the change can say what was deleted
	deleted, _ := m.GetModel(q, ctx, identifier, false)
	deleteFunc, label := m.SoftDelete, "Soft"
	if permanent {
		deleteFunc, label = m.HardDelete, "Hard"
	}
	res, err := deleteFunc(q, ctx, identifier)
	// Treatments delete with RETURNING, so a missing ID is no rows rather than 0 rows affected
	if err == sql.ErrNoRows {
		return nsutils.ErrNotFound
	}
	if err != nil {
		slog.Error("failed to delete by ID", "error", err, "collection", m.Name(), "permanent", permanent)
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
	slog.Info(label+" Deleted", "eid", identifier, "rowsAffected", rowsAffected, "collection", m.Name())

	if err := tx.Commit(); err != nil {
		slog.Error("failed to delete in database", "error", err, "collection", m.Name())
		return nsutils.ErrInternal
	}
	if deleted != nil {
		siteContext.PublishModelChange(sitemodels.Deleted, deleted)
	}
	return nil
}

func (c Collection[NS, DM]) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	m, ok := any(c.db).(CollectionTypeDeleter)
	if !ok {
		return 0, nsutils.ErrNotImplemented
	}
	tx, err := siteContext.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return 0, nsutils.ErrInternal
	}
	defer tx.Rollback()
	q := siteContext.SiteDB.Queries.WithTx(tx)

	deleteFunc := m.SoftDeleteType
	if permanent {
		deleteFunc = m.HardDeleteType
	}
	res, err := deleteFunc(q, ctx, typeString)
	if err == sql.ErrNoRows {
		return 0, nsutils.ErrNotFound
	}
	if err != nil {
		slog.Error("failed to delete by type", "error", err, "collection", c.db.Name(), "permanent", permanent)
		return 0, nsutils.ErrInternal
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		slog.Error("failed to get rows affected for delete", "error", err, "collection", c.db.Name())
		return 0, nsutils.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to delete in database", "error", err, "collection", c.db.Name())
		return 0, nsutils.ErrInternal
	}

	return int(rowsAffected), nil
}

func (c Collection[NS, DM]) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, data map[string]any, canCreate bool, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	m := c.db
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

	cc := nightscoutmodels.ConverterContext{IdentityID: siteContext.IdentityModel.ID, IdentityName: siteContext.IdentityModel.Name}

	creatingNewRecord := false

	// GET
	replacingModel, err := m.GetModel(q, ctx, identifier, true)
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
		_, err := m.HardDelete(q, ctx, identifier)
		if err != nil {
			slog.Error("failed to hard delete in database", "error", err, "collection", m.Name(), "identifier", identifier)
			return nil, false, nsutils.ErrInternal
		}
	}

	if err := ValidateNotChangingSharedFields(replacingModel, nsModel, data, identifier); err != nil {
		slog.Error("validation failed for replacing model", "error", err, "collection", m.Name())
		return nil, false, err
	}

	if err := m.Validate(nsModel, replacingModel); err != nil {
		slog.Error("validation failed for replacing model", "error", err, "collection", m.Name())
		return nil, false, err
	}

	// CREATE
	dbCreateModel, err := m.CreateModel(q, ctx, nsModel, cc, &identifier)
	if err != nil {
		// Bad data is the client's error, as it is for create
		if sqlTranslatedError := nsutils.GetSQLTranslatedError(err); sqlTranslatedError != nil {
			return nil, false, sqlTranslatedError
		}
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

	if creatingNewRecord {
		siteContext.PublishModelChange(sitemodels.Created, dbCreateModel)
	} else {
		siteContext.PublishModelChange(sitemodels.Updated, dbCreateModel)
	}
	return ptr(m.Convert(dbCreateModel, siteContext, raw)), creatingNewRecord, nil

}

func (c Collection[NS, DM]) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, data map[string]any, raw bool) (*nsutils.Record, *nsutils.HTTPError) {
	m := c.db
	patcher, ok := any(m).(CollectionPatcher[NS, DM])
	if !ok {
		return nil, nsutils.ErrNotImplemented
	}
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

	updateingModel, err := m.GetModel(q, ctx, identifier, true)

	if err == sql.ErrNoRows {
		// Check no Record exists
		return nil, nsutils.ErrNotFound
	} else if err != nil {
		slog.Error("failed to get by ID", "error", err, "collection", m.Name(), "identifier", identifier)
		return nil, nsutils.ErrInternal
	}

	if err := ValidateNotChangingSharedFields(updateingModel, nsModel, data, identifier); err != nil {
		slog.Error("validation failed for patching model", "error", err, "collection", m.Name())
		return nil, err
	}

	if err := patcher.ValidatePatch(nsModel, updateingModel); err != nil {
		slog.Error("validation failed for patching model", "error", err, "collection", m.Name())
		return nil, err
	}

	dbModel, err := patcher.UpdateModel(
		q,
		ctx,
		nsModel,
		identifier,
		updateingModel,
		nightscoutmodels.ConverterContext{IdentityID: siteContext.IdentityModel.ID, IdentityName: siteContext.IdentityModel.Name},
	)

	if err != nil {
		// Bad data is the client's error, as it is for create and replace
		if sqlTranslatedError := nsutils.GetSQLTranslatedError(err); sqlTranslatedError != nil {
			return nil, sqlTranslatedError
		}
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

	siteContext.PublishModelChange(sitemodels.Updated, dbModel)
	return ptr(m.Convert(dbModel, siteContext, raw)), nil
}

func ValidateInputData(nsModelMap map[string]any) *nsutils.HTTPError {
	// Check to see if user has submitted read-only fields
	if _, ok := nsModelMap["srvCreated"]; ok {
		return nsutils.ErrCannotChangeCreatedAt
	}
	if _, ok := nsModelMap["srvModified"]; ok {
		return nsutils.ErrCannotChangeModifiedAt
	}
	if _, ok := nsModelMap["modifiedBy"]; ok {
		return nsutils.ErrCannotChangeModifiedBy
	}

	return nil
}

func ValidateNotChangingSharedFields[NS SharedNSFields, DM SharedDBFields](replacingModelPtr *DM, nsModel NS, nsModelMap map[string]any, identifier string) *nsutils.HTTPError {
	if err := ValidateInputData(nsModelMap); err != nil {
		return err
	}

	if replacingModelPtr == nil {
		return nil
	}
	replacingModel := *replacingModelPtr

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

	if nsModel.GetUtcOffsetMinutes() != nil && replacingModel.GetUtcOffsetMinutes() != nil && *nsModel.GetUtcOffsetMinutes() != *replacingModel.GetUtcOffsetMinutes() {
		return nsutils.ErrCannotChangeUTCOffset
	}

	if nsModel.GetIdentifier() != nil && *nsModel.GetIdentifier() != identifier {
		return nsutils.ErrCannotChangeIdentifier
	}

	if replacingModel.GetID() != identifier {
		return nsutils.ErrCannotChangeIdentifier
	}

	return nil
}
