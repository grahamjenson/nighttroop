
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

## Issue and PR Guidelines

- Never create an Issue
- If you create a PR always start PR description with bold letters saying "I AM A LITTLE AI HERE TO HELP"

## Change Size

Small PRs are the goal. Prefer several small PRs over one large one.

- Budget: at most 100 changed lines per PR, counted as added plus removed
  lines of hand-written code. Generated files (see above) do not count.
- Check before pushing: `scripts/changed_lines.sh` prints the count against
  `origin/main` and exits non-zero over budget. CI runs it on every PR.
- Over budget? Split the work: ship the first independently mergeable slice
  as its own PR, stack the next slice on it, and name the next slice in the
  PR description. Do not ask for a bigger budget.
- Do not reformat surrounding code or bundle unrelated fixes; those are
  their own PR.
- Before writing a function, look for an equivalent in `utils` or an
  imported library.

## UI Changes

- Any PR that changes how a page looks includes before and after
  screenshots of that page, rendered at desktop width, in its description
  (inline images or a link to a published comparison).
