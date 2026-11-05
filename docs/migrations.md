# Migrations

Each DB's schema is made by the numbered SQL files in its `data/sql/<db>/migrations` folder, run in order with
[golang-migrate](https://github.com/golang-migrate/migrate) each time the server opens the DB. A DB keeps the version
it reached in its `schema_migrations` table and only runs the files numbered after it.

## Until the first production deploy: `000001` can change

Until then each DB's whole schema is its `000001_init_schema.up.sql`, and a schema change edits it. Staging, the only
server with data, deletes its DBs when one cannot be migrated and prepared (`scripts/try_migrate_or_delete`), so it
starts again from the new `000001`. A change that leaves every query working, like a changed `CHECK`, does not reset
it, so staging keeps the old rule until its DBs are reset.

## From the first production deploy: never change a migration that is on main

A DB at version 1 never runs `000001` again, so a change to it would only reach DBs made afterwards. Every DB made
before would keep the old schema: an older control DB cannot start ("no such column"), and an older site's DB fails
some queries or keeps rules, like a `CHECK`, the code no longer expects.

So the first production deploy freezes the migrations: it adds `data/sql/MIGRATIONS_FROZEN`. Once that is on a pull
request's base branch, CI fails the pull request when it changes, deletes or renames a migration on the base branch
(`scripts/check_migrations.sh`, which can be run locally the same way). From then on, to change a schema, add the next
migration:

1. Add `data/sql/<db>/migrations/00000N_<what>.up.sql`, numbered after the last, and its `00000N_<what>.down.sql`.
   `go generate` migrates down as a check, so each needs a down file; it can be empty when the change cannot be undone.
2. Run `go generate ./...`, which applies every migration in order to make `schema_gen.sql` for sqlc. It also checks
   the migrations with rows in (`data/sql/gen/migration_check.go`): at each version it adds rows with every column at
   the largest and the smallest its CHECKs allow, the longest text, both ends of a range, the first and last of a list,
   then migrates down a version and up again, and finally down to nothing, all on a DB on disk with foreign keys on.
   So a migration that narrows a CHECK, or a down migration that cannot take its version's rows, fails here rather
   than on a deploy. A CHECK it cannot read is not tried at its limit, and one its rows break fails the check, naming
   the table; teach `migration_check.go` the new form.
   Each step up or down must also keep every row of a table there before and after, in the columns it had before and
   after, so a migration that loses rows, or makes a table again with its columns in another order and copies them
   with `SELECT *`, fails. A migration that means to change a table's rows says so with a line
   `-- changes rows of <table>`. Each text column's rows are made of its own letter so swapped text columns differ;
   numbers are not, as CHECKs compare them across columns, so name the columns copied when any are swapped.
3. Prefer additive changes: a new table, or a new column that can be NULL or has a default. A build cannot open a DB
   a newer build migrated, so an additive change is the one a rollback copes with best.

## What a migration runs in

Each file runs in one transaction, on a connection with foreign keys on. A statement that fails rolls the whole
file back, but golang-migrate has already marked the version dirty, and every open of a dirty DB fails ("Dirty
database version N"). As the file was rolled back, none of N was applied: once its bug is fixed, the DB can be
forced back to version N-1 and migrated again.

## Changing a table

SQLite cannot change a column or a `CHECK` in place, so the table is made again. The usual way, renaming a new table
over the old one, fails here: the transaction cannot turn foreign keys off, and tables such as `event_groups`
reference the event tables `ON DELETE RESTRICT`. Instead, in one migration:

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

The table is made again under its own name, so views that select from it, like `temp_basals_vw`, still work, and
rows that reference it still do. Columns copied with `SELECT *` must be in the same order in the new definition; name
them otherwise.

## When a migration fails on a deploy

- **A DB that cannot be opened stops the start**, as does a missing site restored from its backup that cannot be:
  its files are kept, and with `update.sh` the release is rolled back, putting `./data` back from the copy it made.
- **Fix forward**: a new release with a fixed or new migration opens the kept DBs. A dirty one is first forced back
  to the version before the one that failed, e.g. `sqlite3 <db> "UPDATE schema_migrations SET version = N-1, dirty = 0"`
  with the server stopped.
- **Rolling back after a migration ran** needs `./backups/data-*.tar.gz` from `update.sh`, or S3's backups, as the
  older build cannot open DBs at the newer version, and loses what was written since.
- **Staging** (`scripts/try_migrate_or_delete`) deletes every DB when one cannot be migrated and prepared, so it starts
  empty rather than with only some.
