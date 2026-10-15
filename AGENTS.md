
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

Keep every change as small as possible.

- A single change must modify no more than 100 lines of code (not including generated code, not including test code). If a change requires more changed lines, stop and ask for approval before proceeding.
- Do not edit more than 50 files in a single change.
- Do not reformat surrounding code.
- Do not reformat surrounding code or bundle unrelated fixes; those are their own PR.
- Before writing a function, look for an equivalent in `utils` or an imported library.
- Any UI change must be accompanied with 

## UI Changes

- During go test tests/bdd we store screenshots in tests/screenshots. Any UI test must commit these screenshots folder to show the changes