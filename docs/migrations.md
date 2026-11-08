# Migrations

Each DB's schema is made by the numbered SQL files in its `data/sql/<db>/migrations` folder, run in order with
[golang-migrate](https://github.com/golang-migrate/migrate) each time the server opens the DB. A DB keeps the version
it reached in its `schema_migrations` table and only runs the files numbered after it.

## Until the first production deploy: `000001` can change

Until then each DB's whole schema is its `000001_init_schema.up.sql`, and a schema change edits it. Staging, the only
server with data, deletes its DBs when one cannot be migrated and prepared (`cmd/try_migrate_or_delete`), so it
starts again from the new `000001`. It also deletes them when one's schema (each table, index, trigger and view's SQL
in `sqlite_master`) is not the one the migrations make on a new DB, so a change that leaves every query working, like
a changed `CHECK` or even a changed comment in a `CREATE TABLE`, resets staging too, rather than leaving it the old rule.

## From the first production deploy: never change a migration that is on main

A DB at version 1 never runs `000001` again, so a change to it would only reach DBs made afterwards. Every DB made
before would keep the old schema: an older control DB cannot start ("no such column"), and an older site's DB fails
some queries or keeps rules, like a `CHECK`, the code no longer expects.

So the first production deploy freezes the migrations: it adds `data/sql/MIGRATIONS_FROZEN`. Once that is on a pull
request's base branch, CI fails the pull request when it changes, deletes or renames a migration on the base branch
(`cmd/check_migrations.sh`, which can be run locally the same way). From then on, to change a schema, add the next
migration:

1. Add `data/sql/<db>/migrations/00000N_<what>.up.sql`, numbered after the last, and its `00000N_<what>.down.sql`.
   `go generate` migrates down as a check, so each needs a down file; it can be empty when the change cannot be undone.
   A DB only runs migrations numbered after its version, so one numbered at or before the last on the base branch, as
   when another pull request's migration merged first, would never run on DBs already past it. CI fails a pull
   request that adds one (`cmd/check_migrations.sh`, frozen or not); renumber it after the last. CI only runs on a
   push, so bring the base branch in, or re-run the check, before merging when another migration merged meanwhile.
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
   a newer build migrated, so once a release has served, a problem with it is fixed forward, not rolled back (see
   [When a migration fails on a deploy](#when-a-migration-fails-on-a-deploy)), and an additive change is the easiest
   to fix forward from.

## What a migration runs in

Each file runs in one transaction, on a connection with foreign keys on. So `go generate` fails a migration that runs
a PRAGMA other than `PRAGMA defer_foreign_keys`: in a transaction some do nothing without an error, such as
`PRAGMA foreign_keys = OFF`, and the rest, such as `ignore_check_constraints`, last past the migration on its
connection, which the server goes on to use. `BEGIN`, `COMMIT` and `VACUUM` fail the migration already. A statement that fails rolls the whole
file back, but golang-migrate has already marked the version dirty, and refuses to migrate a dirty DB. As the file was
rolled back, none of N was applied, so each open of a DB dirty at N forces it back to version N-1 (or to none, when N
is the first), logs a warning naming the DB and both versions, and migrates it again
(`data/sql/migrateup`). A migration that still fails stops the open with an error naming the DB, the migration and
why it failed. A DB dirty at a version its build has no migration for is not forced back, as which was applied is not
known.

The one case this gets wrong: golang-migrate clears the dirty mark after the file's transaction commits, so a server
that dies between the two leaves N applied but marked dirty, and the next open runs N again. A migration of only
`CREATE`s then fails on what it already made, and stops the open with that error to fix by hand; one that changes rows,
such as an `UPDATE`, would change them twice. Keep a migration that changes rows safe to run twice, e.g. with a
`WHERE` that only matches rows it has not changed yet.

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
them otherwise. `go generate` runs the migration on a DB with rows in it (step 2 above), so a rebuild that loses rows,
or copies them into the wrong columns, fails there rather than on a deploy.

## When a migration fails on a deploy

- **A DB that cannot be opened stops the start**, as does a live site's DB that is missing and not restored from its
  backup, or restored and cannot be opened: the server serves nothing, so no data is written that a rollback would
  lose. Its files are kept, and with `update.sh` the release is rolled back, putting `./data` back from the copy it
  made. The error names the DB or the sites.
- **A rollback is only safe before the new release has served.** `update.sh` does it on its own: the server only
  serves once every DB is opened and migrated, so a release whose migration fails is never healthy, takes no data,
  and its rollback loses nothing.
- **After the new release has served, fix forward**: release a fixed or new migration, or a fix to the code, rather
  than putting back `./data`. The older build cannot open DBs a newer one migrated, so going back needs
  `./backups/data-*.tar.gz` from `update.sh`, or S3's backups, and loses everything written since.
- **Fixing forward after a failed migration**: a new release with a fixed or new migration opens the kept DBs. One the
  failed migration left dirty is forced back to the version before it and migrated again on its own, so nothing needs
  to be done by hand.
- **Staging** (`cmd/try_migrate_or_delete`) deletes every DB when one cannot be migrated and prepared, or its schema
  is not the one the migrations make on a new DB, so it starts empty rather than with only some, or with old rules.
