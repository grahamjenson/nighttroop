# Nighttroop SQL

Nighttroop uses generated code with SQLc, goverter and a few custom scripts to create the data-layer interface.
All queries in Nighttroop are prepared statements, making it possible to test some aspects of their correctness at compile time.

Having all SQL explicit and in sql, rater than ORM-ish, hopefully makes all interactions more obvious and easier to optimize in the future.

## Keep a query prepared

- The server prepares each query one time and then uses it again. If SQLite reads a bound value when it plans a query, it prepares the query again on each run.
- SQLite reads these bound values when it plans:
  - a bare `LIMIT` or `OFFSET` value. `go generate` fails for it.
  - a bare value that a column equals (example: `group_type = ?`), in each query except a lookup on one table by its primary key or a unique index.
- Put the value in a `CAST` (example: `group_type = CAST(sqlc.arg('GroupType') AS TEXT)`). SQLite does not read a `CAST` when it plans, and it uses the same index.
- A query that SQLite prepares again takes 3 to 6 times as long (`BenchmarkReadsByKey` in `data/db`).

# Control DB

This DB has all the 

# Site DB

This has information that is related to the user, typically mastered elsewhere like AAPS or Loop.
It is the 