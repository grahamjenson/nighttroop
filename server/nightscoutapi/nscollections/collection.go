package nscollections

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"

	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

// CollectionStore is what the Nightscout API does with a collection
type CollectionStore interface {
	Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any, raw bool) (nsutils.Document, bool, *nsutils.HTTPError)
	Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (nsutils.Document, *nsutils.HTTPError)
	Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Document, *nsutils.HTTPError)
	Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool, raw bool) (nsutils.Document, bool, *nsutils.HTTPError)
	Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, raw bool) (nsutils.Document, *nsutils.HTTPError)
	Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError
	DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError)
}

// CollectionDB is a collection kept in the site's tables, like a siteapi store: NS is its Nightscout model and DM
// its database model. Collection does each Nightscout operation the same way with it.
type CollectionDB[NS, DM any] interface {
	Name() string
	// ToNS is the database model as its Nightscout model
	ToNS(*DM, *httputils.SiteContext) NS
	// Validate checks a model being created, when old is nil, or replacing old
	Validate(nsModel NS, old *DM) *nsutils.HTTPError
	GetModel(q *sitedb.Queries, ctx context.Context, identifier string, includeDeleted bool) (*DM, error)
	// CreateModel stores the model, with the identifier given when it is replacing one
	CreateModel(q *sitedb.Queries, ctx context.Context, nsModel NS, cc nightscoutmodels.ConverterContext, identifier *string) (*DM, error)
	SearchModels(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]*DM, error)
	// Delete deletes a model, softly or, when permanent, for good
	Delete(q *sitedb.Queries, ctx context.Context, identifier string, permanent bool) (sql.Result, error)
}

// CollectionDeduplicator is a CollectionDB that looks for a model already stored that a new one repeats, which a
// create returns instead. It is only asked when the new model has a date, as with nothing to match on it would find
// the newest.
type CollectionDeduplicator[NS, DM any] interface {
	GetDuplicate(q *sitedb.Queries, ctx context.Context, nsModel NS) (*DM, error)
}

// CollectionStored is a CollectionDB that does more once a model created or replaced is stored, outside its
// transaction, such as keeping what it says in the control DB
type CollectionStored[NS any] interface {
	Stored(ctx context.Context, siteContext *httputils.SiteContext, nsModel NS)
}

// CollectionPatcher is a CollectionDB that can change a model in place; without it a patch is not implemented
type CollectionPatcher[NS, DM any] interface {
	UpdateModel(q *sitedb.Queries, ctx context.Context, nsModel NS, identifier string, old *DM, cc nightscoutmodels.ConverterContext) (*DM, error)
}

// patchValidator is a CollectionPatcher that checks a patch other than as Validate checks a replace
type patchValidator[NS, DM any] interface {
	ValidatePatch(nsModel NS, old *DM) *nsutils.HTTPError
}

// CollectionTypeDeleter is a CollectionDB that can delete everything of one type; without it that is not implemented
type CollectionTypeDeleter interface {
	DeleteType(q *sitedb.Queries, ctx context.Context, typeString string, permanent bool) (sql.Result, error)
}

// first is the first row a search found, or nil when it found none
func first[T any](rows []*T, err error) (*T, error) {
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return rows[0], nil
}

// named is a CollectionDB's name and how it makes its database models into Nightscout models, for it to embed
type named[NS, DM any] struct {
	name nsutils.Collection
	toNS func(*DM, *httputils.SiteContext) NS
}

func (n named[NS, DM]) Name() string {
	return string(n.name)
}

func (n named[NS, DM]) ToNS(dbModel *DM, siteContext *httputils.SiteContext) NS {
	return n.toNS(dbModel, siteContext)
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

// SharedDBFields are the same fields on the database model, and the server's own
type SharedDBFields interface {
	GetOccurredAt() int64
	GetCreatedAt() int64
	GetUpdatedAt() int64
	GetDeletedAt() *int64
	GetRawJsonData() *string
	GetCreatedByAppName() *string
	GetCreatedByDeviceName() *string
	GetCreatedByIdentityName() string
	GetUtcOffsetMinutes() *int64
	GetID() string
}

// Record is a database model as this collection's document, and false when the model is not this collection's
func (c Collection[NS, DM]) Record(model any, siteContext *httputils.SiteContext) (nsutils.Document, bool) {
	dbModel, ok := model.(*DM)
	if !ok || dbModel == nil {
		return nil, false
	}
	return c.document(dbModel, siteContext, false), true
}

// document is a database model as this collection's document, as the JSON it was sent as when raw is asked for and
// that JSON can be read
func (c Collection[NS, DM]) document(dbModel *DM, siteContext *httputils.SiteContext, raw bool) nsutils.Document {
	m, rawJSON := row(*dbModel)
	doc := stored[NS]{meta: m}
	if raw && rawJSON != nil && json.Unmarshal([]byte(*rawJSON), &doc.raw) == nil && doc.raw != nil {
		return doc
	}
	doc.raw, doc.model = nil, c.db.ToNS(dbModel, siteContext)
	return doc
}

func (c Collection[NS, DM]) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any, raw bool) (nsutils.Document, bool, *nsutils.HTTPError) {
	m := c.db
	var nsModel NS
	err := utils.ConvertViaJSON(data, &nsModel)
	if err != nil {
		slog.ErrorContext(ctx, "failed to map data", "error", err, "collection", m.Name())
		return nil, false, nsutils.BadDataBodyErrorWithHint(err.Error())
	}

	if err := ValidateInputData(data); err != nil {
		return nil, false, err
	}

	if err := m.Validate(nsModel, nil); err != nil {
		return nil, false, err
	}

	tx, q, httpErr := c.begin(ctx, siteContext)
	if httpErr != nil {
		return nil, false, httpErr
	}
	defer tx.Rollback()

	var duplicateDBModel *DM
	identifier := nsModel.GetIdentifier()
	if identifier != nil && *identifier == "" {
		identifier = nil
	}
	if identifier != nil {
		if duplicateDBModel, httpErr = c.identifiedDuplicate(q, ctx, *identifier); httpErr != nil {
			return nil, false, httpErr
		}
	} else if deduplicator, ok := any(m).(CollectionDeduplicator[NS, DM]); ok && nsModel.GetDate() != nil {
		duplicateDBModel, err = deduplicator.GetDuplicate(q, ctx, nsModel)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to check for duplicate in database", "error", err, "collection", m.Name())
		return nil, false, nsutils.ErrInternal
	}
	if duplicateDBModel != nil {
		return c.document(duplicateDBModel, siteContext, raw), true, nil
	}

	dbModel, err := m.CreateModel(
		q,
		ctx,
		nsModel,
		converterContext(siteContext),
		identifier,
	)

	if httpErr := c.stored(dbModel, err, "create"); httpErr != nil {
		return nil, false, httpErr
	}

	if httpErr := c.commit(tx, "create"); httpErr != nil {
		return nil, false, httpErr
	}

	siteContext.PublishModelChange(events.Created, dbModel)
	c.afterStored(ctx, siteContext, nsModel)
	return c.document(dbModel, siteContext, raw), false, nil
}

// identifiedDuplicate is the model a create naming its identifier repeats, as in Nightscout, which keeps that
// identifier so the client can delete or replace the document by it. Its only duplicate is the model stored with
// it; one deleted is removed for good, as a client like Juggluco changes a document by deleting it and creating it
// again.
func (c Collection[NS, DM]) identifiedDuplicate(q *sitedb.Queries, ctx context.Context, identifier string) (*DM, *nsutils.HTTPError) {
	if httpErr := nsutils.ValidateIdentifierIsntBad(identifier); httpErr != nil {
		return nil, httpErr
	}
	stored, err := c.db.GetModel(q, ctx, identifier, true)
	if err == sql.ErrNoRows || (err == nil && stored == nil) {
		return nil, nil
	}
	if err == nil && (*stored).GetDeletedAt() == nil {
		return stored, nil
	}
	if err == nil {
		_, err = c.db.Delete(q, ctx, identifier, true)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to check for a document with the identifier", "error", err, "collection", c.db.Name())
		return nil, nsutils.ErrInternal
	}
	return nil, nil
}

func (c Collection[NS, DM]) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (nsutils.Document, *nsutils.HTTPError) {
	m := c.db
	dbModel, err := m.GetModel(siteContext.SiteDB.Queries(), ctx, identifier, query.IncludeDeleted)
	if err == sql.ErrNoRows {
		return nil, nsutils.ErrNotFound
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to get "+m.Name()+" by ID", "error", err)
		return nil, nsutils.ErrInternal
	}
	if dbModel == nil {
		return nil, nsutils.ErrNotFound
	}
	return c.document(dbModel, siteContext, query.Raw), nil
}

func (c Collection[NS, DM]) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Document, *nsutils.HTTPError) {
	models, err := c.db.SearchModels(siteContext.SiteDB.Queries(), ctx, query)
	if err != nil {
		slog.ErrorContext(ctx, "failed to search the database", "error", err, "collection", c.db.Name())
		return nil, nsutils.ErrInternal
	}
	docs := make([]nsutils.Document, len(models))
	for i, model := range models {
		docs[i] = c.document(model, siteContext, query.Raw)
	}
	return docs, nil
}

func (c Collection[NS, DM]) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	m := c.db
	tx, q, httpErr := c.begin(ctx, siteContext)
	if httpErr != nil {
		return httpErr
	}
	defer tx.Rollback()

	// Read first, so a missing record is not found and the change can say what was deleted.
	// A permanent delete can remove a record already soft deleted.
	deleted, err := m.GetModel(q, ctx, identifier, permanent)
	if err == sql.ErrNoRows || (err == nil && deleted == nil) {
		return nsutils.ErrNotFound
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to get "+m.Name()+" to delete", "error", err)
		return nsutils.ErrInternal
	}
	label := "Soft"
	if permanent {
		label = "Hard"
	}
	res, err := m.Delete(q, ctx, identifier, permanent)
	// Treatments delete with RETURNING, so a missing ID is no rows rather than 0 rows affected
	if err == sql.ErrNoRows {
		return nsutils.ErrNotFound
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete by ID", "error", err, "collection", m.Name(), "permanent", permanent)
		return nsutils.ErrInternal
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		slog.ErrorContext(ctx, "failed to get rows affected for delete", "error", err, "collection", m.Name())
		return nsutils.ErrInternal
	}
	if rowsAffected == 0 {
		return nsutils.ErrNotFound
	}
	slog.InfoContext(ctx, label+" Deleted", "eid", identifier, "rowsAffected", rowsAffected, "collection", m.Name())

	if httpErr := c.commit(tx, "delete"); httpErr != nil {
		return httpErr
	}
	siteContext.PublishModelChange(events.Deleted, deleted)
	return nil
}

func (c Collection[NS, DM]) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	m, ok := any(c.db).(CollectionTypeDeleter)
	if !ok {
		return 0, nsutils.ErrNotImplemented
	}
	tx, q, httpErr := c.begin(ctx, siteContext)
	if httpErr != nil {
		return 0, httpErr
	}
	defer tx.Rollback()

	res, err := m.DeleteType(q, ctx, typeString, permanent)
	if err == sql.ErrNoRows {
		return 0, nsutils.ErrNotFound
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete by type", "error", err, "collection", c.db.Name(), "permanent", permanent)
		return 0, nsutils.ErrInternal
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		slog.ErrorContext(ctx, "failed to get rows affected for delete", "error", err, "collection", c.db.Name())
		return 0, nsutils.ErrInternal
	}

	if httpErr := c.commit(tx, "delete"); httpErr != nil {
		return 0, httpErr
	}

	return int(rowsAffected), nil
}

func (c Collection[NS, DM]) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, data map[string]any, canCreate bool, raw bool) (nsutils.Document, bool, *nsutils.HTTPError) {
	m := c.db
	var nsModel NS
	err := utils.ConvertViaJSON(data, &nsModel)
	if err != nil {
		slog.ErrorContext(ctx, "failed to map data to model", "error", err, "collection", m.Name())
		return nil, false, nsutils.BadDataBodyError
	}

	tx, q, httpErr := c.begin(ctx, siteContext)
	if httpErr != nil {
		return nil, false, httpErr
	}
	defer tx.Rollback()

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
		slog.ErrorContext(ctx, "failed to get by ID", "error", err, "collection", m.Name(), "identifier", identifier)
		return nil, false, nsutils.ErrInternal
	}

	// DELETE record if we are not creating a new record
	if !creatingNewRecord {
		_, err := m.Delete(q, ctx, identifier, true)
		if err != nil {
			slog.ErrorContext(ctx, "failed to hard delete in database", "error", err, "collection", m.Name(), "identifier", identifier)
			return nil, false, nsutils.ErrInternal
		}
	}

	// As in Nightscout, a deleted document is replaced whatever changed, as a client like Juggluco changes one by
	// deleting it and sending it again
	checkedModel := replacingModel
	if replacingModel != nil && (*replacingModel).GetDeletedAt() != nil {
		checkedModel = nil
	}

	if err := ValidateNotChangingSharedFields(checkedModel, nsModel, data, identifier); err != nil {
		slog.ErrorContext(ctx, "validation failed for replacing model", "error", err, "collection", m.Name())
		return nil, false, err
	}

	if err := m.Validate(nsModel, checkedModel); err != nil {
		slog.ErrorContext(ctx, "validation failed for replacing model", "error", err, "collection", m.Name())
		return nil, false, err
	}

	// CREATE
	dbCreateModel, err := m.CreateModel(q, ctx, nsModel, converterContext(siteContext), &identifier)
	if httpErr := c.stored(dbCreateModel, err, "replace"); httpErr != nil {
		return nil, false, httpErr
	}

	if httpErr := c.commit(tx, "replace"); httpErr != nil {
		return nil, false, httpErr
	}

	if creatingNewRecord {
		siteContext.PublishModelChange(events.Created, dbCreateModel)
	} else {
		siteContext.PublishModelChange(events.Updated, dbCreateModel)
	}
	c.afterStored(ctx, siteContext, nsModel)
	return c.document(dbCreateModel, siteContext, raw), creatingNewRecord, nil

}

func (c Collection[NS, DM]) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, data map[string]any, raw bool) (nsutils.Document, *nsutils.HTTPError) {
	m := c.db
	patcher, ok := any(m).(CollectionPatcher[NS, DM])
	if !ok {
		return nil, nsutils.ErrNotImplemented
	}
	var nsModel NS
	err := utils.ConvertViaJSON(data, &nsModel)
	if err != nil {
		slog.ErrorContext(ctx, "failed to map data to model", "error", err, "collection", m.Name())
		return nil, nsutils.BadDataBodyError
	}

	tx, q, httpErr := c.begin(ctx, siteContext)
	if httpErr != nil {
		return nil, httpErr
	}
	defer tx.Rollback()

	updateingModel, err := m.GetModel(q, ctx, identifier, true)

	if err == sql.ErrNoRows {
		// Check no Record exists
		return nil, nsutils.ErrNotFound
	} else if err != nil {
		slog.ErrorContext(ctx, "failed to get by ID", "error", err, "collection", m.Name(), "identifier", identifier)
		return nil, nsutils.ErrInternal
	}
	// As in Nightscout, a deleted record is gone, so it cannot be patched
	if updateingModel != nil && (*updateingModel).GetDeletedAt() != nil {
		return nil, nsutils.ErrGone
	}

	if err := ValidateNotChangingSharedFields(updateingModel, nsModel, data, identifier); err != nil {
		slog.ErrorContext(ctx, "validation failed for patching model", "error", err, "collection", m.Name())
		return nil, err
	}

	validate := m.Validate
	if v, ok := any(m).(patchValidator[NS, DM]); ok {
		validate = v.ValidatePatch
	}
	if err := validate(nsModel, updateingModel); err != nil {
		slog.ErrorContext(ctx, "validation failed for patching model", "error", err, "collection", m.Name())
		return nil, err
	}

	dbModel, err := patcher.UpdateModel(
		q,
		ctx,
		nsModel,
		identifier,
		updateingModel,
		converterContext(siteContext),
	)

	if httpErr := c.stored(dbModel, err, "patch"); httpErr != nil {
		return nil, httpErr
	}

	if httpErr := c.commit(tx, "patch"); httpErr != nil {
		return nil, httpErr
	}

	siteContext.PublishModelChange(events.Updated, dbModel)
	return c.document(dbModel, siteContext, raw), nil
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

// begin starts a write to the site's database, which is rolled back unless it is committed
func (c Collection[NS, DM]) begin(ctx context.Context, siteContext *httputils.SiteContext) (*sql.Tx, *sitedb.Queries, *nsutils.HTTPError) {
	tx, err := siteContext.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, nsutils.ErrInternal
	}
	return tx, siteContext.SiteDB.QueriesWithTx(tx), nil
}

func (c Collection[NS, DM]) commit(tx *sql.Tx, action string) *nsutils.HTTPError {
	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit "+action+" in database", "error", err, "collection", c.db.Name())
		return nsutils.ErrInternal
	}
	return nil
}

// stored is the error from storing a model with CreateModel or UpdateModel. Data the database refuses is the client's error.
func (c Collection[NS, DM]) stored(dbModel *DM, err error, action string) *nsutils.HTTPError {
	if err != nil {
		if sqlTranslatedError := nsutils.GetSQLTranslatedError(err); sqlTranslatedError != nil {
			return sqlTranslatedError
		}
		slog.Error("failed to "+action+" in database", "error", err, "collection", c.db.Name())
		return nsutils.ErrInternal
	}
	if dbModel == nil {
		slog.Error("failed to "+action+" in database, the model is nil", "collection", c.db.Name())
		return nsutils.ErrInternal
	}
	return nil
}

// converterContext is who is making a change, which a stored model records
func converterContext(siteContext *httputils.SiteContext) nightscoutmodels.ConverterContext {
	return nightscoutmodels.ConverterContext{IdentityID: siteContext.IdentityModel.ID, IdentityName: siteContext.IdentityModel.Name}
}

// afterStored has the collection do what more it does once the model is stored, if it does any
func (c Collection[NS, DM]) afterStored(ctx context.Context, siteContext *httputils.SiteContext, nsModel NS) {
	if after, ok := any(c.db).(CollectionStored[NS]); ok {
		after.Stored(ctx, siteContext, nsModel)
	}
}
