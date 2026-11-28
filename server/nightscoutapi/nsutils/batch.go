package nsutils

import (
	"context"
	"database/sql"
	"database/sql/driver"
)

// Batch stores an upload's documents in one transaction of the site's database, so it syncs to disk once for them all,
// not once for each. A document's Commit waits for the batch's, as does what is done once it is committed (Stored). Its
// Rollback rolls the batch back, as nothing undoes one document's writes alone: End is then false, and its caller
// stores the documents again, each in a transaction of its own.
type Batch struct {
	tx                *sql.Tx
	committed, Undone bool
}

type batchKey struct{}

// In is ctx with the batch, which the documents stored with it are in
func (b *Batch) In(ctx context.Context) context.Context { return context.WithValue(ctx, batchKey{}, b) }

// BatchOf is the batch of ctx, nil when it has none
func BatchOf(ctx context.Context) *Batch {
	batch, _ := ctx.Value(batchKey{}).(*Batch)
	return batch
}

// Begin begins a document's write to db: a transaction of its own with no batch, or else its part of the batch's
// transaction, which the batch's first write begins
func (b *Batch) Begin(ctx context.Context, db *sql.DB) (driver.Tx, *sql.Tx, error) {
	if b == nil || b.tx == nil {
		tx, err := db.BeginTx(ctx, nil)
		if b == nil || err != nil {
			return tx, tx, err
		}
		b.tx = tx
	}
	b.committed = false
	return b, b.tx, nil
}

// Commit of a document's write waits for the batch's
func (b *Batch) Commit() error {
	b.committed = true
	return nil
}

// Rollback of a document's write rolls the batch back, so the site's database can be read again at once, unless it
// follows the document's Commit, as a deferred one does
func (b *Batch) Rollback() error {
	if b.committed {
		return nil
	}
	b.Undone = true
	return b.tx.Rollback()
}

// pending stands for a document until its batch ends
type pending struct {
	Document
	made func() Document
}

// Stored is the document made, with what else is done once its write is committed. In a batch, which holds the site's
// database until it ends, End makes it, and one standing for it is given.
func (b *Batch) Stored(made func() Document) Document {
	if b == nil {
		return made()
	}
	return pending{made: made}
}

// End commits the batch, then makes each of documents standing for one stored in it, in their order. It is false, with
// no document made, when the batch was rolled back or cannot be committed.
func (b *Batch) End(documents []Document) bool {
	if b.Undone || (b.tx != nil && b.tx.Commit() != nil) {
		return false
	}
	for i, document := range documents {
		if waiting, ok := document.(pending); ok {
			documents[i] = waiting.made()
		}
	}
	return true
}
