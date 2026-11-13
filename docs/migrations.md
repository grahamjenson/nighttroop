# Migrations

- Each DB's schema comes from the numbered SQL files in `data/sql/<db>/migrations`.
- [golang-migrate](https://github.com/golang-migrate/migrate) runs them in order each time the server opens the DB.
- A DB keeps its version in `schema_migrations`, and runs only the files after it.

## Until the first production deploy: `000001` can change

- Each DB's whole schema is its `000001_init_schema.up.sql`. Edit it to change the schema.
- Staging, the only server with data, deletes its DBs (`cmd/try_migrate_or_delete`) and starts from the new `000001`
  when a DB:
  - cannot be migrated and prepared, or
  - has a schema (each table, index, trigger and view's SQL in `sqlite_master`) unlike the one the migrations make on
    a new DB. Thus even a changed `CHECK` or `CREATE TABLE` comment resets staging.

## From the first production deploy: never change a migration that is on main

- A DB at version 1 never runs `000001` again. Older DBs keep the old schema: an older control DB cannot start ("no
  such column"), and an older site DB fails some queries or keeps old rules (example: a `CHECK`).
- The first production deploy adds `data/sql/MIGRATIONS_FROZEN`. Once it is on a pull request's base branch, CI fails
  a pull request that changes, deletes or renames a base branch migration (`cmd/check_migrations.sh`, also runs
  locally).

### To change a schema

1. Add `data/sql/<db>/migrations/00000N_<what>.up.sql`, numbered after the last on the base branch.
   - A DB runs only migrations after its version. A lower number (example: another pull request's merged first)
     never runs on DBs past it. CI fails it (`cmd/check_migrations.sh`, frozen or not): renumber it.
   - CI runs only on a push. If another migration merged meanwhile, bring the base branch in, or run the check again,
     before you merge.
2. Add `00000N_<what>.down.sql`, empty if the change cannot be undone. `go generate` migrates down as a check.
3. Run `go generate ./...` (see [What `go generate` checks](#what-go-generate-checks)).
4. Prefer additive changes: a new table, or a new column that can be NULL or has a default. They are the easiest to
   fix forward from, as a build cannot open a DB a newer build migrated (see
   [When a migration fails on a deploy](#when-a-migration-fails-on-a-deploy)).

### What `go generate` checks

- It applies every migration in order to make `schema_gen.sql` for sqlc.
- It tests them on a DB on disk with foreign keys on (`data/sql/gen/migration_check.go`):
  1. At each version, it adds rows with every column at the largest and smallest its CHECKs allow (longest text,
     both ends of a range, first and last of a list), then migrates down a version and up again.
  2. At the end, it migrates down to nothing.
- Thus a migration that narrows a CHECK, or a down migration that cannot take its version's rows, fails here, not on a
  deploy.
- A CHECK it cannot read is not tried at its limit. One its rows break fails, naming the table: teach
  `migration_check.go` the new form.
- Each step must keep the rows of each table there before and after, in its columns. Thus losing rows, or copying
  reordered columns with `SELECT *`, fails.
- To change a table's rows on purpose, add the line `-- changes rows of <table>`.
- Each text column's rows use their own letter, so swapped text columns differ. Numbers do not, as CHECKs compare them
  across columns: if any columns are swapped, name the columns copied.

## What a migration runs in

- Each file runs in one transaction, on a connection with foreign keys on.
- `go generate` fails a PRAGMA other than `PRAGMA defer_foreign_keys`. In a transaction, some do nothing and give no
  error (example: `PRAGMA foreign_keys = OFF`). The rest (example: `ignore_check_constraints`) stay on the connection,
  which the server then uses.
- `BEGIN`, `COMMIT` and `VACUUM` already fail.

### Dirty DBs

A failed statement rolls the whole file back, so none of N is applied. But golang-migrate has already marked the
version dirty, and refuses a dirty DB. Thus each open of a DB dirty at N (`data/sql/migrateup`):

1. forces it back to N-1 (or to none, if N is the first);
2. logs a warning naming the DB and both versions;
3. migrates it again. If that fails, the open stops with an error naming the DB, the migration and the cause.

A DB dirty at a version its build has no migration for is not forced back: which was applied is not known.

**Caution:** Make a migration that changes rows (example: an `UPDATE`) safe to run twice (example: a `WHERE` that
matches only unchanged rows). golang-migrate clears the dirty mark after the commit. If the server stops between the
two, N is applied but dirty, and the next open runs it again. A migration of only `CREATE`s then fails and stops the
open, to fix by hand.

## Changing a table

SQLite cannot change a column or a `CHECK` in place, so the table is made again. Renaming a new table over the old
one fails here: the transaction cannot turn foreign keys off, and tables such as `event_groups` reference the event
tables `ON DELETE RESTRICT`. Instead, in one migration:

```sql
-- Foreign keys are checked at commit, when every row referenced is back
PRAGMA defer_foreign_keys = ON;
CREATE TEMP TABLE temp_basals_copy AS SELECT * FROM temp_basals;
DROP TABLE temp_basals;
CREATE TABLE temp_basals (...);              -- its new definition, under its own name
INSERT INTO temp_basals SELECT * FROM temp_basals_copy;
DROP TABLE temp_basals_copy;
CREATE INDEX temp_basals_occurred_at_idx ON temp_basals(occurred_at);  -- dropping the table dropped its indexes
```

- The table keeps its name, so views on it (example: `temp_basals_vw`) and rows that reference it still work.
- Columns copied with `SELECT *` must keep their order. If not, name them.
- `go generate` catches a rebuild that loses rows or puts them in the wrong columns.

## When a migration fails on a deploy

- **A DB that cannot be opened stops the start.** So does a live site's DB that is missing and not restored, or
  restored and cannot be opened.
  - The server serves nothing, so it writes nothing a rollback would lose. The files are kept.
  - `update.sh` rolls the release back, putting `./data` back from its copy.
  - The error names the DB or the sites.
- **Roll back only before the new release has served.** `update.sh` does it automatically. The server serves only once
  every DB is opened and migrated, so a failed release takes no data, and its rollback loses nothing.
- **After it has served, fix forward** with a fixed or new migration, or a code fix. Do not put back `./data`: the
  older build cannot open the DBs. Going back needs `./backups/data-*.tar.gz` from `update.sh`, or S3's backups, and
  loses everything written since.
- **A fix forward** opens the kept DBs. A DB left dirty is forced back a version and migrated again, with nothing to do
  by hand.
- **Staging** deletes every DB in the cases in
  [Until the first production deploy](#until-the-first-production-deploy-000001-can-change), so it starts empty, not
  with only some DBs or old rules.
