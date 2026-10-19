
# Agent Development Guide

A file for [guiding coding agents](https://agents.md/).

## Commands

- **Generate** `go generate ./...`
- **Build:** `go build main.go`
- **Test:** `go test ./...`
- **Run** `docker-compose up -d && air` will run and rebuild as files change
- **Clean** `docker-compose 

## Directory Structure

- `background`: background jobs running with golang Cron and LiteQ-ish framework
- `clients`: third party client integration
- `data`: everything SQLite using SQLc generated code
- `docs`: posts and docs
- `models`: models for the various inputs, including converters built with goverter
- `scripts`: necessary scripts for generating the initial key bundle and other functions
- `server`: primary routers and controllers for the requests
- `tests`: integration tests across the site
- `utils`: all the useful stateless functions

## Generated Code

Do not change generated code. All generated files either:

1. has a `z_` prefix
1. has a `_temp.go` suffix
1. has a `.gen.go` suffix
1. has a `_gen.sql` suffix

## SQL

- Only ever use prepared statements: add a named query to `data/sql/*/queries` and call the generated sqlc method.
- Never run raw SQL strings (e.g. `ExecContext`, `QueryContext`), including in tests.

## Issue and PR Guidelines

- Never create an Issue
- If you create a PR always start PR description with bold letters saying "I AM A LITTLE AI HERE TO HELP"

## Change Size

Keep every change as small as possible.

- A single change must modify no more than 100 lines of code (not including generated code, not including test code). If a change requires more changed lines, stop and ask for approval before proceeding.
- Do not edit more than 50 files in a single change.
- Do not reformat surrounding code.
- Do not reformat surrounding code or bundle unrelated fixes; those are their own PR.
- Before writing a function, look for an equivalent in `utils` or an imported library.
- Any UI change must be accompanied with a BDD test to render and check the UI renders and works.

## UI Changes

- During go test tests/bdd we store screenshots in tests/screenshots. Any UI test must commit these screenshots folder to show the changes

## Tests

We will rely more on the integration tests in the `/tests` folder, both API and BDD. 
The goal there is to first make sure that the user experience is never broken and working as intended.

A unit test is preferred if:

1. Too many options: A user path has a combinatorial explosion making it is too large a space for integration tests to be practical, e.g. checking all possible constraints on a model. In this case we would prefer unit tests of test matrices, building test structs and a for loop, this will make those many tests easier to read and debug.
1. Impractical: A function or user path is impractical to test, e.g. integration with other services, or testing side effects like logging or error handling. In these cases custom mocks, stubs, or dependency injection will need to be used to test those code pathways.
1. Async Timing: We don't want to wait for background job, cron or some other async task to run to confirm a test works. Either: either we implement a feature in a way that avoids this by e.g. calling immediately if test, or we write unit tests with of the async task, but confirm it is scheduled in the bdd test.
