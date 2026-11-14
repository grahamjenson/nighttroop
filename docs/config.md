# Configuration

## How the server reads a variable

The server reads environment variables in `utils/config/config.go`, from the first of:

1. A command line flag (example: `-HOST_PORT=8080`). Only `MODE`, `HOST_NAME`, `HOST_PORT` and `HOST_PROTOCOL` have one.
2. The environment.
3. The mode's `.env` file: `.env.dev` in development, `.env.test` in test, `.env` in production.
4. The default below.

- An unknown variable in the `.env` file stops the server.
- A limit per minute is a rate. Its burst is how many can be used at once before the rate applies.

## Server

| Variable | Default | Meaning |
| --- | --- | --- |
| `MODE` | None. `test` under `go test`. | Required, except under `go test`. See [Modes](#modes). |
| `HOST_NAME` | `nt.local.t1d.nz` | The domain of site subdomains (example: `abc.nt.local.t1d.nz`), for links and cookies. |
| `HOST_PORT` | `3000` | The listen port. Links to sites (example: the create site page's login) have it with `http`, and none with `https`, as a proxy takes HTTPS on 443. |
| `HOST_PROTOCOL` | `https` | `http` or `https`, for links and cookies. `http` only in development (which allows only `http`) and test. |
| `DATA_DIR` | `./.dbs` | The SQLite DBs (`db/`) and backups that wait to upload (`backup/`). |
| `DEPLOYMENT` | `development` | This server's own name (examples: `staging-nz`, `prod-voyager`): its Sentry environment, in its startup log. `MODE` sets how it operates. |
| `OPERATOR_NAME` | `T1DNZ` in development and test. Required in production and staging, so no server names the project's operator by mistake. | Who runs this server, in the terms of service and privacy policy. |
| `SUPPORT_EMAIL` | `nighttroop@t1d.nz` in development and test. Required in production and staging. | How users contact the operator: every page's footer, the terms of service and privacy policy. |
| `FEEDBACK_URL` | Empty | Where users give feedback (examples: a Google Form, GitHub issues). If set, every page's footer and the home page link to it. |

### Modes

- `development`: reads `.env.dev`. Only for local use. `air` passes `-MODE=development`.
- `test`: reads `.env.test`. Only under `go test`.
- `production`: reads `.env`.
- `staging`: reads `.env`. Operates as `production`, but:
  - makes the control admin from `STAGING_ADMIN_PASSWORD` and `STAGING_ADMIN_TOTP_SECRET`, with no setup;
  - shows a large banner on every page: a test site whose data is frequently wiped.
- Any other value stops the server.
- `development` and `test`: cheap bcrypt, fixed site names and keys, every initial password `password` and TOTP `123123`.
- Only `development`, `test` and `staging` offer the virtual connector, which makes up glucose.

## Client IP and location

| Variable | Default | Meaning |
| --- | --- | --- |
| `TRUSTED_PROXIES` | `127.0.0.1,::1` | Comma-separated IPs or CIDR ranges of reverse proxies (example: Caddy) whose `X-Forwarded-For` is believed: the client is the last hop before them. On Fly.io, also the only peers whose `Fly-Client-IP` is believed: set it to where Fly's proxy connects from (see the staging `fly.toml`). |
| `FLY_APP_NAME` | Unset | Set by Fly.io. Then the client IP is the `Fly-Client-IP` header, if the TCP peer is in `TRUSTED_PROXIES` and the header is an IP. Other peers (example: Fly private network machines) can set it to anything. |
| `GEOIP_DB_PATH` | Empty. The Docker image sets its DB-IP IP to City Lite database. | A DB-IP IP to City Lite `.mmdb` file, for the city of a sign in. If missing or not readable, the built-in IP to Country Lite gives only the country. |
| `ALLOWED_COUNTRIES` | Empty: all | Comma-separated ISO 3166-1 alpha-2 codes (example: `NZ,AU`) of the only countries that can create sites, sign in and recover. |

`ALLOWED_COUNTRIES`:

- Addresses of unknown country, signed in sites and the Nightscout API are not limited.
- `utils/config/config.go` lists every DB-IP code beside `AllowedCountries`, to copy and remove from.
- A code that is not two letters stops the server.

## Staging

All are used only in staging.

| Variable | Default | Meaning |
| --- | --- | --- |
| `STAGING_ADMIN_PASSWORD` | Empty | Required. The control admin password, minimum 16 characters. Keep it secret, never in a committed file. |
| `STAGING_ADMIN_TOTP_SECRET` | Empty | Required. The control admin TOTP secret, minimum 16 base32 characters, also in your authenticator app. The admin is made only when the server starts without one, so a change does not change it. |
| `CAPTURE_REQUESTS` | Empty | `true` in staging. See [Request capture](#request-capture). |

### Request capture

Every request and response is kept as a `Captured <method> <path>` log in the logs DB, to show what clients (Loop,
AAPS, Trio) send.

- Each log has the site, identity, IP, headers and bodies (each up to 64 KB), and is up to 256 KB.
- Credentials and connector logins (examples: keys, TOTP codes, usernames, URLs) are `REDACTED` in headers, query,
  path and JSON or form bodies.
- Other bodies (a page, text, or over 64 KB): only size and content type.
- Styles, scripts, images and fonts: not kept.
- If the logs DB becomes too large, decrease `LOGS_DB_MAX_ROWS_PER_TABLE`.
- To read them, the control admin uses **Download Logs DB** on the control logs page.

## Security

| Variable | Default | Meaning |
| --- | --- | --- |
| `KEY_SETS` | Empty | Required, except in test mode (fixed keys). The four keys that encrypt secrets in the DBs and backups, and sign logins. |

- The keys in `.env.dev` and `.env.test` are public. Production and staging do not start with them.
- Make them with `cmd/keysets/create_keysets.go`.
- **Caution:** Keep a copy somewhere safe. Backups cannot be read without them.

## Request load

| Variable | Default | Meaning |
| --- | --- | --- |
| `MAX_REQUEST_BODY_BYTES` | `4000000` | The largest request body, above an uploader's largest batch (Juggluco: 5000 readings). Larger gets 413. |
| `REQUESTS_IN_FLIGHT_PER_CPU` | `16` | Requests each CPU works on at once. |
| `REQUESTS_BACKLOG_PER_CPU` | `64` | Requests for each CPU that wait, up to 5 seconds, for one in flight to finish. |

- A request that finds the backlog full, or waits longer, gets 503 and `Retry-After` at once. Thus a server that
  falls behind (example: a throttled CPU) answers what it can, and not every request times out.
- Not counted: the health check, live updates and Socket.IO.
- A quarter of both is kept for Nightscout API writes, so page loads do not cause refused device uploads.

## Storage and backups (S3)

| Variable | Default | Meaning |
| --- | --- | --- |
| `S3_URL` | None | The S3-compatible endpoint for backups and archives (example: `http://localhost:9000`, the docker-compose MinIO). |
| `S3_BUCKET` | `mybucket` | The bucket. Versioning and lifecycle rules are turned on at startup. |
| `S3_ACCESS_KEY` | `minioadmin` (docker-compose MinIO) | The access key ID. |
| `S3_SECRET_KEY` | `minioadmin` (docker-compose MinIO) | The secret key. |
| `CONTROL_BACKUP_INTERVAL_IN_SECONDS` | `30` | How often the control DB is backed up. |
| `BATCH_BACKUP_INTERVAL_IN_SECONDS` | `300` | How often each batch of site DBs is backed up. |
| `CONTROL_BACKUP_RETENTION_DAYS` | `90` | How long S3 keeps a control backup after a newer upload replaces it. |
| `SITE_BACKUP_RETENTION_DAYS` | `30` | The same, for a site batch backup or archive. |

- `S3_URL` unset or empty: S3 is off. No restore on start, no backup upload, and sites cannot be deleted.
- `S3_URL` set:
  - The server tests an upload at startup, and again over about half a minute. If it never works, the server stops.
  - Production and staging require both keys, and do not start with `minioadmin`.

## Sites and connectors

| Variable | Default | Meaning |
| --- | --- | --- |
| `MAX_SITES` | `10000` | The most sites the public can create from the home page. A control admin can create more. |
| `MAX_IDENTITIES_PER_USER` | `100` | The most identities on one site. Rate limiters are kept for 20 a site, so more costs no more memory. |
| `DELETE_AFTER_INACTIVITY_MINS` | `10080` (a week) | How long a never used site is kept before deletion. |
| `DISABLE_AFTER_INACTIVITY_DAYS` | `7` | How long a site is not written or seen before control can disable it. |
| `BACKGROUND_CONNECTOR_WORKERS` | `5` | Syncs that run at once. |
| `CONNECTOR_INTERVAL_IN_SECONDS` | `5` | How often due syncs are queued. |
| `BACKFILL_DAYS` | `7` | How far before a site's oldest glucose reading a backfill gets records. |

## Logs and alerts

| Variable | Default | Meaning |
| --- | --- | --- |
| `TRUNCATE_INTERVAL_IN_SECONDS` | `60` | How often the logs DB is trimmed and the jobs are summarised. |
| `LOGS_DB_MAX_ROWS_PER_TABLE` | `100000` | The newest logs, events and errors each logs DB table keeps. |
| `SLOW_QUERY_MS` | `100` | Milliseconds before a DB query (or a transaction that waits for the write lock) is logged as a `Slow query`, with its DB, sqlc name, duration and rows, never its values. `0` logs none. |
| `QUERY_SAMPLE_PERCENT` | `10` | Percent (0 to 100) of DB queries whose time is kept, for the control usage page's slowest queries (longest in total since start). `0` keeps none. |
| `SENTRY_DSN` | Empty | Where errors are reported. Empty turns Sentry off. |
| `MEMORY_ALERT_PERCENT` | `60` | Memory used (of the container's limit, else the machine's) at or above which the control site is warned. |
| `DISK_ALERT_PERCENT` | `60` | The same, for space used on the DBs' disk. |
| `DISK_ALERT_FREE_GB` | `100` | Free GB on the DBs' disk at or above which the disk never warns, however full. |
| `CPU_PEGGED_ALERT_SECONDS` | `20` | How long the CPU is at 100% of every core before the control site is warned. |

## Rate limits

Each is a default. A control admin can change it on the control site's **Rate Limits** page. The control DB's server
settings keep that value, which replaces the env value until set back.

| Variable | Rate / burst | Counts |
| --- | --- | --- |
| `GLOBAL_RATE_LIMIT` / `GLOBAL_RATE_BURST` | `1200000` / `1200000` | Requests to the whole server, to stop a DDoS. The control site has its own limit at the same rate, so sign in stays possible. Neither limits a signed in control admin. |
| `GLOBAL_IP_PER_MINUTE_LIMIT` / `_BURST` | `720` / `1440` | Requests from one IP. Enough for 4 sites. |
| `SITE_LOGIN_PER_MINUTE_LIMIT` / `_BURST` | `1` / `10` | Logins, passkey sign ins, recoveries and setup codes from one IP. Stops credential stuffing. |
| `SITE_LOGIN_NAME_PER_MINUTE_LIMIT` / `_BURST` | `30` / `60` | Logins to one name on a site (or a site's passkey sign ins, recoveries or setup code) from all IPs. Far above one IP's, so one IP cannot lock out others, but guesses over many IPs are capped. |
| `SITE_UNAUTH_PER_MINUTE_LIMIT` / `_BURST` | `60` / `360` | Requests from one IP to a site that fail auth or need none, so bad tokens cannot use the site's limit. Each Socket.IO handshake counts (about 12 for an AAPS phone's namespaces). The burst lets a household's dozen apps reconnect at once after a deploy. |
| `SITE_CREATE_GLOBAL_PER_MINUTE_LIMIT` / `_BURST` | `7` / `20` | Sites created on the whole server. |
| `SITE_CREATE_IP_PER_MINUTE_LIMIT` / `_BURST` | `0.1` / `1` | Sites created from one IP: one each 10 minutes. |
| `SITE_GET_PER_MINUTE_LIMIT` / `_BURST` | `180` / `360` | An identity's authenticated reads. Its site's total can be 4 times this, so one identity cannot use it all. |
| `SITE_POST_PER_MINUTE_LIMIT` / `_BURST` | `60` / `720` | An identity's authenticated writes. Its site's total can be 4 times this, so a phone that uploads a backlog leaves room for a carer's app or AAPSClient. |
| `SITE_POST_FLAGGED_PER_MINUTE_LIMIT` / `_BURST` | `12` / `36` | Writes of a site at abuse level OverUse or above (see `REQUIREMENTS.md`, "Abuse levels"), spent as well as its other write limits. A write costs 1 at OverUse and 4 at Abusive or above. A v1 or live update upload costs that for each record. |
| `SITE_UPDATE_CONNECTORS_PER_MINUTE_LIMIT` / `_BURST` | `1` / `4` | A site's saves of connector settings, to add or change one. |
| `SITE_IDENTITY_CREATE_PER_MINUTE_LIMIT` / `_BURST` | `1` / `5` | A site's new identities and credential resets. Each hashes a password with bcrypt or makes a TOTP secret. |
| `SITE_DOWNLOAD_PER_MINUTE_LIMIT` / `_BURST` | `1` / `3` | A site's data downloads. Each copies its whole DB. |
| `SITE_REPORT_PRINT_PER_MINUTE_LIMIT` / `_BURST` | `5` / `10` | A site's printable reports. Each builds four months of reports. |
| `SITE_PUSH_PER_MINUTE_LIMIT` / `_BURST` | `10` / `30` | A site's pushes to its devices' apps. Each is a request to Apple or Firebase. |

## Socket.IO live updates

| Variable | Default | Meaning |
| --- | --- | --- |
| `SOCKET_PING_INTERVAL_IN_SECONDS` | `25` | How often a client is pinged: Nightscout's interval, which v2 clients expect. |
| `SOCKET_PING_TIMEOUT_IN_SECONDS` | `20` | How long a ping waits for its pong before the client is closed. |
| `SOCKET_MAX_PAYLOAD_BYTES` | `100000` | The largest POST body or websocket message. |
| `SOCKET_MAX_SESSIONS_PER_SITE` | `300` | The most live sessions for one site. |
| `SOCKET_MAX_UNAUTH_SESSIONS_PER_IP` | `30` | The most of a site's live sessions one IP (an IPv6 /64) can hold before they get access, so it cannot take them all from the site's followers. |
| `SOCKET_MAX_BUFFERED_PACKETS` | `1000` | Packets that wait for a client before it is closed as too slow. |
| `SOCKET_MAX_BUFFERED_BYTES` | `2000000` | The same, in bytes. |
| `SOCKET_MAX_PACKETS_PER_SECOND` | `20` | Packets a client can send a second, after a burst of 5 seconds' worth, before it is closed. |

## Scripts

Only the scripts in `cmd/keysets` read these, not the server.

| Variable | Meaning |
| --- | --- |
| `KEY_SETS` | The keys to decrypt with |
