# Security Policy

Nighttroop holds people's diabetes data, so we take security reports seriously and thank everyone who sends one.

## Supported versions

Only the latest release and `main` get security fixes. Releases are the `v*` tags, e.g. `v1.2.3`, each published as a
Docker image tagged with its version and `latest` (see `.github/workflows/docker.yml`). If you run your own server,
update to the newest release to get a fix.

## Reporting a vulnerability

Report it privately, in one of two ways:

1. On GitHub, with **Report a vulnerability** on the repo's
   [Security tab](https://github.com/grahamjenson/nighttroop/security/advisories/new).
1. By email to <nighttroop@t1d.nz>.

Please:

- Do not open a public issue, pull request or discussion about it.
- Only test against your own local instance, or a site you made yourself on the staging server
  (<https://staging.nighttroop.nz>).
- Do not access, change or delete other people's sites or data, including other testers' sites on staging.
- Do not run denial of service, load or brute force tests against any shared server.

## What to include

- What the vulnerability is, and what an attacker could do with it.
- The steps to reproduce it, e.g. the requests sent, and the version or commit you tested.
- Any fix you suggest.
- Whether, and how, you would like to be credited.

## What to expect

- We acknowledge your report within a few days.
- We confirm the issue, and keep you updated while we fix it.
- We release the fix and publish a GitHub security advisory, crediting you if you want.

Nighttroop is maintained by volunteers, so we cannot promise fix times, and there is no bug bounty.

## Scope

In scope is anything that lets one user reach what they should not, for example:

- Cross site data exposure: one site reading or changing another site's data, or learning its subdomain or site ID.
- Authentication or authorisation bypass: as a site, an identity, or the control admin.
- SSRF: making the server send requests to hosts it should not, e.g. through the Nightscout, Dexcom or Libre
  connectors.
- Leaks of PII or credentials: Dexcom usernames, Nightscout URLs and tokens, profile names or notes, in pages,
  APIs, logs, errors or backups.
- Weak encryption of secrets at rest or in backups.
- A cheap request that makes the server do a lot of work, e.g. many rows, expensive queries or bcrypt calls, getting
  past the rate limits and caps.

Out of scope are issues that need the server's operator to be careless, such as using the dev keys below, and
reports from automated scanners with no working attack.

## Self hosting

If you run your own Nighttroop:

- Make your own `KEY_SETS` with `go run cmd/keysets/create_keysets.go`, and keep a copy somewhere safe.
- Never use the `KEY_SETS` in `.env.dev` or `.env.test`: they are public, so anyone could read your secrets.
- Set `MODE=production`. Development and test use cheap bcrypt, fixed site names and fixed keys.

See [docs/config.md](docs/config.md) and [docs/deploy.md](docs/deploy.md) for more.
