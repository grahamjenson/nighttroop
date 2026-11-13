package nscollections

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"slices"
	"time"

	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/apiutils/apiquery"
	"nighttroop.t1d.nz/utils/httputils"
)

// CollectionStore is what the Nightscout API does with a collection
type CollectionStore interface {
	Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (nsutils.Document, bool, *nsutils.HTTPError)
	Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (nsutils.Document, *nsutils.HTTPError)
	Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Document, *nsutils.HTTPError)
	// History is what changed since, deleted too, the oldest change first, as API v3 history
	History(ctx context.Context, siteContext *httputils.SiteContext, since int64, query nsquery.Query) ([]nsutils.Document, *nsutils.HTTPError)
	Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool) (nsutils.Document, bool, *nsutils.HTTPError)
	Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (nsutils.Document, *nsutils.HTTPError)
	Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError
	DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError)
	DeleteQuery(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) (int, *nsutils.HTTPError)
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
	// AscendingIDs are the identifiers of the models in the query's dates, the oldest first, paged as it asks
	AscendingIDs(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]string, error)
	// HistoryIDs are the identifiers of at most limit models changed since, deleted too, the oldest change first
	HistoryIDs(q *sitedb.Queries, ctx context.Context, since, limit int64) ([]string, error)
	// Delete deletes a model, softly or, when permanent, for good
	Delete(q *sitedb.Queries, ctx context.Context, identifier string, permanent bool) (sql.Result, error)
}

// CollectionKeyed is a CollectionDB whose model sent again is found by its fallback key when no model has its
// identifier, as Nightscout finds a duplicate (treatment-fallback-key.js, API v3's dedup). Without one, only the
// identifier finds a duplicate. A create returns the duplicate instead.
type CollectionKeyed[NS, DM any] interface {
	FallbackKey() FallbackKey[NS, DM]
}

// FallbackKey is the fields a model sent again has the same as the one stored. An app's own id, like xDrip+'s uuid or
// Loop's syncIdentifier, is the identifier of a model sent without an _id, so the identifier finds it first.
type FallbackKey[NS, DM any] struct {
	// search finds at most limit models stored at the sent model's date, with an index of the key's fields
	search func(q *sitedb.Queries, ctx context.Context, nsModel NS, limit int64) ([]*DM, error)
	// fields must each be the same, each compared in its own way
	fields []KeyField[NS, DM]
	// clientIDs are the apps' ids a stored model can have: one stored with an app's id is found only by its identifier
	clientIDs []func(stored *DM) *string
	// whateverItsIdentifier finds a duplicate of a model sent with an identifier no model has too, as Nightscout
	// upserts an entry on its date and type whatever its _id
	whateverItsIdentifier bool
}

// KeyField is a field of a fallback key, and whether the model sent and the one stored have the same value of it
type KeyField[NS, DM any] struct {
	name string
	same func(sent NS, stored *DM) bool
}

// maxSameTimeDuplicates is how many models a fallback key's search looks through for the one sent again, so a lookup
// stays cheap however many are stored at one date
const maxSameTimeDuplicates = 10

// matches is whether the model sent repeats the one stored, by the key
func (k FallbackKey[NS, DM]) matches(sent NS, stored *DM) bool {
	for _, clientID := range k.clientIDs {
		if clientID(stored) != nil {
			return false
		}
	}
	for _, field := range k.fields {
		if !field.same(sent, stored) {
			return false
		}
	}
	return true
}

// sameDate is a fallback key's date, the same to the millisecond
func sameDate[NS SharedNSFields, DM SharedDBFields]() KeyField[NS, DM] {
	return KeyField[NS, DM]{"date", func(sent NS, stored *DM) bool {
		return sent.GetDate() != nil && *sent.GetDate() == (*stored).GetOccurredAt()
	}}
}

// sameDevice is a fallback key's device. A model sent without one has the same as any, as a search without one finds
// any device's.
func sameDevice[NS SharedNSFields, DM SharedDBFields]() KeyField[NS, DM] {
	return KeyField[NS, DM]{"device", func(sent NS, stored *DM) bool {
		return kept(sent.GetDevice(), (*stored).GetCreatedByDeviceName())
	}}
}

// kept is whether a value sent leaves the stored one as it is, as one not sent does
func kept[T comparable](sent *T, stored *T) bool {
	return sent == nil || (stored != nil && *sent == *stored)
}

// CollectionStored is a CollectionDB that does more once a model created, replaced or refreshed is stored, outside its
// transaction, such as keeping what it says in the control DB
type CollectionStored[NS any] interface {
	Stored(ctx context.Context, siteContext *httputils.SiteContext, nsModel NS)
}

// CollectionRefresher is a CollectionDB whose duplicate is updated with the model repeating it, as Nightscout
// upserts one sent again, when Refreshes says it is: the fields sent are merged into it as a patch's are. A duplicate
// the merge leaves as it is is not changed, so a client sending a batch again changes nothing.
type CollectionRefresher[NS, DM any] interface {
	Refreshes(nsModel NS, duplicate *DM) bool
}

// CollectionPatcher is a CollectionDB that can change a model in place; without it a patch is not implemented.
// RewriteModel writes the model, converted as a create converts it, over the stored one, every column, keeping its
// identifier, when and by whom it was created, and what of it only the database has, which may depend on the fields
// sent.
type CollectionPatcher[NS, DM any] interface {
	RewriteModel(q *sitedb.Queries, ctx context.Context, nsModel NS, sent map[string]any, stored *DM, cc nightscoutmodels.ConverterContext) (*DM, error)
}

// CollectionTypeDeleter is a CollectionDB that can delete everything of one type; without it that is not implemented
type CollectionTypeDeleter interface {
	DeleteType(q *sitedb.Queries, ctx context.Context, typeString string, permanent bool) (sql.Result, error)
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
	return c.document(dbModel, siteContext), true
}

// document is a database model as this collection's document
func (c Collection[NS, DM]) document(dbModel *DM, siteContext *httputils.SiteContext) nsutils.Document {
	return stored[NS]{meta: row(*dbModel), model: c.db.ToNS(dbModel, siteContext), unitsLabels: unitsLabels(siteContext)}
}

func (c Collection[NS, DM]) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any) (nsutils.Document, bool, *nsutils.HTTPError) {
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
	if nsModel, err = withUtcOffset(data, nsModel, siteUtcOffset(siteContext)); err != nil {
		return nil, false, nsutils.BadDataBodyErrorWithHint(err.Error())
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
	}
	if duplicateDBModel == nil {
		duplicateDBModel, err = c.keyDuplicate(q, ctx, nsModel, identifier != nil)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to check for duplicate in database", "error", err, "collection", m.Name())
		return nil, false, nsutils.ErrInternal
	}
	// One naming its identifier repeats its duplicate only with the same date, as Loop sends a dose again when it ends
	if duplicateDBModel != nil && (identifier == nil || (nsModel.GetDate() != nil && *nsModel.GetDate() == (*duplicateDBModel).GetOccurredAt())) {
		return c.repeated(ctx, siteContext, tx, q, nsModel, data, duplicateDBModel)
	}
	if duplicateDBModel != nil {
		return c.document(duplicateDBModel, siteContext), true, nil
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
	// A treatment or reading sent already deleted, isValid false, is kept soft deleted, as Nightscout keeps isValid as
	// sent: AAPS uploads what it deleted while offline so, and a follower would otherwise take it as current. Presets
	// are not soft deleted, so a food or profile keeps being stored as before.
	valid, sent := data["isValid"].(bool)
	deleted := sent && !valid && (m.Name() == string(nsutils.Treatments) || m.Name() == string(nsutils.Entries))
	if deleted {
		if dbModel, httpErr = c.softDeleted(q, ctx, dbModel); httpErr != nil {
			return nil, false, httpErr
		}
	}

	if httpErr := c.commit(tx, "create"); httpErr != nil {
		return nil, false, httpErr
	}

	if deleted {
		siteContext.PublishModelChange(events.Deleted, dbModel)
		return c.document(dbModel, siteContext), false, nil
	}
	siteContext.PublishModelChange(events.Created, dbModel)
	c.afterStored(ctx, siteContext, nsModel)
	return c.document(dbModel, siteContext), false, nil
}

// softDeleted deletes the model just stored, as a soft delete does, and is it deleted
func (c Collection[NS, DM]) softDeleted(q *sitedb.Queries, ctx context.Context, dbModel *DM) (*DM, *nsutils.HTTPError) {
	id := (*dbModel).GetID()
	if _, err := c.db.Delete(q, ctx, id, false); err != nil {
		slog.ErrorContext(ctx, "failed to delete a model created deleted", "error", err, "collection", c.db.Name())
		return nil, nsutils.ErrInternal
	}
	deleted, err := c.db.GetModel(q, ctx, id, true)
	if err == nil && deleted != nil {
		err = q.MarkStale(ctx, deleted)
	}
	if err != nil || deleted == nil {
		slog.ErrorContext(ctx, "failed to read a model created deleted", "error", err, "collection", c.db.Name())
		return nil, nsutils.ErrInternal
	}
	return deleted, nil
}

// repeated is the duplicate a create repeats, with the fields sent merged into it when the collection is a
// CollectionRefresher and it refreshes the duplicate. It is changed only when its document is, so the write is rolled
// back otherwise, before the duplicate's document is read, as that may read the site's database.
func (c Collection[NS, DM]) repeated(ctx context.Context, siteContext *httputils.SiteContext, tx *sql.Tx, q *sitedb.Queries, nsModel NS, data map[string]any, duplicate *DM) (nsutils.Document, bool, *nsutils.HTTPError) {
	refresher, ok := any(c.db).(CollectionRefresher[NS, DM])
	if !ok || !refresher.Refreshes(nsModel, duplicate) {
		_ = tx.Rollback()
		return c.document(duplicate, siteContext), true, nil
	}
	refreshed, merged, httpErr := c.rewrite(ctx, siteContext, q, duplicate, data)
	if httpErr != nil {
		return nil, false, httpErr
	}
	if !c.changed(duplicate, refreshed, siteContext) {
		_ = tx.Rollback()
		return c.document(duplicate, siteContext), true, nil
	}
	if httpErr := c.commit(tx, "refresh"); httpErr != nil {
		return nil, false, httpErr
	}
	siteContext.PublishModelChange(events.Updated, refreshed)
	c.afterStored(ctx, siteContext, merged)
	return c.document(refreshed, siteContext), true, nil
}

// rewrite writes the stored model's document with the fields sent merged into it, see merged, converted and checked as
// a create's model is, over the stored model. It is the model written and the merged one.
func (c Collection[NS, DM]) rewrite(ctx context.Context, siteContext *httputils.SiteContext, q *sitedb.Queries, stored *DM, data map[string]any) (*DM, NS, *nsutils.HTTPError) {
	var merged NS
	rewriter, ok := any(c.db).(CollectionPatcher[NS, DM])
	if !ok {
		return nil, merged, nsutils.ErrNotImplemented
	}
	merged, err := c.merged(stored, data, siteContext)
	if err != nil {
		slog.ErrorContext(ctx, "failed to merge the fields sent into the stored model", "error", err, "collection", c.db.Name())
		return nil, merged, nsutils.BadDataBodyErrorWithHint(err.Error())
	}
	if err := c.db.Validate(merged, stored); err != nil {
		slog.ErrorContext(ctx, "validation failed for the merged model", "error", err, "collection", c.db.Name())
		return nil, merged, err
	}
	rewritten, err := rewriter.RewriteModel(q, ctx, merged, data, stored, converterContext(siteContext))
	if httpErr := c.stored(rewritten, err, "update"); httpErr != nil {
		return nil, merged, httpErr
	}
	return rewritten, merged, nil
}

// merged is the stored model's document, as API v1 gives it, with the fields sent laid over it, see mergeFields, read
// as a create reads the fields sent
func (c Collection[NS, DM]) merged(stored *DM, data map[string]any, siteContext *httputils.SiteContext) (NS, error) {
	var model NS
	var aliases map[string][]string
	var glucose []string
	if fields, ok := any(model).(interface{ MergeAliases() map[string][]string }); ok {
		aliases = fields.MergeAliases()
	}
	if fields, ok := any(model).(interface{ GlucoseFields() []string }); ok {
		glucose = fields.GlucoseFields()
	}
	document := c.document(stored, withoutSiteDB(siteContext)).Fields(false)
	err := utils.ConvertViaJSON(mergeFields(document, data, aliases, glucose), &model)
	return model, err
}

// changed is whether the rewritten model's document is not the stored one's, the server's fields left out
func (c Collection[NS, DM]) changed(stored, rewritten *DM, siteContext *httputils.SiteContext) bool {
	before, _ := json.Marshal(c.db.ToNS(stored, withoutSiteDB(siteContext)))
	after, _ := json.Marshal(c.db.ToNS(rewritten, withoutSiteDB(siteContext)))
	return string(before) != string(after)
}

// withoutSiteDB is the site's context for a document read in a write's transaction, which holds the site's database:
// what the document reads from another table, a device status's configuration, is left out, and the collection keeps
// it unless it is sent
func withoutSiteDB(siteContext *httputils.SiteContext) *httputils.SiteContext {
	without := *siteContext
	without.SiteDB = nil
	return &without
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

// keyDuplicate is the model a model sent with a date repeats, found by the collection's fallback key. A model sent
// with an identifier, which no model has, has one only for a key that finds one whatever its identifier.
func (c Collection[NS, DM]) keyDuplicate(q *sitedb.Queries, ctx context.Context, nsModel NS, identified bool) (*DM, error) {
	keyed, ok := any(c.db).(CollectionKeyed[NS, DM])
	if !ok || nsModel.GetDate() == nil {
		return nil, nil
	}
	key := keyed.FallbackKey()
	if identified && !key.whateverItsIdentifier {
		return nil, nil
	}
	found, err := key.search(q, ctx, nsModel, maxSameTimeDuplicates)
	for _, stored := range found {
		if key.matches(nsModel, stored) {
			return stored, nil
		}
	}
	return nil, err
}

// Get is the model with the identifier. As in Nightscout, a deleted one is gone, unless deleted ones are asked for.
func (c Collection[NS, DM]) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (nsutils.Document, *nsutils.HTTPError) {
	m := c.db
	dbModel, err := m.GetModel(siteContext.SiteDB.Queries(), ctx, identifier, true)
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
	if (*dbModel).GetDeletedAt() != nil && !query.IncludeDeleted {
		return nil, nsutils.ErrGone
	}
	return c.document(dbModel, siteContext), nil
}

func (c Collection[NS, DM]) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Document, *nsutils.HTTPError) {
	search := c.db.SearchModels
	if query.SortOrder != nil && *query.SortOrder == nsquery.SortOccurredAtAsc {
		search = c.ascending
	}
	models, err := search(siteContext.SiteDB.Queries(), ctx, query)
	if err != nil {
		slog.ErrorContext(ctx, "failed to search the database", "error", err, "collection", c.db.Name())
		return nil, nsutils.ErrInternal
	}
	return c.documents(models, siteContext), nil
}

// History finds what changed in updated_at's index, then reads each by its primary key, so a page starts from since
// however old what changed is. A query with filters of its own searches for what changed with them instead, which
// reads through the dates searched.
func (c Collection[NS, DM]) History(ctx context.Context, siteContext *httputils.SiteContext, since int64, query nsquery.Query) ([]nsutils.Document, *nsutils.HTTPError) {
	q := siteContext.SiteDB.Queries()
	ids, err := c.db.HistoryIDs(q, ctx, since, query.Limit)
	models := make([]*DM, 0, len(ids))
	if err == nil && query.HasFilters && len(ids) > 0 {
		idIn, _ := json.Marshal(ids)
		query.IDIn, query.Skip, query.IncludeDeleted = new(string(idIn)), 0, true
		models, err = c.db.SearchModels(q, ctx, query)
	}
	if err == nil && !query.HasFilters {
		models, err = c.byKey(q, ctx, ids, true)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to read the history", "error", err, "collection", c.db.Name())
		return nil, nsutils.ErrInternal
	}
	// Nightscout sorts history by srvModified, then date; then the identifier, as its HistoryIDs query does, so a page
	// a limit cuts off among changes made in the same millisecond is the first of them in this order
	slices.SortStableFunc(models, func(a, b *DM) int {
		return cmp.Or(cmp.Compare((*a).GetUpdatedAt(), (*b).GetUpdatedAt()), cmp.Compare((*a).GetOccurredAt(), (*b).GetOccurredAt()),
			cmp.Compare((*a).GetID(), (*b).GetID()))
	})
	return c.documents(models, siteContext), nil
}

// documents are the database models as this collection's documents
func (c Collection[NS, DM]) documents(models []*DM, siteContext *httputils.SiteContext) []nsutils.Document {
	docs := make([]nsutils.Document, len(models))
	for i, model := range models {
		docs[i] = c.document(model, siteContext)
	}
	return docs
}

// ascending is what a search sorted by date finds, the oldest first, as AAPS pages forward by date. The identifiers in
// the dates searched are read from the date's index, the oldest first, then each by its primary key. A query with
// other filters searches for those identifiers with them instead, so a page has fewer than its limit when some of
// them do not match.
func (c Collection[NS, DM]) ascending(q *sitedb.Queries, ctx context.Context, query nsquery.Query) ([]*DM, error) {
	ids, err := c.db.AscendingIDs(q, ctx, query)
	if err != nil || len(ids) == 0 {
		return nil, err
	}
	if !query.DatesOnly() {
		idIn, _ := json.Marshal(ids)
		query.IDIn, query.Skip, query.HasFilters = new(string(idIn)), 0, true
		models, err := c.db.SearchModels(q, ctx, query)
		slices.SortStableFunc(models, func(a, b *DM) int { return cmp.Compare((*a).GetOccurredAt(), (*b).GetOccurredAt()) })
		return models, err
	}
	return c.byKey(q, ctx, ids, query.IncludeDeleted)
}

// byKey are the models with identifiers ids, in their order, each read by its primary key. One deleted for good since
// its identifier was read is left out.
func (c Collection[NS, DM]) byKey(q *sitedb.Queries, ctx context.Context, ids []string, includeDeleted bool) ([]*DM, error) {
	models := make([]*DM, 0, len(ids))
	for _, id := range ids {
		model, err := c.db.GetModel(q, ctx, id, includeDeleted)
		if err == sql.ErrNoRows || (err == nil && model == nil) {
			continue
		}
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
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
		// One deleted already is gone rather than not found, which API v3 answers as done, as Nightscout does, since
		// AAPS deletes a treatment again after a follower's delete reaches it
		if !permanent {
			if gone, goneErr := m.GetModel(q, ctx, identifier, true); goneErr == nil && gone != nil && (*gone).GetDeletedAt() != nil {
				return nsutils.ErrGone
			}
		}
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
	if err := q.MarkStale(ctx, deleted); err != nil {
		slog.ErrorContext(ctx, "failed to mark summaries stale", "error", err, "collection", m.Name())
		return nsutils.ErrInternal
	}

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

// DeleteQuery soft deletes what a search with the query finds, at most a search's worth, the newest first, and says how
// many: the rest are deleted by asking again. Each is deleted on its own, so it is published as deleted.
func (c Collection[NS, DM]) DeleteQuery(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) (int, *nsutils.HTTPError) {
	tx, q, httpErr := c.begin(ctx, siteContext)
	if httpErr != nil {
		return 0, httpErr
	}
	defer tx.Rollback()

	query.IncludeDeleted, query.Skip, query.Limit = false, 0, apiquery.MaxQueryLimit
	found, err := c.db.SearchModels(q, ctx, query)
	for i := 0; err == nil && i < len(found); i++ {
		_, err = c.db.Delete(q, ctx, (*found[i]).GetID(), false)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete by query", "error", err, "collection", c.db.Name())
		return 0, nsutils.ErrInternal
	}

	if httpErr := c.commit(tx, "delete"); httpErr != nil {
		return 0, httpErr
	}
	for _, deleted := range found {
		siteContext.PublishModelChange(events.Deleted, deleted)
	}
	return len(found), nil
}

func (c Collection[NS, DM]) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, data map[string]any, canCreate bool) (nsutils.Document, bool, *nsutils.HTTPError) {
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

	replaced, httpErr := c.removeReplaced(ctx, q, identifier, canCreate)
	if httpErr != nil {
		return nil, false, httpErr
	}
	if httpErr := c.validateReplace(ctx, nsModel, replaced, data, identifier); httpErr != nil {
		return nil, false, httpErr
	}
	offset := siteUtcOffset(siteContext)
	if replaced != nil && (*replaced).GetUtcOffsetMinutes() != nil {
		kept := *(*replaced).GetUtcOffsetMinutes()
		offset = func(int64) int64 { return kept }
	}
	if nsModel, err = withUtcOffset(data, nsModel, offset); err != nil {
		return nil, false, nsutils.BadDataBodyErrorWithHint(err.Error())
	}
	// One created with an identifier none has repeats one with its fallback key, for a key that finds them so
	if replaced == nil {
		duplicate, err := c.keyDuplicate(q, ctx, nsModel, true)
		if err != nil {
			slog.ErrorContext(ctx, "failed to check for duplicate in database", "error", err, "collection", m.Name())
			return nil, false, nsutils.ErrInternal
		}
		if duplicate != nil {
			return c.repeated(ctx, siteContext, tx, q, nsModel, data, duplicate)
		}
	}

	dbCreateModel, err := m.CreateModel(q, ctx, nsModel, converterContext(siteContext), &identifier)
	if httpErr := c.stored(dbCreateModel, err, "replace"); httpErr != nil {
		return nil, false, httpErr
	}
	if err := q.MarkStale(ctx, replaced); err != nil {
		slog.ErrorContext(ctx, "failed to mark summaries stale", "error", err, "collection", m.Name())
		return nil, false, nsutils.ErrInternal
	}

	if httpErr := c.commit(tx, "replace"); httpErr != nil {
		return nil, false, httpErr
	}

	creatingNewRecord := replaced == nil
	if creatingNewRecord {
		siteContext.PublishModelChange(events.Created, dbCreateModel)
	} else {
		siteContext.PublishModelChange(events.Updated, dbCreateModel)
	}
	c.afterStored(ctx, siteContext, nsModel)
	return c.document(dbCreateModel, siteContext), creatingNewRecord, nil
}

// removeReplaced is the record a replace takes the place of, deleted for good so its replacement can take its
// identifier. It is nil when there is none, so the replace creates the record, which only canCreate allows.
func (c Collection[NS, DM]) removeReplaced(ctx context.Context, q *sitedb.Queries, identifier string, canCreate bool) (*DM, *nsutils.HTTPError) {
	replaced, err := c.db.GetModel(q, ctx, identifier, true)
	if err == sql.ErrNoRows {
		if !canCreate {
			return nil, nsutils.ErrNotFound
		}
		return nil, nil
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to get by ID", "error", err, "collection", c.db.Name(), "identifier", identifier)
		return nil, nsutils.ErrInternal
	}
	if nsutils.ModifiedSince(ctx, (*replaced).GetUpdatedAt()) {
		return nil, nsutils.ErrModifiedSince
	}
	if _, err := c.db.Delete(q, ctx, identifier, true); err != nil {
		slog.ErrorContext(ctx, "failed to hard delete in database", "error", err, "collection", c.db.Name(), "identifier", identifier)
		return nil, nsutils.ErrInternal
	}
	return replaced, nil
}

// validateReplace checks the model replacing the record replaced, which is nil when the replace creates the record.
// As in Nightscout, a deleted document is replaced whatever changed, as a client like Juggluco changes one by
// deleting it and sending it again.
func (c Collection[NS, DM]) validateReplace(ctx context.Context, nsModel NS, replaced *DM, data map[string]any, identifier string) *nsutils.HTTPError {
	if replaced != nil && (*replaced).GetDeletedAt() != nil {
		replaced = nil
	}
	if err := ValidateNotChangingSharedFields(replaced, nsModel, data, identifier, nsutils.DateMovable(ctx)); err != nil {
		slog.ErrorContext(ctx, "validation failed for replacing model", "error", err, "collection", c.db.Name())
		return err
	}
	if err := c.db.Validate(nsModel, replaced); err != nil {
		slog.ErrorContext(ctx, "validation failed for replacing model", "error", err, "collection", c.db.Name())
		return err
	}
	return nil
}

// Patch merges the fields sent into the stored model, see rewrite, once it is found not deleted, not modified since the
// client read it, and the fields sent do not change those a document keeps
func (c Collection[NS, DM]) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, data map[string]any) (nsutils.Document, *nsutils.HTTPError) {
	m := c.db
	if _, ok := any(m).(CollectionPatcher[NS, DM]); !ok {
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
	if nsutils.ModifiedSince(ctx, (*updateingModel).GetUpdatedAt()) {
		return nil, nsutils.ErrModifiedSince
	}

	if err := ValidateNotChangingSharedFields(updateingModel, nsModel, data, identifier, false); err != nil {
		slog.ErrorContext(ctx, "validation failed for patching model", "error", err, "collection", m.Name())
		return nil, err
	}

	dbModel, _, httpErr := c.rewrite(ctx, siteContext, q, updateingModel, data)
	if httpErr != nil {
		return nil, httpErr
	}
	if err := q.MarkStale(ctx, updateingModel); err != nil {
		slog.ErrorContext(ctx, "failed to mark summaries stale", "error", err, "collection", m.Name())
		return nil, nsutils.ErrInternal
	}

	if httpErr := c.commit(tx, "patch"); httpErr != nil {
		return nil, httpErr
	}

	siteContext.PublishModelChange(events.Updated, dbModel)
	return c.document(dbModel, siteContext), nil
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

// ValidateNotChangingSharedFields refuses a change to the fields a document keeps, its date and UTC offset too unless
// dateMovable, as an API v1 replace may move them
func ValidateNotChangingSharedFields[NS SharedNSFields, DM SharedDBFields](replacingModelPtr *DM, nsModel NS, nsModelMap map[string]any, identifier string, dateMovable bool) *nsutils.HTTPError {
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
	if !dateMovable && nsModel.GetDate() != nil && replacingModel.GetOccurredAt() != 0 && *nsModel.GetDate() != replacingModel.GetOccurredAt() {
		return nsutils.ErrCannotChangeDate
	}

	if !dateMovable && nsModel.GetUtcOffsetMinutes() != nil && replacingModel.GetUtcOffsetMinutes() != nil && *nsModel.GetUtcOffsetMinutes() != *replacingModel.GetUtcOffsetMinutes() {
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

// stored is the error from storing a model with CreateModel or RewriteModel. Data the database refuses is the client's error.
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

// withUtcOffset is the model with the offset at its date when it was sent with a date but no utcOffset, which
// would otherwise be stored as 0
func withUtcOffset[NS SharedNSFields](data map[string]any, nsModel NS, offset func(date int64) int64) (NS, error) {
	if nsModel.GetUtcOffsetMinutes() != nil || nsModel.GetDate() == nil {
		return nsModel, nil
	}
	data["utcOffset"] = offset(*nsModel.GetDate())
	var withOffset NS
	err := utils.ConvertViaJSON(data, &withOffset)
	return withOffset, err
}

// siteUtcOffset is the offset in minutes of the site's timezone at a date, as a document sent with none is in it,
// as Nightscout takes the server's
func siteUtcOffset(siteContext *httputils.SiteContext) func(date int64) int64 {
	location := siteContext.GetSiteTimezone()
	return func(date int64) int64 {
		_, seconds := time.UnixMilli(date).In(location).Zone()
		return int64(seconds / 60)
	}
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
