# Nighttroop

[![CI](https://github.com/grahamjenson/nighttroop/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/grahamjenson/nighttroop/actions/workflows/ci.yml)

**Test it out** on the staging server, <https://staging.nighttroop.nz>, which runs the newest `main`. Its data is wiped
now and then, so keep nothing on it you need.

Nighttroop is a multi user rewrite of Nightscout.
Its core goals are to be:

1. Simple         - Single binary deployment, limited and optional external dependencies
1. Cheap Per User - It should consume very little resources per user
1. Multi-user     - A single server should support at least 10,000 users

To run your own Nighttroop look at the `docs/deploy` folder.

# GoTHSS Framework. (Go + Templ + HTMx + SQLite + SQLc + Tailwind)

The "framework" that Nighttroop uses uses Go and SQLc, relying heavily on generated code build by tools like SQLc, Templ and Tailwind.

# Security 

We generate the `tink-go` library to generate keysets for encryption, hashing, JWT with `go run utils/keysets/create_keysets.go`

These keysets are uploaded to the server to the ENVAR `KEY_SETS`

On initial startup you will have to set the control admin username, password and save and prove the TOTP.


