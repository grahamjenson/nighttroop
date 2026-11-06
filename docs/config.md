# Configuration

The server is configured by environment variables, read in `utils/config/config.go`. Each is taken from, highest first:

1. a command line flag (only `MODE`, `HOST_NAME`, `HOST_PORT` and `HOST_PROTOCOL` have one, e.g. `-HOST_PORT=8080`)
2. the environment
3. the mode's `.env` file: `.env.dev` in development, `.env.test` in test, `.env` in production
4. the default below

The server will not start with a variable in its `.env` file that it does not know.

Limits written per minute are rates; their burst is how many can be used at once before the rate applies.

## Server

| Variable | Default | Use |
| --- | --- | --- |
| `MODE` | `development`, or `test` under `go test` | `development`, `test`, `production` or `staging`. Picks the `.env` file (`staging` reads `.env`, as production does), and dev and test use cheap bcrypt, fixed site names and fixed keys. `staging` runs as production does, but makes its control admin from `STAGING_ADMIN_PASSWORD` and `STAGING_ADMIN_TOTP_SECRET` instead of being set up, and puts a large banner at the top of every page saying it is a test site whose data is frequently wiped |
| `HOST_NAME` | `nt.local.t1d.nz` | The domain sites are subdomains of, e.g. `abc.nt.local.t1d.nz`, used in links and cookies |
| `HOST_PORT` | `3000` | The port the server listens on. Links to sites, such as the create site page's login, have no port with `HOST_PROTOCOL` `https`, as a proxy takes HTTPS on 443; with `http` they have this port |
| `MAX_REQUEST_BODY_BYTES` | `4000000` | The largest request body the server reads, above an uploader's largest batch, Juggluco's 5000 readings. A larger body is refused with 413 |
| `REQUESTS_IN_FLIGHT_PER_CPU` | `16` | Requests each CPU works on at once. Past these and the backlog a request is refused at once with 503 and `Retry-After`, so a server that falls behind, such as when its CPU is throttled, answers what it can rather than every request timing out. The health check, live updates and Socket.IO are not counted. A quarter of these, and of the backlog, is kept for writes to the Nightscout API, so device uploads are not refused for page loads |
| `REQUESTS_BACKLOG_PER_CPU` | `64` | Requests for each CPU that wait, up to 5 seconds, for one in flight to finish, once `REQUESTS_IN_FLIGHT_PER_CPU` are being worked on. One that waits longer, or arrives with the backlog full, is refused at once with 503 and `Retry-After` |
| `HOST_PROTOCOL` | `https` | `http` or `https`, used in links and cookies. `http` is only allowed in development and test |
| `DATA_DIR` | `./.dbs` | Where the SQLite DBs (`db/`) and backups waiting to upload (`backup/`) are kept |
| `TRUSTED_PROXIES` | `127.0.0.1,::1` | Comma separated IPs or CIDR ranges of reverse proxies, such as Caddy, whose `X-Forwarded-For` is believed for the client's IP |
| `FLY_APP_NAME` | unset | Set by Fly.io. When set, the client's IP is read from the `Fly-Client-IP` header |
| `SUPPORT_EMAIL` | `nighttroop@t1d.nz` | Where users contact the server's operator, in every page's footer and in the terms of service and privacy policy |
| `FEEDBACK_URL` | | Where users give feedback, e.g. a Google Form or GitHub issues; when set, linked from every page's footer and the home page |
| `STAGING_ADMIN_PASSWORD` | empty | Required in staging, and only used there: the control admin's password, at least 16 characters. Keep it a secret, never in a committed file |
| `STAGING_ADMIN_TOTP_SECRET` | empty | Required in staging, and only used there: the control admin's TOTP secret, at least 16 base32 characters, the same in your authenticator app. The admin is made when the server starts without one, so changing these does not change an admin it already has |
| `CAPTURE_REQUESTS` | empty | `true` in staging, and only used there: every request and its response are kept as a `Captured <method> <path>` log in the logs DB, with the site, identity, IP, headers and bodies (each up to 64 KB, as text or base64), to see what clients such as Loop, AAPS and Trio send. Credentials in headers, the query, the path and JSON or form bodies are `REDACTED`, and styles, scripts, images and fonts are not kept. Each log can be up to 256 KB, so lower `LOGS_DB_MAX_ROWS_PER_TABLE` if the logs DB grows too large |
| `SENTRY_DSN` | empty | Where errors are reported. Empty turns Sentry off |
| `DEPLOYMENT` | `development` | The name of this server among those running Nighttroop, e.g. `staging-nz` or `prod-voyager`, which its errors are reported to Sentry under as their environment, and its startup log shows. Unlike `MODE`, which is how the server behaves, each server gets its own |
| `GEOIP_DB_PATH` | empty; the Docker image sets it to its DB-IP IP to City Lite database | A DB-IP IP to City Lite `.mmdb` file, for the city sign ins come from. Without one, or if it cannot be opened, the built in IP to Country Lite gives only the country |
| `ALLOWED_COUNTRIES` | empty, allowing every country | Comma separated ISO 3166-1 alpha-2 codes, e.g. `NZ,AU`, of the only countries that can create sites, sign in and recover. An address whose country is not known is allowed, and signed in sites and the Nightscout API are not limited. Every code DB-IP knows is listed beside `AllowedCountries` in `utils/config/config.go`, to copy and remove from. A code that is not two letters stops the server starting |

## Security

| Variable | Default | Use |
| --- | --- | --- |
| `KEY_SETS` | empty | Required: the server will not start without it, except in test mode, which uses fixed keys. The four keys that encrypt secrets in the DBs and backups, and sign logins. Make them with `scripts/keysets/create_keysets.go`, and keep a copy somewhere safe: backups cannot be read without them |

## S3 backups

| Variable | Default | Use |
| --- | --- | --- |
| `S3_URL` | none | The S3 compatible endpoint backups and archives upload to, e.g. `http://localhost:9000` for the docker-compose MinIO. Unset or empty, S3 is off: there is no restore on start, no backup upload, and sites cannot be deleted. Set, the server tests uploading to it at startup, again over about half a minute, and does not start if it never works |
| `S3_BUCKET` | `mybucket` | The bucket. Versioning and lifecycle rules are turned on for it at startup |
| `S3_ACCESS_KEY` | `minioadmin` | The S3 access key ID |
| `S3_SECRET_KEY` | `minioadmin` | The S3 secret key |
| `CONTROL_BACKUP_INTERVAL_IN_SECONDS` | `30` | How often the control DB is backed up to S3 |
| `BATCH_BACKUP_INTERVAL_IN_SECONDS` | `300` | How often each batch of site DBs is backed up to S3 |
| `CONTROL_BACKUP_RETENTION_DAYS` | `90` | How long S3 keeps a control backup after a newer upload replaces it |
| `SITE_BACKUP_RETENTION_DAYS` | `30` | How long S3 keeps a site batch backup or archive after a newer upload replaces it |

## Sites

| Variable | Default | Use |
| --- | --- | --- |
| `MAX_SITES` | `10000` | The most sites the public can create; creating another from the home page is refused, though a control admin can |
| `MAX_IDENTITIES_PER_USER` | `10` | The most identities one site can have |
| `DELETE_AFTER_INACTIVITY_MINS` | `10080` (a week) | How long a site that was never used is kept before it is deleted |
| `DISABLE_AFTER_INACTIVITY_DAYS` | `7` | How long a site goes unwritten or unseen before control can disable it |
| `TRUNCATE_RECORDS_AFTER_DAYS` | `30` | How old a site's records are before truncating clears their raw JSON |
| `TRUNCATE_INTERVAL_IN_SECONDS` | `60` | How often the logs DB is trimmed and the jobs are summarised |
| `LOGS_DB_MAX_ROWS_PER_TABLE` | `100000` | How many of the newest logs, events and errors each logs DB table keeps |
| `SLOW_QUERY_MS` | `100` | How many milliseconds a DB query (or a transaction waiting for the write lock) runs before it is logged as a `Slow query`, with its DB, its sqlc name, how long it took and how many rows it gave, but never its values; `0` logs none |
| `QUERY_SAMPLE_PERCENT` | `10` | How many in a hundred DB queries have their time kept, from 0 to 100, for the control usage page's slowest queries: those that took the longest in all since the server started; `0` keeps none |
| `MEMORY_ALERT_PERCENT` | `60` | The memory used, of the container's limit or else the machine's, at or over which the control site is warned |
| `DISK_ALERT_PERCENT` | `60` | The space used on the DBs' disk at or over which the control site is warned |
| `DISK_ALERT_FREE_GB` | `100` | The space free on the DBs' disk, in GB, at or over which the disk never warns, however full it is |
| `CPU_PEGGED_ALERT_SECONDS` | `20` | How long the server's CPU is at 100% of every core before the control site is warned |

## Sync

| Variable | Default | Use |
| --- | --- | --- |
| `BACKGROUND_SYNC_WORKERS` | `5` | How many syncs run at once |
| `SYNC_INTERVAL_IN_SECONDS` | `5` | How often the syncs that are due are queued |
| `BACKFILL_DAYS` | `7` | How far back before a site's oldest glucose reading a backfill fetches records |

## Rate limits

Each is a default: a control admin can set any of them on the control site's Rate Limits page, which is kept in the
control DB's server settings and used in place of the env's from then on, until it is set back to the env's.

| Variable | Default | Use |
| --- | --- | --- |
| `GLOBAL_RATE_LIMIT` / `GLOBAL_RATE_BURST` | `100000` / `100000` | Requests to the whole server, to stop a DDoS. The control site has a limit of its own at the same rate, so it can still be signed in to, and a control admin signed in is not limited by either |
| `GLOBAL_IP_PER_MINUTE_LIMIT` / `_BURST` | `720` / `1440` | Requests from one IP, enough to manage 4 sites |
| `SITE_LOGIN_PER_MINUTE_LIMIT` / `_BURST` | `1` / `10` | Logins to a site, to stop credential stuffing |
| `SITE_UNAUTH_PER_MINUTE_LIMIT` / `_BURST` | `60` / `120` | Requests from one IP to a site that fail auth or need none, so bad tokens cannot spend the site's limit |
| `SITE_CREATE_GLOBAL_PER_MINUTE_LIMIT` / `_BURST` | `7` / `20` | Sites created on the whole server |
| `SITE_CREATE_IP_PER_MINUTE_LIMIT` / `_BURST` | `0.1` / `1` | Sites created from one IP, one every 10 minutes |
| `SITE_GET_PER_MINUTE_LIMIT` / `_BURST` | `180` / `360` | A site's authenticated reads |
| `SITE_POST_PER_MINUTE_LIMIT` / `_BURST` | `60` / `240` | A site's authenticated writes |
| `SITE_UPDATE_SYNC_PER_MINUTE_LIMIT` / `_BURST` | `1` / `4` | A site's changes to its sync settings |
| `SITE_IDENTITY_CREATE_PER_MINUTE_LIMIT` / `_BURST` | `1` / `5` | A site's new identities and its credential resets, each of which hashes a password with bcrypt or makes a TOTP secret |
| `SITE_DOWNLOAD_PER_MINUTE_LIMIT` / `_BURST` | `1` / `3` | A site's data downloads, each of which copies its whole DB |
| `SITE_PUSH_PER_MINUTE_LIMIT` / `_BURST` | `10` / `30` | A site's pushes to its devices' apps, each a request to Apple or Firebase |

## Socket.IO live updates

| Variable | Default | Use |
| --- | --- | --- |
| `SOCKET_PING_INTERVAL_IN_SECONDS` | `25` | How often a client is pinged, Nightscout's interval, which v2 clients expect |
| `SOCKET_PING_TIMEOUT_IN_SECONDS` | `20` | How long a ping waits for its pong before the client is closed |
| `SOCKET_MAX_PAYLOAD_BYTES` | `100000` | The largest POST body or websocket message |
| `SOCKET_MAX_SESSIONS_PER_SITE` | `100` | The most live sessions one site may hold |
| `SOCKET_MAX_BUFFERED_PACKETS` | `1000` | Packets waiting for a client before it is closed as too slow |
| `SOCKET_MAX_BUFFERED_BYTES` | `2000000` | Bytes waiting for a client before it is closed as too slow |
| `SOCKET_MAX_PACKETS_PER_SECOND` | `20` | Packets a client may send a second, after a burst of 5 seconds' worth, before it is closed |

## Scripts

These are read only by the scripts in `scripts/keysets`, not the server.

| Variable | Use |
| --- | --- |
| `KEY_SETS` | The keys to decrypt with |
