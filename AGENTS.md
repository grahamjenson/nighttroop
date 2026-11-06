
# Agent Development Guide

A file for [guiding coding agents](https://agents.md/).
DO NOT UPDATE THIS FILE UNLESS SPECIFICALLY ASKED TO DO SO.

## Commands

- **Generate** `go generate ./...`
- **Build:** `go build main.go`
- **Test:** `go test ./...`
- **Run** `docker-compose up -d && air` will run and rebuild as files change
- **Clean** `docker-compose 

## Directory Structure

- `background`: background jobs running with golang Cron and LiteQ-ish framework
- `clients`: third party client integration
- `data/sql`: database schema, migrations and queries generated with SQLc
- `data/db`: objects used to as clients to the dbs
- `data/events`: event bus code
- `docs`: posts and docs
- `models`: models for the various inputs, including converters built with goverter
- `scripts`: necessary scripts for generating the initial key bundle and other functions
- `server`: primary routers and controllers for the requests
- `socketio`: Socket.IO and Engine.IO (long-polling and websocket) server and client, for Nightscout live updates
- `tests`: integration tests across the site
- `utils`: all the useful stateless functions

## Generated Code

Do not change generated code. All generated files either:

1. has a `z_` prefix
1. has a `_temp.go` suffix
1. has a `.gen.go` suffix
1. has a `_gen.sql` suffix

Use code generators where possible, especially for building models and params for the DB.

- Run `go generate ./...` and commit what it changes; CI fails if generating changes any file.

## Config

- Adding, removing or changing a configuration environment variable (in `utils/config`) must update `docs/config.md` in the same change.

## Documents

When making a change kep these documents up to date:

- `docs/REQUIREMENTS.md` 
    - lists every feature, under Control, Site and Identity. A PR that adds, changes or removes a feature must update it in the same PR.
    - Write it as the file is written: one feature per line, as if to a quality tester, making sure everything works.

- `clients/nighttroop/openapi.json`
    - a openapi description of all the API endpoints and their responses which is used to generate a client for nighttroop
    - Use abstractions and models as much as possible

- `TODOs.md`
    - a list of active and prioritized development and feature goals.
    - update this list to remove any if not available, add any if you think something you just completed needs additional work
    - categorize the work and add priority values P0 being most important P5 being least.

## SQL

- Only ever use prepared statements: add a named query to `data/sql/*/queries` and call the generated sqlc method.
- Never run raw SQL strings (e.g. `ExecContext`, `QueryContext`), including in tests.
- A column holding JSON is named with `json` (or is a list of `errors` or `attributes`) and CHECKs `json_valid` on its line; `go generate` fails otherwise.
- The one exception is generated upkeep sqlc cannot express, such as each DB's `Vacuum`, in `data/sql/gen/db.go`.

## Models and the JSON API

- Model fields have their real names and types, which are also the JSON API: times are `int64` milliseconds, nullable fields are pointers, and JSON tags are snake_case.
- Glucose is stored in mg/dL. The site API returns and accepts it in the units from the settings (identity, then site, then default), with unit-neutral names (`reading`, `low_target`) and a `units` field.
- Searches page with `limit` and `skip` (at most 1000), newest first, and leave out soft-deleted rows. Deletes from the site are soft deletes.


## Issue and PR Guidelines

- All PRs should be against the main branch. No stacked PRs. Break it up into non-conflicting independent changes, or leave conflicting and we can resolve later.
- Never create an Issue
- If you create a PR always start PR description with bold letters saying "I AM A LITTLE AI HERE TO HELP" with the model name at the end
- One PR per feature or fix. When asked for several things, make separate PRs; stack a PR on the branch it depends on and say so in its description.
- A PR description says what changed and why, notes any approved size over the limit, and lists the tests.
- Remove code a change makes dead in the same PR.

## Change Size

Keep every change as small as possible.

- A single change must modify no more than 150 lines of code (not including generated code, not including test code). If a change requires more changed lines, stop and ask for approval before proceeding.
- Do not edit more than 50 files in a single change.
- Do not reformat surrounding code or bundle unrelated fixes; those are their own PR.
- Before writing a function, look for an equivalent in `utils` or an imported library.
- Any UI change must be accompanied with a BDD test to render and check the UI renders and works.
- When reducing size do not rely on lots of generics, a function should have a max of two generic types.

## UI Changes

- During go test tests/bdd we store screenshots in tests/screenshots. Any UI test must commit these screenshots folder to show the changes
- Only commit screenshots a `.bdd` file takes with `SCREENSHOT`.
- Use the theme colours (`primary`, `brand-night`, `brand-gold`) for branding, not raw Tailwind colours. Amber, blue and red stay for warnings, info and errors.
- Give every input an ID unique to its page (e.g. `site--glucose--create--reading`), so typed values do not carry between pages; BDD tests use these IDs.
- With htmx: errors go to `#toaster` with `hx-status:4xx`, success toasts go out-of-band with `<div hx-swap-oob="beforeend:#toaster">`
- Check items in a menu or scrolling box with `ASSERT_IN_VIEW`; `ASSERT_VISIBLE` passes for clipped elements.
- Check a mark in a graph, which other marks may be drawn over, with `ASSERT_DRAWN`; `ASSERT_VISIBLE` also needs it to be what a click at its centre would hit.

## Tests

We will rely more on the integration tests in the `/tests` folder, both API and BDD. 
The goal there is to first make sure that the user experience is never broken and working as intended.

A unit test is preferred if:

1. Too many options: A user path has a combinatorial explosion making it is too large a space for integration tests to be practical, e.g. checking all possible constraints on a model. In this case we would prefer unit tests of test matrices, building test structs and a for loop, this will make those many tests easier to read and debug.
1. Impractical: A function or user path is impractical to test, e.g. integration with other services, or testing side effects like logging or error handling. In these cases custom mocks, stubs, or dependency injection will need to be used to test those code pathways.
1. Async Timing: We don't want to wait for background job, cron or some other async task to run to confirm a test works. Either: either we implement a feature in a way that avoids this by e.g. calling immediately if test, or we write unit tests with of the async task, but confirm it is scheduled in the bdd test.


- If an Site tests depends on having clean data, create their own site `POST http://localhost/site`
- Socket.IO `.sio` tests in `tests/socket` use the `.http` lines plus `CONNECT`, `EMIT` (with `ACK`), `WAIT`, `COLLECT` and `DISCONNECT`, with `B` for a body and `T` for a time; see `tests/dsl/flow`.
- BDD tests get realistic data by enabling virtual sync and running `RUN_BACKGROUND_JOBS` rather than spending a lot of cycles browsing to specific pages to enter data.
- For a bug fix, show the new test fails without the fix.
- Run the full `go test ./...` before opening a PR.

# Malicious users

Malicious users can cause havoc by:

1. creating lots of tables rows
1. running a lot of expensive queries
1. creating sync operations against non-affiliated parties causing nighttroop to be a proxy
1. forcing expensive CPU calls, e.g. auth or create identity via bcrypt

We stop these by having sane limits that should never be hit by a user in normal operation, like:

1. Rate limiting read and write operations against sites, identities, and ips
1. Slowing down background job frequencies if not behaving
1. Refusing values far outside real use, e.g. glucose above 1000 mg/dL or a temp target longer than a week
1. Capping how many rows one query returns
1. For tables that require a SiteID like identity we pull it directly from the siteContext and never from a users input.

When writing code, make sure to consider how an operation may be used by malicious users.

# Privacy

## PII
We aim to not store any PII or user data that could be used to identify a user. This is not possible because:

1. Profile names or notes may have personal information in them
1. Dexcom username has to be recorded for sync to work
1. Nightscout URL or token may have personal information in them

Dexcom username and Nightscout URL and their credentials should also be encrypted at rest and be treated very carefully.

Profile names and notes, and other fields that may have user data inserted should be treated very carefully as well.

## Cross Site Information

We also aim to never expose a users SiteID/Subdomain or any information of another site to any other user. 
Any API/Model e.g. Logs/Events/Errors/Sync Reports should be cleaned of any such information that may be visible to another, e.g. reports on user batch backups.
