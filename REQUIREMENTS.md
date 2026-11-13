# Requirements

This file lists every feature of Nighttroop. A tester can use each item to check that the feature works.

| Part | What it is | Address |
| ---- | ---------- | ------- |
| **[Control](#control)** | The admin site of the server, for the server operator | `control.<host>` |
| **[Site](#site)** | The site of each person, and its Nightscout API | `<site>.<host>`, `<site>.ns.<host>` |
| **[Identity](#identity)** | The users of a site: how they sign in and what each can do | |

Keep this file up to date. A PR that adds, changes or removes a feature must change this file (see `AGENTS.md`).

## Control

### Server setup

#### Start checks

- The server does not start, and says what to set, when `MODE` is not set, unknown, or `test` outside `go test`.
- `development`: all passwords `password`, all TOTP codes `123123`. Does not start over `https`.
- Production or staging: does not start if `KEY_SETS` holds a public key from `.env.dev` or `.env.test`. Says to make
  new keys with `cmd/keysets/create_keysets.go`.

#### First control admin

- With no control admin, every page goes to `control.<host>/setup`, the APIs reply "not set up", and only the health
  check works.
- Any Admin, of any name, keeps the server set up. Test: add an Admin, delete `admin`, restart. Sites stay online.
- The server writes a one-time setup code to its logs. The setup page asks for it.
  - Wrong codes are rate limited for each IP, and for all IPs at a much higher rate.
  - An error toast shows a wrong setup code or TOTP code.
- The setup page shows the username, password and TOTP secret (with QR code) one time, and saves the admin only after
  a code from that secret.

#### TOTP QR codes

- Every TOTP QR code (setup page, new site admin, new or reset identity):
  - holds an `otpauth://totp` URI, issuer Nighttroop, for the name at its site's or control plane's host;
  - has a hint to scan it, with links (new tab) to Google Authenticator, Microsoft Authenticator and 2FAS.
- New site admin, recovered site, and new or reset identity QR codes also:
  - are a link that opens the authenticator app on a phone;
  - show the secret in groups of four, with a button that copies it without spaces.

#### Staging server

- `MODE=staging` is never in setup, and otherwise runs as production (site admins get random credentials).
  - Its control admin `admin` comes from `STAGING_ADMIN_PASSWORD` and `STAGING_ADMIN_TOTP_SECRET` (required), and
    is made again when its data is wiped.
- A staging deploy wipes every DB and its backups if one DB cannot be migrated and prepared, or has a schema different
  from a new DB's (for example, after a migration's `CHECK` was edited). Otherwise it keeps them.

### Country limits

- With `ALLOWED_COUNTRIES` set, only those countries can create sites, sign in (sites and control) and recover.
  Others get "Nighttroop is not available in your country".
- Not limited: unknown countries, signed in pages, and the Nightscout API.

### Control sign in

#### Password and TOTP

- Sign in with username, password and TOTP code. Sign out.
- Tab order: username, password, code (skips the show password button).
- As on a site, the sixth code digit (typed or pasted) sends the login, if username and password are in. A code typed
  first waits for Log in. A wrong code is not sent again.
- A failed login shows an error toast, "Wrong name, password or code. Codes change every 30 seconds, so use the newest
  one." or too many login attempts. Inputs are invalid only from then to the next send, and have no placeholders.

#### Passkeys

- "Log in with a passkey" signs in the admin whose passkey the browser has, with no name, password or code.
  - Rate limited for each IP, and for all IPs at a much higher rate. Each, failed or not, is a sign in event.
- Account (Security menu) shows the signed in admin, and adds and removes its passkeys as a site's My Account does:
  - only for the control site's own address;
  - at most 10, none the browser already has;
  - named for the admin, with no name, email or device of the person.
- Passkeys of admins and identities share one table. Each is found, listed or removed only for its owner.

#### Global rate limit

- The global rate limit does not apply to control admins. The control site has its own at the same rate. Address and
  sign in attempt limits still apply.

### Control pages

- The header (with menu) and status banner show only after sign in.
- The banner shows when:
  - S3 backups or Sentry are not connected;
  - sites reach 90% of `MAX_SITES`, or all of it;
  - a resource is high (see [Resource alerts](#resource-alerts)).
- Browser tab name: the page's name, for example `Sites · Nighttroop Control`.
- Times in UTC, for example `2026-10-26 12:00:00`, and `-` when not set.

### Control admins

- Admins (Security menu) lists every admin, newest first, with role (Admin or Read only), when made and by whom.
  "You" marks the signed in admin. Never secrets.

| Action | Rules | Result |
| --- | --- | --- |
| Add | Name 1 to 32 characters, unique. Role default Read only. At most 20. | Password and TOTP secret show one time in the Users & Devices credentials dialog, with its "never shown again" warning on Escape or a click beside it. |
| Change role | Role select | Toast. A refused change puts the select back. The admin is signed out. |
| Reset password or TOTP | Manage menu, confirm | New value shows one time in the credentials dialog. Passkeys removed. Credentials changed event, secret not logged. Signed out (old sign ins get 401); the resetting Admin is not. |
| Delete | Manage menu, confirm dialog | Row and passkeys removed. Signed out. |

- Only an Admin can do these, and not to itself. A Read only admin sees only the list.
- The last Admin cannot be deleted or made Read only, checked at the write, so two admins that remove each other at
  once cannot leave none.

### Sites

#### Sites list

- Dashboard, archived sites and site details show state and DB size in MB (two decimal places), newest first.
- Site details also show the DB size limit: 20 MB for a new site.
- Sort by time since last write (first to archive) or DB size.
- Not written to or seen within `DISABLE_AFTER_INACTIVITY_DAYS`: inactive.
- On a phone: no wider than the window. Sort buttons and both cards wrap. The table scrolls sideways.

#### Site creation

- Create a site and Public site creation cards are side by side, each saying what it does.
- An admin creates a site at a selected ID (8 lowercase letters, for example a vanity ID):
  - a taken ID is refused;
  - credentials show one time over the sites list, which has the new site with no reload;
  - allowed past `MAX_SITES` and with public site creation off.
- The ID field accepts only a to z. It lowercases capitals and removes other characters, typed or pasted.
- An admin turns public site creation (from the home page) off and on.

#### Site page

- `/sites/{id}` opens on a click, laid out like a site's summary. Back returns to the sites. An API client gets the
  site.
- Top: ID, state badge, Inactive and Over limit badges; created, batch, and disabled, archived or deleted times; Open
  Site link (new tab) while live.
- DB size:
  - measured on disk when the page opens, kept on the site's record;
  - large, against its limit, with a bar: amber from 80%, red with a warning icon over it;
  - not on disk: from its tables. Neither: not measured yet.
- DB size bar:
  - the five largest tables (with their indexes), darkest first, then Other;
  - a key: name, records in thousands, MB and share (for example, `7.9k · 6.11 MB · 54%`). Other sums its tables'
    records. Hover names a part;
  - table sizes as last measured (at creation, by the 5-minute resource check after a size change, at truncate), not
    read from the whole DB.
- Tiles: last seen, last written to, identities (people and devices), time since last error.
- Identities, and the latest 10 events and 10 errors (no site column). Each opens its details. A link shows all for
  the site.
- On a phone: no wider than the window, all identity actions tappable. Role, TOTP and last seen are a line under the
  name. Actions stack.

#### Site actions

- Actions grid: Open Site, and for an admin the actions below. Each says what it does, and has its button or says when
  it is offered. The page shows the change.

| Action | When | Result |
| --- | --- | --- |
| Download DB | Live or disabled (DB here) | See below. |
| Disable | After a confirmation | Not served. Socket.IO connections close. |
| Archive | After disable. Needs S3. | Uploaded to S3. |
| Hydrate | From the archived sites page | Live again. |
| Truncate | Live or disabled, any size | Truncates all device statuses (as the 10-minute job does the newest), deletes old loop settings no record refers to, shrinks the DB. |

- Download DB (`/sites/{id}/db`):
  - the same zip and code as a site admin's Download Data: `<site>-<date>.zip` holding `<site>-<date>.sqlite`, a
    consistent copy even during writes;
  - control admins only, within the site's download limit;
  - records a "Site DB downloaded by control" event, with the admin, on the site's Events page.
- Permanent delete (for example, of a site never used) deletes its identities (with passwords, TOTP secrets and push
  tokens) and push credentials. Archive keeps them for hydrate.

### A site's identities

- The site page lists identities: who, TOTP or not, last seen. Never secrets. Reset password, Reset TOTP and Sign in
  as line up in columns.
- An admin resets an identity's password (shown one time, with a device's new Nightscout token) or TOTP secret. This
  signs out its sessions and records an event.
- Sign in as, after a confirmation:
  - a new tab opens the site as the identity, with its role, for one hour, via a one-time link (one minute, its site
    only) that sets the sign in cookie;
  - the site's Events page shows a "Signed in by control" event, with the identity and admin;
  - not for Read only control identities.

### Background jobs

- Jobs page: newest first, with duration. A click shows details and errors.
- Job stats page: each day, newest first, with totals, errors, and mean and spread of durations in suitable units.
- A failed job, and each day's stats, link to the logs of failed runs (the day's last 16), filtered by `job_id`.
- Each cron or queued run logs its name and own `job_id`, plus `queued_job_id` if queued and `site_id` for site work.

### Connectors

- Every site's connectors, most recently attempted first. Failed counts red. Turned off after failures: "disabled
  after failing" in amber.
- Counts and durations are whole numbers (12345, never 1.234e+04). Averages (EMAs) have one decimal place.
- A click shows its sync reports.

### Push devices

- `/push-devices`: every site's apps registered for pushes, newest first, at most 500. `?site=` for one site.
- Each: site, identity, app, service, bundle ID, development build or not, registered time, last push and error. Never
  the token.
- Test (`POST /push-devices/{id}/test`), within the site's push limit, shows the new last push or the failure. Read
  only control identities get no button and 403.

### Push credentials

- Push Credentials (`/push-credentials`, menu under Users): each site's credentials and "Every site" ones, newest first.
  - Each: site, name, service, team or project, key ID, apps, bundle IDs, added by.
  - Never the secret, in the page or JSON API (`/push-credentials`, `/push-credentials/{id}`).
- Add Push Credential (`POST /push-credentials`, with `site_id`):
  - live sites only, checked and encrypted as a site admin's, within the site's 5;
  - added by the admin's name, and shown on the site's Push Certificates page;
  - a site not live, or a key that cannot be read, is refused with the modal open.
- For every site (no `site_id`): refused (400).
- Delete (any, after a confirmation) removes it and its secret. Pushes stop using it.
- Read only control identity: no Add Push Credential or Delete buttons. Refused (403) for the add form
  (`/push-credentials/new`), add and delete.

### Logs

#### Events, errors and logs pages

- Events: every site's and the server's, newest first, with readable names, start time and duration. Filter by type,
  search ignores case.
- Errors and logs have their own pages. Narrow all three to one site, request or job run by ID. Details link to
  their request and job.
- Each control sign in, and each refused one, is an event with the IP's estimated country (and city with DB-IP's City
  Lite). The IP is not kept.

#### Log storage

- Every server log is kept in the logs, written in batches, without IP addresses.
- Logs, events and errors keep their `request_id` or `job_id`.
- Logs waiting to be written are capped at 1 MB. Past that, they are dropped and the logs say how many.
- Request logs and Sentry show `REDACTED` for the query's `token`, `secret` and Socket.IO `sid`.

#### Client errors

- A request refused for what it sent (400, 405, 413, 422), from any API or page, is a `client` error with status,
  method, route pattern (never path or query), the response's reason, request ID, identity, role and User-Agent.
- Sentry gets it as a warning, grouped by status, route and reason, with role, User-Agent, site and identity.
- For each site: the same one at most once a minute, at most 20 a minute.
- Not client errors: missing credentials, permissions or records, and rate limits.

#### Request bodies on errors

- A client error or 5xx keeps the first 8 KB of its request body, sent to Sentry as `request_body` with any error
  reported during the request.
- JSON or form only, with credentials and connector logins `REDACTED` (tokens, secrets, passwords, keys, TOTP codes,
  usernames, emails, URLs).
- Other bodies, or one cut at 8 KB (for example, JSON that did not parse): size and content type only.

#### Sentry error reports

- Environment: `DEPLOYMENT` (default `development`). The Fly.io staging deploy sets it to its app name.
- Tags: site, identity, role, request ID, background job and run, connector type, request method and route (never
  path).
- User: `<site>/<id>` or `control/<id>`, named by the identity's name. Transaction: method and route.
- The stack trace ends at the logging code, with no slog frames.

#### Logs DB download

- Download Logs DB (`GET /logs-db`), on the events, errors and logs pages: the whole logs DB as `logs-<time>.sqlite`.
  - A consistent copy while logs are written, made beside the backups in `DATA_DIR`, deleted after it is sent.
  - Not cut off by the 30 second request timeout, at any size.
- Refused: Read only control identity (403, no button), signed out (401), a second download at once (429).

#### Slow queries

- Queries are timed until their rows are read. At `SLOW_QUERY_MS` (100 ms) or more: a `Slow query` log with DB,
  query name, time, rows, request, site and identity.
  Never values.
- A write transaction that waits for the write lock logs as `BEGIN`.

#### Captured requests (staging)

- Staging with `CAPTURE_REQUESTS=true` logs each request and response as `Captured <method> <path>`. Other servers do
  not.
- Each: site, identity, IP address, status, time, headers, and bodies up to 64 KB.
- `REDACTED` in headers, query, path and JSON or form bodies: authorization, cookies, API secrets, tokens, Socket.IO
  session IDs, passwords, keys, TOTP codes, usernames, emails, URLs.
- Other bodies (a page, text, past 64 KB): size and content type only (so no new site credentials).
- Not kept: page styles, scripts, images and fonts; bodies of event streams, Socket.IO connections and logs DB
  downloads.

### Usage

- Refreshes every 10 seconds. Counts only signed in requests. Shows:
  - active sites and users, use since start, server resources, most used routes;
  - sites created today, this week, this month;
  - pushes sent and failed today and this week (from midnight and Monday UTC), in memory since start;
  - memory: the server's total as htop's RES, with SQLite's (not in Go's figures), and the Go heap part;
  - CPU: the last 5 seconds as htop shows it, 100% for each core (up to 400% on 4 cores);
  - the 25 slowest DB queries by total time, of the `QUERY_SAMPLE_PERCENT` (10%) sampled since start: DB, name, runs,
    and mean, 95th percentile, longest and total ms.

### Rate limits

- Refreshes every 10 seconds. Lists each limiter kind: every request, the control site's signed out, an IP's, logins,
  sites created, a site's and an identity's reads and writes, connector changes, identities created, downloads.
- Each kind:
  - what one is for (server, IP, site, identity, login name or IP on a site);
  - rate a minute, burst, and how many are kept;
  - the most used one's spent burst: amber from half, red while any would refuse;
  - how many would refuse now.
- Reading spends nothing and names no site, identity or IP.
- Above them: each limit's env variables and defaults, and its rate and burst now.
  - A control admin can save new values. All limiters of its kind (kept and new) use them immediately and after
    restarts. It shows "set here".
  - The env default, or blank, goes back to the env value.
  - Refused: rate 0 or less, burst under 1.
  - Other control identities only see them.
- A site's reads and writes have 4 times an identity's limit. Test: one identity spends all its own (for example, a
  day's backlog upload). A carer's carbs and AAPSClient's commands still go through.
- Over a site's or identity's limit: 429, with `Retry-After` the seconds until one is accepted, at least 1.

### Resource alerts

- Every 5 minutes the server checks:
  - memory, of the container's limit (else the machine's);
  - space used on the DBs' disk;
  - how long the CPU stayed at 100% of every core, sampled every 5 seconds;
  - each site's DB file and write-ahead log. DB size thus includes sync writes, and is at least its tables' sum.

| Alert | Env variable | Default |
| --- | --- | --- |
| Memory | `MEMORY_ALERT_PERCENT` | 60% or more |
| Disk | `DISK_ALERT_PERCENT` | 60% or more, never with `DISK_ALERT_FREE_GB` (100 GB) or more free |
| CPU | `CPU_PEGGED_ALERT_SECONDS` | 100% for 20 seconds or more |

- The banner shows an alert until a check finds it back under, with memory used and disk free in bytes of their
  totals (for example, "Disk high: 12.0 GB free of 954.0 GB."), and CPU seconds at 100%.
- Events: "Resource high" when over, "Resource recovered" when back under.
- The last 3 days of checks are kept while the server runs, and reset at restart.
- The usage page graphs mean CPU (100% a core), memory and DB disk used, each with a dashed line at its alert level.

## Site

### Site creation and recovery

#### Create a site

- One click on the public home page creates a site, with no account.
- A new site gets its browser's time zone and language, and its country's units and formats.
- Limits:
  - Rate limited for each IP and for the server.
  - Refused when the server holds `MAX_SITES`.
  - If a control admin turns public site creation off, the home page says so in place of the offer, and creation is
    refused.
- A site never used is deleted after `DELETE_AFTER_INACTIVITY_MINS`.

#### Welcome page

- After creation, `/welcome` opens on the new site's address and shows one time:
  - Site URL, site ID, admin password, the phone's Nightscout URL and backup key.
  - A large button to copy all credentials for a password manager, also with the phone's and read only device's
    Nightscout tokens.
- The phone's token is not shown as text, only copied, also by its own button.
- Credentials:
  - Kept only in server memory, for 10 minutes.
  - Found by an ID after the page's `#`, which only the creating (or recovering) browser has. Never in a request
    address or log.
  - The page removes the ID from its address. Opened again (later, or on another site), it says they were shown and
    links to the recover page.

#### Admin login step

- A separate step: username, password, TOTP secret as a QR code and text, and a code check.
- The code check runs at the sixth digit, with no click, and changes nothing. A wrong code says so. A recovered site
  has the same check.
- **Log in**:
  - Disabled until an authenticator code matches. Text under it says to check a code in step 2 first.
  - Opens the login in a new tab, admin name filled in, so the credentials stay open.
- Buttons have the shape of **Copy all credentials**, not round pills.
- **Log in with a passkey**, beside **Log in** on a new or recovered site, with no authenticator app:
  - Adds an admin passkey on the site's address (fingerprint, face or PIN) and signs the admin in.
  - Only for the browser shown the credentials, for 15 minutes, one time. Its cookie sent again adds nothing once the
    admin has a passkey.
  - Later passkeys are added on My Account.

#### Recover page

- The public **recover page** takes a site ID and backup key, then gives the admin a new password and TOTP, gives the
  site a new backup key, and removes the admin's passkeys (a lost device cannot sign in).
- A key works one time. A wrong key is refused with no reason.
- Rate limited for each IP, and for each site from all IPs at a much higher limit.
- It opens `/welcome` on the site's address with the new credentials, as for a new site, less the phone and read only
  device tokens. These do not change, and the recover page never shows them.

### Common page parts

#### Footer and legal notes

- Every footer links (each in a new tab) to contact, terms of service, privacy policy and API docs.
- With `FEEDBACK_URL` set (a Google Form or GitHub issues), the footer links "feedback" and the home page asks for it
  under its features, each in a new tab. Unset, neither shows.
- Pages with estimated locations (events pages) also link DB-IP, whose free databases make the estimates.
- Terms of service and privacy policy name the operator, `OPERATOR_NAME`, link `SUPPORT_EMAIL`, and give the
  operator's governing law. Another operator's server never names T1DNZ or New Zealand.
- Production and staging do not start without `OPERATOR_NAME` and `SUPPORT_EMAIL`. Development and test default to
  T1DNZ and `nighttroop@t1d.nz`.
- Public pages (create, recover, terms, privacy) say above the footer: not a medical device, no medical advice, and
  data and alarms can be late, missing or wrong.

#### API docs

- `/api-docs`, on every host, shows `/openapi.json` with Redoc in the site's colours:
  - Menu and search of all endpoints, grouped Site, Data, Settings and Nightscout, named by their summaries.
  - For each: sign in method, parameters, body, responses and samples.
  - A schema download link. No images from other hosts.
- Each endpoint's host on this server, in the docs and served schema:

  | Endpoints | Host |
  |---|---|
  | `POST /site` | Public (`https://staging.nighttroop.nz`) |
  | Nightscout API | `https://{site_id}.ns.staging.nighttroop.nz` |
  | All others | `https://{site_id}.staging.nighttroop.nz` |

#### Layout

- No scroll on a laptop or phone: create site (with medical note and footer), site login and control login.
- Site login, control login and control setup: a white card centred on a grey page that fills the window, above the
  control site's footer.
- On a phone, all six TOTP boxes fit in their card, and shrink when it is narrower (site and control login).
- `MODE=staging`: a large banner on every page (public, site, control) says it is a test site, frequently wiped.
- Toasts: at the top of every page (site and control), newest on top. A fourth hides the oldest until it times out,
  so three at most take space and the newest always shows.
- Tab title: the page's name (`Glucose Readings · Nighttroop`), after a full load or a link or tab. With no title
  (login): `Nighttroop`.

### Navigation

#### Header and hubs

- The logo and home icon open the summary (`/`). Data and Reports open their hub's first page. No Summary or Settings
  link.
- The shown page's hub is marked, and the home icon on the summary, after a full load or a link or tab.

| Hub | From | Tabs |
|---|---|---|
| Data | Header | Side groups: **Readings & treatments** (Glucose Readings, Insulin Doses, Food Intakes, Temp Targets, Temp Basals, Profile Switches), **Events** (Activities, Notifications, Event Groups), **Devices** (Device Status), **Presets** (Food Presets, Profiles, Loop Settings) |
| Reports | Header | Summaries (first), Overview, AGP, TDD |
| Settings | User menu | Side groups: **This site** (General), **People & devices** (Users & Devices, also another identity's settings), **Push** (Push Devices, Push Certificates) |
| Connectors | User menu | None (`/settings/connectors`) |
| Alarms | User menu | None (`/settings/alarms`) |
| Logs & Events | User menu | A row: Events (first), Errors, Logs |

- The shown page's tab is marked. Each tab keeps its URL, so links, back and forward work.

#### User menu and access

- User menu, with no scroll on a computer: who is signed in, My Account, Getting Started, Settings, Connectors
  (admins), Alarms, Logs & Events, Usage (admins), Logout. Each has its icon and is marked when its page shows.
- Only tabs and items the identity can open show. A non-admin has no Users & Devices or Push Certificates tab, and no
  Connectors, Logs & Events or Usage item.

#### Page search

- The header search finds pages as typed, in any case, by:
  - Name, group or hub.
  - Headings and words on the page, shown beside its name: "Nightscout" or "Dexcom" finds Connectors, "low" Alarms
    (Low Alarm), "timezone" General settings.
  - Other names: "BGL" or "blood sugar" finds Glucose Readings, "bolus" Insulin Doses, "carbs" Food Intakes and Food
    Presets, "exercise" Activities.
- Results show under their hub (Data, Reports, Settings, Connectors, Alarms, Logs & Events, Usage), only pages the
  identity can open, or "No pages match".
- Arrows move, Enter or a click opens, Escape closes. Ctrl+K (⌘K on a Mac) goes to the search from all pages.

#### Phone layout and history

- On a phone: hub links and search are icons, the search opening a box across the page. Reports and Logs & Events tabs
  scroll sideways; Data and Settings tabs wrap above the page. No page gets wider. The user menu is unchanged.
- Links and buttons that replace the page (tabs, pagers, filters) keep its URL, so back returns. Quick clicks show the
  last one clicked.

### Summary

#### Contents

- The home page: latest glucose and its age, the loop's IOB and COB, treatments, and each device's latest status.
- Under the graph, **Insulin & Carbs** lists doses and meals' carbs, newest first, with age, in Insulin and Carbs
  columns. Beside it, **Other Treatments** lists temp targets and temp basals.
- SMBs (a loop's SMBs, Trio's `SMB` doses, Loop's automatic boluses) are marked SMB.
- Links: latest glucose to all readings; each other treatment's kind (Temp Basal) to its page; with no readings in its
  time, to Connectors.

#### Glucose graph

- Glucose, carbs and insulin on one grid. Range buttons: 6 hours (default), 24 hours, 48 hours, a week.
- Axis always 0 to 16 mmol/L (288 mg/dL), marked each quarter. A reading above 16 shows past the top, with no stretch;
  its tooltip gives the real value.
- Drawn:
  - Urgent range with the range shaded across it; low and high dashed.
  - Targets in force shaded over them, edged by darker low and high target lines that step as they change.
  - Readings as dots coloured low, in range or high, joined by a thin line broken at gaps over 15 minutes.
  - Basal in force along the bottom, its highest rate 3 mmol/L up, labelled right. A profile with no timezone has its
    schedules read in the site's, for all viewers.
  - Carbs as labelled up triangles above insulin as labelled down triangles; on overlap the latest label stays. SMBs
    smaller, fainter, unlabelled.
- Legend, as drawn: insulin and SMB (down), carbs (up), basal (outlined area), target (line).
- Pointing shows the time, nearest reading within 5 minutes and its time, basal and targets in force, and treatments
  within 5 minutes.
- Each minute's targets and basal: a temp target or temp basal, else the profile switch, else the default profile in
  force. One from before the summary's hours counts. When a temp or timed switch ends, the one under it returns.

#### Periods and time in range

- Today, yesterday, last week, last month: time in range, total daily dose, last sensor and site change.
  - No data: "–" for time in range and daily dose, not 0. No insulin: 0 U.
- Time in range, on the summary, Summaries page, overviews, AGP reports and readings' glucose days: whole percent,
  "<1%" above 0 that rounds to 0, "–" for none.

#### Glucose colour, age and live updates

- Alarm settings set readings' colour and label: low, in range, high. High text is darker amber than the graph's.
- Latest glucose older than the no data alarm's minutes (default 20) is grey, with an amber badge ("No reading for 45
  min"). An open summary changes when it becomes old, with no data change or reload, reading the summary one time then,
  not every few seconds.
- Ages in words, nearest hour after one hour: "45 minutes ago", "2 days 1 hour ago".
- Tiles and devices show units smaller and grey beside numbers.
- Live updates:
  - New data shows live; the page catches up after the background. "● Live" shows beside the title.
  - Going to another page closes them (the alarms') with no browser console error.
  - Connected up to one hour, not cut by the 30 second request timeout, so an open summary is not read every 30
    seconds.

### Reports

#### Common report rules

- In range is green, not the glucose pages' brand colour: time in range bar, best day, and AGP percentiles and median
  within range.
- Days, weeks and months are the site's: from midnight, dated, listed and linked in its timezone, for all viewers.
- Quick clicks on tabs, periods and arrows show the last one, logging no error for replaced pages.

| Report | Path | Default |
|---|---|---|
| Summaries | `/reports`, `/reports/summaries`; `?month=2026-10` | This month |
| Overview | `/reports/overview`; `/reports/overview/{period}-{date}` | 7 days to yesterday |
| AGP | `/reports/agp`; `/reports/agp/{period}-{date}` | 7 days to yesterday |
| TDD | `/reports/tdd`; `?end=` last day (`2026-01-31`) | 30 days to today, today so far |
| Printable report | `/reports/print` | This month so far and 3 before |

#### Overview and AGP periods

- Any day, week, fortnight or month back one year. Other overview IDs are not found.
  - Day and month: the one with the date.
  - Week and fortnight: the 7 or 14 days to the date, not a summary's calendar week or fortnight. To today, they
    include today so far.
- Day, Week, Fortnight, Month buttons: that kind ending on the report's last day, or today.
- ‹ and › beside the dates (Previous and Next for a screen reader): the report before and after; for a week or
  fortnight, 7 or 14 days, or to today when fewer are over. Disabled past today and one year ago.
- Back and forward move between reports.
- If data in its days before today changed after they were calculated (backfill, late sync, a record moved or
  deleted), a blue notice above says it is being updated. A refresh after the rebuild removes it.
- Under average glucose, GMI, coefficient of variation, time CGM active and time in range: ▲ or ▼ with the amount, or
  "No change", from the previous report (previous 7 or 14 days, day before, previous month). Nothing when either has no
  value (no data, or over one year ago).

#### Overview

- Glucose: average and standard deviation (settings' units), GMI, coefficient of variation. Each "–" with no readings.
- Time in range: stacked bar of very high (above 250 mg/dL) on top, high, in range, low, very low (below 54 mg/dL),
  each percent beside it ("<1%" under one half), target range below.
- Sensor usage: days with data of days so far (6/9), percent of time the CGM read, meter readings a day.
- Cautions (international consensus on CGM metrics):
  - Under GMI, with under 14 days of data: reliable from 14 days.
  - Under time CGM active, under 70%: numbers may not be typical.
- Best day (week, fortnight, month; none on a day's):
  - Whole day with most time in range; on a tie, the later. Only from days the CGM read 70% or more, if any. Not today.
  - A small glucose graph opens that day's overview.
- Daily pattern, days averaged into one from midnight:
  - A bar each 15 minutes, 25th to 75th percentile, parts above range high and below range low coloured, a dot at the
    mean.
  - Range bounds as labelled lines. Hours in the settings' time format.
  - A day's overview uses the 7 days to it.
  - With no glucose, it says so, and that a day's readings show the day after.
- Daily dose and daily carbs, linking the TDD report of the month to its last day. Under them, a bar of basal and
  boluses, each with units and percent.

#### AGP report

- Says which days it covers.
- Time in ranges against type 1 and type 2 goals: a stacked bar with each percent, the target range, and very high and
  very low bounds (settings' units). Goals:
  - Very high <5%; with high, bracketed with their total, <25%.
  - In range >70%.
  - Very low <1%; with low, bracketed with their total, <4%.

- Metrics against goals: average glucose (under 154 mg/dL), GMI (<7%), coefficient of variation (≤36%), time CGM
  active (>70%). Also days with data, daily dose, daily carbs. Each "–" with no days with data.
- AGP over a day from midnight:
  - Shaded 5th–95th and 25th–75th percentiles, median line, high colour above range, low colour below.
  - Very low, range low and high, very high and top marked with values.
  - Percentiles named at line ends, kept apart when close. Hours in the settings' time format.
  - A day's report draws its glucose over the AGP of the 7 days to it.
  - With no glucose, it says so, and that reports use whole days, so a day's readings show the day after. The
    summaries' empty tables say the same.
- Tooltip:
  - On point, tap, or focus and arrow keys (Home and End: first and last).
  - Beside that time, or the nearest with glucose: the time (settings' format) and each percentile, 95% to 5%, in the
    site's units, also marked on the graph.
  - Closes when the pointer leaves or focus goes. A note under the legend explains it.
- Daily glucose profiles: a calendar under weekdays (a month from the site's week start day, a week or fortnight from
  its first day). Each day shows glucose from midnight over the shaded range, coloured above and below, and opens its
  AGP report. A day's report has only its own.

#### Summaries page

- One month: this month, or `?month=2026-10`.
- Tables, longest first: kept months of its year, fortnights and weeks with a day in it, its days.
  - Columns: days with data, time below range (low, very low), time in range, time above range (high, very high), daily
    dose, daily carbs.
  - Newest first, each opening its AGP report.
  - A past month's empty table says none in that month (or year, for months).
- Arrows beside the month: before (to one year ago) and after (to this month). Other months are refused.
- **Printable report** opens `/reports/print` in a new tab.
- Regenerate summaries:
  - Below the tables, not the main action, for summaries that look wrong (after a timezone change).
  - Admins only, after a confirm.
  - A background job rebuilds from site data the last year's days with data, yesterday, and their weeks, fortnights
    and months.
  - Recorded as an event. A repeat before it runs gives one job.

#### Summary builds

- Data for a day before today written (backfill, late sync), changed or deleted rebuilds only that day, the day a
  record moved from, and the whole weeks, fortnights and months with them.
- Counted: glucose, insulin, food, temp basals, temp targets, profile switches, device statuses.
- Starts after 2 minutes with no new site data, or 15 minutes of writes at most, so a backfill gets one build.
- Today is built when read.
- Within one minute of a restart, the server builds:
  - Days marked stale in the site's DB at the write: a record moved away (site or Nightscout API) or permanently
    deleted (Nightscout API).
  - Data found by `updated_at` since the site's last check, `summaries_checked_at`, written just before the stop, to a day built or not.
- A connector's finished backfill: its days are built on the next run, within one minute, with no wait for quiet.
- Data from the last 15 minutes (a phone uploading) does not delay older days. With a looping phone, they build 2
  minutes after their writes stop.

#### TDD report

- Says which days it covers, dated in the site's timezone.
- ‹ and ›: the 30 days before and after. › disabled for the latest month, ‹ past one year ago. A last day after today,
  or over one year before yesterday, is refused. Back and forward work.
- Chart: a bar of each day's total daily dose, oldest first, on an axis of every day, a few pixels apart.
  - Basal (gold) under boluses, with a legend.
  - No insulin: no bar, not counted.
  - Hover: date and basal or boluses in U.
  - Today's bar is paler, "so far" on its right and in its tooltip date.
- Statistics: days before today with insulin, and their mean, lowest and highest daily dose. Today is left out (not
  over). Says so when none have insulin.
- Table under the bars: every day, newest first, basal, boluses and total in U, or "–" with no insulin. Today marked
  "so far", with or without insulin.
- Note under the bars: basal is estimated from the basal rate in force each minute (scheduled or temp basal); the rest
  is boluses. With today in it: today (paler, so far) is left out of the days with insulin and their daily dose, lowest
  and highest.

#### Printable report

- This month so far and 3 months before, newest first; each month's overview and AGP report on its own sheet.
- No site header, links or period buttons.
- Each page, laid out as Dexcom Clarity's:
  - Report name large, over its number of days and first and last day (this month to today).
  - "Generated by Nighttroop" with a small logo, right.
  - The report flat under ruled headings.
  - Foot: "Nighttroop", page number ("1 of 8"), date and time generated in the site's timezone.
- Screen: white A4 sheets on grey, under a **Print** button.
- Printed: one A4 page each, in colour, no grey or **Print** button, no section split.
- 5 a minute for each site, burst 10, then told to wait.
- A page for people: a JSON request is not found.

### Getting started

- No glucose for over a month, or ever: a setup guide replaces the summary, until readings arrive and show live.
- Each step opens its page:
  - Admin: get glucose in (the Nightscout URL for AAPS, Trio or xDrip+, or a Dexcom, LibreLinkUp or Nightscout sync).
  - Admin: add people and devices.
  - All users: settings and alarms.
- Non-admins see that the admin sets up data and users.
- In the user menu ("Getting Started"), and on My Account, which explains and opens it.
- The last step: readings, treatments and presets under Data; reports under Reports; settings, alarms, logs and
  events in the user menu.

### Data

#### Data pages and tables

- A table for each kind, newest first and paged: add, view, edit, delete.
- Add (the main action): right of the title, below it on a phone. Same on presets and push credentials.
- On a phone, a wide table scrolls sideways in its card, not the page. Cells do not wrap (a time, or a value and
  unit, stays together), except notes.
- Add or delete updates the "Showing" count without a reload. Deleting the last record shows the empty table.

#### Confirmation dialogs

- Each delete asks first. All confirmations (site and control) are dialogs in the site's style, not the browser's.
- Question, Cancel, and red Delete (also for the site), red Remove (a passkey) or Continue (others).
- Cancel, Escape, or a click outside does nothing.

#### Paging and record counts

- 50 records a page, with Next while there are more. The last page starts at record 32001, also for one further back.
- A top box, for all pages:
  - The count ("1 temp target", "2 temp targets"). A device's history counts only that device's statuses.
  - Earliest and latest days, without times, with how long ago: "Today", "Yesterday",
    "Sun 8 Nov" (this year), "8 Nov 2025" (earlier).
  - No records: 0, and a dash for each day.
  - Phone: the three figures stay side by side.
- Site API search gives the same:

  | Field | Rule |
  | --- | --- |
  | `items` | the records |
  | `skip` | at most 32000; negative becomes 0 |
  | `limit` | default 50; at most 1000 |
  | `total`, `earliest_at`, `latest_at` | for these data kinds |

#### Data kinds

- **Glucose**: CGM and meter readings with trend arrows, in the settings' units (mmol/L or mg/dL).
  - Lows and highs: the summary's colours, and L or H. Screen readers read "Low" or "High".
  - Add: type (Meter or Sensor), trend as arrow and name (`↗ FortyFiveUp`), as in table and details.
  - Above page one: small graphs of the last 10 days, same scale, oldest first, each with the range shaded and time in
    range. The last is today so far, from clean readings. Title: "Last N days", or "Today" on a new site's first day.
- **Insulin doses**
  - Combo bolus split: the fractions given at once and over time.
  - `automatic`: if a loop gave it. `true`/`false` as Loop uploaded, `null` if the app does not tell. Site API
    returns it, ignores it on create or update.
- **Food intakes**
- **Temp basals**: rate in U/hr, and a multiplier of the scheduled basal (0 for a zero temp).
- **Temp targets**: in the settings' units.
- **Profile switches**: the profile's schedules, and a multiplier of its insulin.
  - Details: basal, carb ratio, sensitivity and target schedule tables, as on profiles, in the settings' units (named
    in headings).
- **Activities**
- **Notifications**
- **Event groups**: see below.
- **Device statuses**: each device's latest 3, and all a device sent. Header units: Phone battery (%),
  Reservoir (U), IOB (U), COB (g).

#### Percent fields

- API: a fraction. Site: a percent.

  | Field | Label | API example | Site range |
  | --- | --- | --- | --- |
  | Combo bolus split | Given at once (%), Extended (%) | 0.7 = 70% | 0% to 100% |
  | Temp basal multiplier | Basal (%) | 0.5 = 50% | 1% to 1100% |
  | Profile switch multiplier | Insulin (%) | 1.2 = 120% | 10% to 500% |

- Both multipliers have the help "A percent above 0: 100% is unchanged, 120% is 1.2 times."

#### Event groups

- Treatments that came in together through the Nightscout API, with their events.
- Add: the Add Event Group menu lists Nightscout types (for example Meal Bolus). Each opens a modal with:
  - A time, and up to one event of each kind the type can have (Meal Bolus: glucose, insulin, food; Exercise: only
    an activity).
  - Events within one hour of the time, updated when the time changes. Times and the window use the time format.
- Change types (Sensor Start, Sensor Stop, Sensor Change, Site Change, Cannula Fill, Prime, Insulin Change, Reservoir
  Change, Pump Battery Change):
  - No events, and no separate page. Listed in the menu.
  - Modal: "no events to pick"; only time, reason, notes, and for sensor start, stop or change a sensor code and
    transmitter ID.
- Refused on add or edit: an event kind the type cannot have, an unknown type, or no events (except change types).
- Click to edit type, time, notes and events. A new type offers only its event kinds. A removed event is in no group.
- An event is in one group at most: grouped events are not offered, and grouping again is refused.
- A group added or edited on the site is a Nightscout treatment, with its events.
- Details: the pump its app named (ID, serial, type), as AAPS sends; a sensor change's sensor code and transmitter
  ID (also in its row's events).

#### Value limits

- All writes (pages, Nightscout API, syncs) have broad limits far outside real use: bolus 500 units, carbs 1000 g,
  dates 1970 to 9999.
- Insulin and food can be negative, to correct too much entered.
- Pages are stricter (a dose: at most 100 units either way).

#### Add modal

- All typeable fields, each with what it is, units, if required, and limits. Types: a dropdown or suggestions.
- Time: in the settings' timezone; blank is now. Other blank fields: not set.

#### Record details form

- Details (data and presets pages) are an edit form with labelled inputs, wrapped on a phone so each can be tapped.
- The add modal's fields, labels, help and limits, filled in, except the time and fields an edit does not change (a
  reading's type and method, a temp target's insulin needs).
- Emptied notes or reason: cleared. A blank number: not changed.
- Save (Rename for a profile): at the foot by Close, as Add is. Enter also saves.
- Each field shows once. Below the form: only what it does not edit (a reading's type, a profile switch's DIA), not
  what it edits (a dose's split, type, notes, reason; a reading's trend).
- Then the same Record section on all pages: ID, UTC offset, app, device, who created and last updated it, and
  when. Unset values do not show.

#### Keyboard and focus

- Table rows (also on control): Tab reaches each, outlined gold. Enter opens details. Enter on delete deletes,
  without them.
- Row delete icons (also in settings): grey, red only on hover or focus.
- Modals:
  - Take the focus on open. Tab goes to fields. Screen readers read the title.
  - Tab after Close goes to the first field; Shift+Tab back. The focus never reaches the page behind.
  - On close (Escape, Close or save), the focus returns to the row or button that opened it.

### Presets

#### Preset pages

- In the Data hub under Presets, with the data they help to enter.
- **Food presets**: add (in a modal of all fields, as for data), view, edit, delete.
- **Profiles**: profile stores a loop uploaded. Read their schedules, rename, delete.

#### Loop Settings

- Each distinct set of settings a loop app sent, kept once:

  | App | Settings | Sent in |
  | --- | --- | --- |
  | Loop | its settings | profiles |
  | Trio | remote settings | profiles |
  | AAPS | configuration | device statuses (an empty one is not a set) |

- Newest first: app, what it came in, device, first and last seen.
- Click: the settings as sent, indented, and changes from the same app's previous set:
  - Each setting changed, added or removed, by name (a preset's by preset name:
    `overridePresets.Running.targetRange`). Old, then new ("not sent" if absent). Glucose in the site's units.
  - An app's first set tells there is no set before it.
  - API: `changes`, `previous_id` (the set before), `units`.
- Read only identities (site API): no Loop or Trio `deviceToken`, in settings or `changes`.
- The Nightscout API reads each profile and device status back with its settings. Deleting a set keeps them; they
  read back without it.

### Connectors

Third party integrations, each syncing from one service: **Dexcom** and **LibreLinkUp** (each with its region),
another **Nightscout** site, or the **virtual connector** (realistic test data).

#### Virtual connector

- Only on development, test or staging servers (`MODE`). In production: not offered, save or enable refused, and one
  saved before never syncs or backfills.
- Writes readings, treatments and device statuses on the 5 minute grid, so overlapping windows write each once.
- Glucose, the same at a given time on all sites:
  - A 6 hour wave (a meal on the rise, a correction at the peak) over a 25 hour wave, so days differ.
  - About 2% very low (below 54 mg/dL), 4% low, 63% in range, 27% high, 4% very high (above 250 mg/dL).
  - Never below 40 mg/dL.
  - A finger prick meter reading at 07:02 and 19:02 UTC, 6 mg/dL above the CGM.
- First sync: the last 30 minutes, always 6 CGM readings and no meter reading. A backfill adds a meter reading.

#### Connector form

- Before a save: only Save. After, without a reload: also Backfill, and Delete (red outline, apart from Save).
- Enable Connector: ticked by typing a new Dexcom, LibreLinkUp or Nightscout source's details. A click on it, before
  or after, is kept. Never changed on a saved source.
- Save does not check details. A failed first sync turns syncing off and tells why.
- A site's connector saves (add or change) share one rate limit (`SITE_UPDATE_CONNECTORS_PER_MINUTE_LIMIT`); over: 429.
- Password, API secret, token, username, email and URL are not offered to browsers or password managers. The password
  is a masked text box (not a password box) with an eye to show it.

#### Sync errors

- A failing source syncs less often, and shows when and why. The reason (stored, logged, Sentry) never has the
  Nightscout URL or host, the Dexcom session, or the followed LibreLinkUp patient's ID, also for a timeout or
  connection error.
- Syncs and backfills skip records they cannot read (a Dexcom or LibreLinkUp reading in an unknown shape) or that the
  site refuses (a Nightscout treatment with an unknown eventType).
- Each sync's first skipped record: a connector error with the record (up to 8 KB); in logs and Sentry, only field
  names and types.

#### Backfill

- A working source can backfill the days before its oldest reading, from midnight in the site's timezone.
- Nightscout and virtual:
  - Also refetch the last 2 days, to fill gaps (as after a deploy), and the toast says so.
  - No device status older than 30 days. Other records: any age.
- Dexcom and LibreLinkUp fetch all they keep: 24 hours and 12 hours.
- Nightscout backfill device statuses drop the loop status (`openaps` or `loop`) and `pump` as sent, keeping IOB, COB,
  and the pump's battery, reservoir, clock, manufacturer, model, `pumpID`, `suspended` and `bolusing`. Existing ones
  are not changed.

#### Connectors page status

- For each source: how and how often it synced; records got in total, from syncs, from backfills, in its last sync
  and in its backfill; recent events; the date its last backfill reached and how long ago ("-" before one).
- After each sync or backfill, new errors, stats and events show without a reload, with "● Live" next to the title.
  A form being edited does not change. Open histories stay open, also ones opened during the update.

#### Encrypted secrets

- Connector credentials are encrypted at rest.
- Bound to their kind and site: connector credentials, TOTP secrets, push secrets and tokens, Nightscout tokens, the
  welcome's passkey cookie. Backups: bound to their object key, without its host.
- Not decrypted: a secret copied in as another site's or kind, a backup as another batch's, or one encrypted unbound.
- Backups copied to a bucket for another host still restore.

#### Nightscout connector

- Copies the newest readings, treatments and device statuses.
- Copies the current profile store when Nightscout profiles change, however long ago it started. An edit in the
  Nightscout profile editor replaces the copy, whoever saved it, whatever its uploader's identifier (for example
  AAPS). A new one is added beside it.
- URL refused on save and on connect, on:
  - This server's host.
  - Private, local or shared (carrier-grade NAT, Tailscale) addresses.
  - Multicast, reserved, broadcast, benchmarking or `0.x.x.x` addresses.
  - IPv4-mapped, NAT64 or 6to4 IPv6 addresses that reach a private or local IPv4 address.
- Refused: this server's host with a trailing dot or capitals (`https://abcdefgh.ns.HOST./`). Trailing dots are
  removed on save.
- A pasted link with a token (`https://site/api/v1?token=...`, `?api_secret=...`, `https://secret@site/api/v1`)
  splits into the URL and the API Secret or Token box.

### Alarms

- Each off by default: urgent low, low, predicted low, high, urgent high, fast rise, fast drop, no data, low device
  battery, low reservoir.
- Site alarms are each identity's default, which it can change.
- An active alarm shows live in a banner below the header on all pages, and can be snoozed.
- Impossible settings are refused (for example a high below the low).

### Settings

#### General settings

- Units, time zone, language, date and time format, week start, number format, insulin and carb defaults.
- Admins set site settings. Identities can override them.
- Identity value, then site value, then fallback:

  | Setting | Fallback | Notes |
  | --- | --- | --- |
  | Language | browser's | English if no translation; only English is written now |
  | Date format | YYYY-MM-DD | |
  | Time format | 24-hour HH:mm:ss | or 12-hour, with or without seconds |
  | Timezone | UTC | |
  | Number format | 1,234.5 | or 1.234,5 or 1 234,5 |

- All pages, including logs, use them. Control and public pages use the browser's language.
- Number format: glucose, insulin, carbs, and summary and report values use it. With 1.234,5, glucose shows 5,4 mmol/L.
- Input fields always use a "." for the decimal point.
- Phone: settings, alarms and another identity's settings are one column (label, then named Site and Yours controls);
  save buttons fit.

#### Delete the site

- Admins can **delete** the site, after a confirmation: archived to S3, everyone signed out.

#### Push credentials

- Settings, Push Certificates (`/settings/push-credentials`). Only admins can add, list, see and delete them.
- An Apple APNs key (team ID, key ID, the `.p8` file pasted in) or a Firebase service account (project ID, its JSON).
- Apps it pushes to, ticked (`loop`, `trio`, `loopfollow`, `t1dcarer`). Optional: the only bundle IDs, comma separated.
- Refused, with the modal left open: an unreadable key, an unknown app.
- Secrets encrypted, never shown or returned. At most 5 a site.
- Delete asks first, then deletes permanently, so repeated adds and deletes do not grow the control DB.

### Logs

#### Events, errors and logs

- Admins see events (for example created, alarms, recoveries, credential changes), errors, logs and batch backups,
  filtered by type and searched.
- Errors include requests refused for what they sent (kind `client`), with why, and the identity and app.
- Each event, error and log shows its request or background job run. Batch events do not.
- A search, or a filter by error kind, reads only the newest 10,000 of each. Paging without one reaches all rows.

#### Error toasts

- Never the server's own words. Server errors and database-worded refusals say:
  - Admins: "Something went wrong. It has been recorded on the Errors page.", with a link there.
  - Others: that it is recorded for the admins.
- The Errors page keeps the full error.
- Refusals written for users show their reason, without the API's "Bad or missing request body" or "Constraint check
  failed" prefix.
- Toasts show above an open modal, so its form's errors can be read and their links clicked.

#### Sign in events

- Each sign in, and each refused one, is an event with the IP address's estimated country, and city if the server has
  DB-IP's City Lite.
- A refused one does not keep the name it tried.

### Nightscout API

#### Hosts and API versions

- Each site has a Nightscout compatible API at `<site>.ns.<host>`, for AAPS, Loop, Trio, xDrip+ and followers.
- The Nightscout host root opens the site's own pages (LoopFollow's Nightscout tab).
- **v1**:
  - entries, treatments, device statuses, profiles, food and activities, with Nightscout's filters.
  - status, versions, verifyauth, slice, count, times, echo, experiments/test, notifications/ack, adminnotifies.
- **v2**: authorization (a JWT for a token, a check of one JWT permission), properties, summary.
- **v3**: create, read, update, patch, delete, search and history of each collection; version, status, lastModified,
  settings.
- **Socket.IO**: live updates and alarms, over long polling or websockets.

#### v1 endpoints

- `echo`: the request's method, URL, query, headers (Authorization, Cookie, api-secret, token as `REDACTED`) and
  how the API reads the search.
- `notifications/ack?level=&group=&time=`:
  - Snoozes the identity's alarms in that group, up to that level, for `time` ms, as the `/alarm` socket's ack does.
  - Answers `OK`. Any identity that can read can ack its own.
- `adminnotifies`: always empty.
- `experiments/test`: `{"status":"ok"}` for an identity that can upload readings, else 401. The Loop and xDrip4iOS
  connection tests pass with a PWD phone's token.
- `alexa` and `googlehome` (Nightscout's Alexa skill and Google Home action):
  - Need an identity that can read.
  - Intents: `NSStatus`, `UploaderBattery`, `MetricNow` (bg, delta, iob, cob, uploader battery).
  - Answers use Nightscout's words, in the settings' units.
  - Alexa launch: a welcome, then Alexa continues to listen.
  - Unknown intent or metric: "not known". Google Home then listens for another.
- `profile/current`: the newest profile set as one object, not a list (Loop), with `enteredBy`. `null` with no
  profile.
- `/pebble` (Pebble and other watch faces):
  - The newest `count` readings of the last two days. Default 1, maximum 1000.
  - Each glucose is a string in the `units` asked for (`mmol` or `mg/dl`), else the settings' units, with its trend
    number.
  - The newest also has the delta, uploader battery and IOB, as in Nightscout.

#### Authentication and permissions

- The `api-secret` header accepts the SHA1 of a device's Nightscout token (xDrip+, with the token in its URL, and
  Loop). After a password reset, the old token's SHA1 does not work.
- v1 status `authorized`: a JWT and the token's permissions, with `api:treatments:create` for a token that can write.
  LoopFollow shows a read token as "Read", a write token as "Read & Write".
- v3 status: what the token can do to each collection (e.g. `crud` for an uploader, `r` for read only). AAPS uploads
  only with a token that can write.
- v3 version: needs no token.

### Nightscout API: searches and deletes

#### Filters

- Unknown field or operator: an empty list. v1 `find[...]` (e.g. `find[notes]=x`), v3 `field$operator` (e.g.
  `notes$eq=x`).
- Other query parameters (e.g. xDrip+'s `rr` or `raw`) are ignored. A document is always returned as stored, not as
  the JSON sent.
- A v1 `find[_id]` that finds nothing (e.g. a deleted treatment): an empty list, not 404.
- v3 `field$eq` is the same as `field` (e.g. `eventType$eq=Site Change`).
- v1 `find[field]` with no operator is its `$eq`, for each field that has one:
  - `find[date]=<ms>`: the entry at that date (xDrip+).
  - `find[created_at]=<ISO time>`: the treatment at that time.
  - A field with no `$eq` (e.g. `find[srvCreated]`): nothing.
- Treatment filters (Trio, iAPS): `find[carbs][$exists]` (also `duration`), `enteredBy`, and `find[enteredBy][$ne]`
  (repeated or in `find[$and][n]`).

#### Date filters

- Milliseconds with a fraction (e.g. `date$gte=1761234567890.25`): the fraction is dropped.
- `created_at` filter and sort: the event time, as `date`. `srvCreated`: the upload time.
- `created_at` with no zone (LoopFollow), or to the minute (e.g. `find[created_at][$gte]=2026-11-06T04:46`): UTC. To
  the minute with a zone: that time in that zone.

#### Sort and paging

- v1 or v3 `sort` on the date (e.g. `sort=date`, `sort=created_at`): oldest first, paged with `limit` and `skip`.
  AAPS pages forward from its newest date and gets every document.
- Other searches: newest first.
- A filter on a field other than the date: a page has the matches of the page the dates alone give, still in order,
  and can have fewer than `limit`.
- `skip`: maximum 32000 (Nightguard reads 32 pages of 1000). Negative is 0.

#### Deletes

- v1 `DELETE /api/v1/entries.json` (xDrip4iOS):
  - Filters: `find[type]`, and `find[date][$gte]`, `find[date][$lte]` or `find[date][$in][]`.
  - Soft deletes the entries found, maximum 1000 each request. Returns the number in `deletedCount`.
  - Other filters, or no date filter: 400.
  - Needs a token that can delete.
- v1 `DELETE /api/v1/treatments.json` or `entries.json` also accepts the Trio and iAPS filters:
  - `find[id][$eq]`.
  - `find[created_at][$eq]` with `find[eventType][$eq]`.
  - `find[dateString][$eq]`.
  - `find[carbs][$exists]`.
  - Up to 10 `find[$or]` alternatives, each with a date or an id.

### Nightscout API: v3 changes and history

- `If-Unmodified-Since` on a replace or patch of a document changed in a later second: 412, no change. The server
  checks in the same transaction as the change.
- A patch of a temp target keeps the fields it does not send:
  - `{"duration": 30}` keeps the targets.
  - `{"targetTop": 9, "units": "mmol"}` keeps the bottom target and duration.
  - `{"duration": 0}` cancels it.
- History:
  - Documents with `srvModified` at or after its time, up to its `limit`.
  - Oldest `srvModified` first, then by date, as in Nightscout.
  - Includes deleted documents with `isValid` false. A search does not.
  - `ETag`: the newest `srvModified`. AAPS pages from the `ETag` and gets every change.
  - In a profile history, the last profile store is the newest changed.
- A delete moves the document's `srvModified` and the v3 lastModified forward, so a history since a sync has later
  deletes.
- A treatment or reading created with `isValid` false (AAPS uploads one it deleted while offline) stays deleted:
  - Read and search do not find it.
  - History has it with `isValid` false.
  - Storage sockets get it as a delete. Followers (e.g. AAPSClient) do not restore it.
- A food or profile created with `isValid` false is stored as usual.

### Nightscout API: v3 settings

- The identity's own settings, units included.
- Each, a default too, has `srvCreated` and `date` in ms.
- Replace or patch: answers with the setting as stored.
- An identifier that is not a Nighttroop setting: 404 for read, replace, patch and delete; 400 for create.
- Units labels (set only through the settings endpoint) choose how the API writes `units`:

  | Setting            | Values             |
  | ------------------ | ------------------ |
  | `units_mgdl_label` | `mg/dL` or `mg/dl` |
  | `units_mmol_label` | `mmol/L` or `mmol` |

  - Default: Nighttroop's label. The API writes only mg/dL.
  - Units are stored Nighttroop's way. All labels are read.
- AAPS documents, kept as sent and shared by all identities of the site:
  - `aaps`, `aaps-state`: the running config.
  - `aaps_clientcontrol_...`: AAPS and AAPSClient pairing and control messages.
  - All operations and `/storage` sockets work as for other documents. The settings' last modified is their last
    change.
- Limits:
  - Document over 256 KiB: refused.
  - New document when the site has 100: refused.
  - Client control messages with no change for a week: deleted to make space.
  - Settings key: up to 128 characters (AAPS client control keys include the client UUID). Other collections'
    identifiers: up to 36.

### Nightscout API: glucose units

- Glucose is stored and always returned in mg/dL, whatever the identity or site units, over HTTP and Socket.IO:
  - `sgv`, `mbg`, treatment `glucose`, temp target `targetTop` and `targetBottom`, profile targets.
  - Entries, treatments, profiles and their `units`.
- Input: in the units sent with it, else guessed from the value.
- A profile, or a profile switch's `profileJson`, in mmol/L comes back in mg/dL in each glucose field:
  - `sens`, `target_low`, `target_high`.
  - Loop's `loopSettings`: `minimumBGGuard`, `preMealTargetRange`, each override preset's `targetRange`, the running
    `scheduleOverride`'s `targetRange`.
- Each treatment comes back with `units`, temp targets and therapy events (e.g. Note, Sensor Change) too.
- v1 status `settings.units`: the identity's units, as Nightscout writes them (`mg/dl` or `mmol`). Followers,
  `/pebble`, `alexa` and `googlehome` use them.

### Nightscout API: dates and times

#### UTC offset

- A document with a date and no `utcOffset` (e.g. an AAPS device status): the site timezone's offset at that date.
- A replace with no `utcOffset`: keeps the replaced document's offset.
- A UTC time with `Z` (Loop) and no `utcOffset`, in a treatment's or device status's `created_at` or a profile's
  `startDate`: the site timezone's offset. A Loop device status: its pump's `secondsFromGMT` (the phone's offset).
- A time with an offset (e.g. `+13:00`), or a sent `utcOffset`: kept.
- Juggluco's `-03:-30` zone in a `dateString`: read as `-03:30`.

#### Time format

- `created_at` as Nightscout writes it (JavaScript `toISOString`): UTC, three fractional digits.
  - `…23.940Z` comes back as sent, not `…23.94Z`. `…23Z` comes back as `…23.000Z`.
- Same format:
  - Treatment `timestamp`, profile `startDate`: always UTC, whatever `utcOffset` (LoopFollow).
  - Entry `dateString`: in its UTC offset. Trio's deleteGlucose finds and deletes an entry by the `.SSS` `dateString`
    it sent.
- Dates kept as sent (e.g. device status `pump.clock`, which LoopFollow reads without fractional seconds): unchanged.
- Entry `date`: kept to the ms, a fraction dropped, also when the `dateString` has whole seconds (Loop). The
  `dateString` gives the UTC offset.

#### Treatment times

- `eventTime` (xDrip4iOS, Nightscout careportal) is the `created_at`.
- A `timestamp` (string or ms) with no `created_at` or `date`: the treatment's time, not its arrival time.
- With a `created_at` or `date`, a sent `timestamp` is not read.

### Nightscout API: identifiers and duplicates

#### Identifiers

- A record made with no `_id` (Nightscout API or site) gets a MongoDB ObjectId shape: 24 lower case hex characters,
  the second it was made, then 8 random bytes. xDrip+ deletes only treatments with such an `_id`.
- An older record's 8 character `_id` still finds it.
- An identifier over 36 characters (e.g. an Omnipod dose's `syncIdentifier`) is kept as a hash. Two with the same
  start stay two records.
- With no `_id`, the identifier is a treatment's or entry's `id` (Trio), or a treatment's `uuid` (xDrip+).
- A create that names its `identifier` keeps it. Only a document with that identifier is its duplicate.

#### Replace

- A v1 upload with an `_id` replaces that document, or creates it with that `_id` (Loop sends a changed override again
  with its `_id`, and deletes it by that `_id`).
- A create with a deleted document's identifier, or a replace of one, replaces it whatever changed (Juggluco deletes a
  treatment and sends it again).
- A v1 treatment replace (`PUT /api/v1/treatments`, or a POST with a stored `_id`) can move its time and `utcOffset`
  (Loop sends an edited carb entry or override with its `_id`).
- A v3 replace, or a v1 replace of another document, refuses a change to the date or `utcOffset`.

#### Duplicates

- Duplicate uploads are ignored.
- A treatment with the date and type of a stored one, or its identifier and date (Loop sends a temp basal or bolus again
  when it ends, Trio a temp target or override run):
  - Merged into it. Its `_id` and the fields not sent are kept.
  - Followers get it as changed. Sent again unchanged: no change.
- Treatment with no app id (`_id`, `identifier`, xDrip+'s `uuid`, Loop's `syncIdentifier`, Trio's `id`): a duplicate
  only if the stored one has no app id and the same carbs and insulin (zero is none), as in Nightscout. Else a new
  treatment:
  - A careportal note in the same second as an xDrip+ bolus: two treatments.
  - AAPS Bolus Wizard's bolus and carbs (two Meal Boluses at one date, deleted one by one): two treatments.
- Entry with the date and type of a stored one, from any device, with no `_id` or an unknown `_id`:
  - The stored entry takes the new value, direction and raw values (xDrip4iOS smooths readings).
  - Followers get it as changed. Same values: no change, last modified too.
  - A reading Trio downloads (Nightscout as its CGM) and uploads with its own id, or xDrip+ sends again: one reading.
- Meter reading (`mbg`) with the date of a stored one, from any device: the same finger prick. A careportal BG Check
  that xDrip+ downloads and uploads again is stored once.

### Nightscout API: fields kept as sent

v1 and v3 return these as sent. For the fields marked \*, an update without them keeps them.

| Document                                 | Fields                                                                                     | Client     |
| ---------------------------------------- | ------------------------------------------------------------------------------------------ | ---------- |
| `mbg` entry                              | `isCalibration`\*, `sensorMgdlAtPairing`\* (mg/dL)                                          | AAPS       |
| Entry                                    | `filtered`, `unfiltered`, 0 too                                                            | AAPS       |
| Treatment                                | `pumpId`, `pumpSerial`, `pumpType`                                                         | AAPS       |
| Bolus                                    | `isBasalInsulin`                                                                           | AAPS       |
| Bolus, profile switch                    | `icfg`: `insulinLabel`, `insulinEndTime` and `insulinPeakTime` (ms), `concentration`       | AAPS       |
| Temp basal                               | `type`\* (e.g. `NORMAL`, `PUMP_SUSPEND`)                                                   | AAPS       |
| `OpenAPS Offline` (running mode)         | `mode`\* (e.g. `DISCONNECTED_PUMP`), `autoForced`\*, `reasons`\*, `originalDuration`\* (ms) | AAPS       |
| Profile switch, effective profile switch | `originalProfileName`\*, `originalCustomizedName`\*, `originalTimeshift`\*, `originalPercentage`\*, `originalDuration`\*, `originalEnd`\* | AAPS |
| Bolus, temp basal                        | `automatic` (true or false)                                                                | Loop       |
| Treatment                                | `syncIdentifier` (up to 256 characters), `insulinType`                                     | Loop       |
| Carb entry                               | `userEnteredAt` (UTC)                                                                      | Loop       |
| Bolus                                    | `unabsorbed`                                                                               | Loop       |
| Treatment                                | `uuid`, `sensorCode`, `transmitterId`                                                      | xDrip+     |
| Treatment, profile                       | `enteredBy` (e.g. `loop://iPhone`, `Loop`, `Trio`), else the uploading identity            | All        |

- An effective profile switch is AAPS's Note with a `profileJson`.
- Other `icfg` fields are left out.
- Loop finds its own treatments by `syncIdentifier`, though an Omnipod dose's `_id` is a hash.
- LoopFollow shows an `automatic` bolus as an SMB, and leaves it out of its bolus stats.
- xDrip+ finds its own treatments by `uuid`. Trio needs a profile's `enteredBy` to import it.
- `bolusCalculatorResult` (AAPS Bolus Wizard calculation, JSON text): kept exactly as sent. Not JSON, or over 4096
  characters: left out, and the rest of the treatment kept.

### Nightscout API: entries

- `sgv` for a CGM reading, `mbg` for a meter reading, never both.
- `direction`: a Nightscout name (e.g. `Flat`, `RATE OUT OF RANGE`), from any letter case or an arrow (`↗`, `↓↓`).
  Unknown: `NOT COMPUTABLE`. Empty: left out.
- `noise`: a whole number.
- Trio reads a fetch of readings as one list, and gets none if one is not valid, including readings stored before.
- Type `cal` (CGM calibration slope and intercept): accepted, not stored. The finger pricks sent with it are kept.

### Nightscout API: treatments

#### Returned fields

- Each treatment has (Loop's NightscoutKit drops treatments without them):
  - `created_at` also as `timestamp`.
  - A bolus's insulin also as `programmed`.
  - A temp basal's `temp` `absolute`.
- `timestamp`: only API v1 and its sockets. API v3 and its storage socket leave it out (AAPS drops a treatment, or a
  page, whose `timestamp` is a date).

#### Notes and values

- `notes`, `reason` and `reasons` of a treatment or activity come back exactly as sent, with HTML, entities and
  surrounding white space (xDrip+ adds a changed note to its own).
  - Only white space: none.
  - Over 1024 characters: cut at a character.
  - The site (pages and API) gets each note, reason and notification without HTML tags and script contents.
- Values that mean none are not stored, and are left out as if not sent: an empty string, a CGM `noise` or `rssi` of
  0, a heart rate of 0. An update with one keeps the stored value.
- `false`, and other zeros (e.g. a combo bolus split), are kept.

#### Limits and errors

- Values far outside real use: refused.
- Insulin or carbs past 500 U or 1000 g, positive or negative: 400 with the field, range and value (e.g.
  `insulin must be between -500 and 500 U (was 8882.5)`). A negative value in range (to take back too much) is kept.
- A value the database refuses: 400 with the column and reason (e.g. `dose is out of range`, `notes is too long`),
  never the database's words. The server logs those.

#### Durations

- `durationInMilliseconds` (AAPS: early ended temp targets, extended carbs, therapy events):
  - Kept on create and update. Returned as sent while within a minute of the stored `duration` (not after the site
    changes it).
  - Sent with no `duration`: the `duration` is the nearest minute.
- A treatment with a duration always has both `duration` and `durationInMilliseconds`, the missing one calculated.
- A `duration` with a fraction (e.g. a Loop bolus `1.6666666666666667`, or a suspend) is kept to the ms.
- API v1 `duration`: not floored. When `durationInMilliseconds` is not whole minutes, it is that value in minutes
  (e.g. an AAPS temp target ended at 359587 ms reads 5.9931). A quick bolus is not 0 minutes. LoopFollow draws a temp
  basal with its full length.
- API v3 and its storage socket `duration`: whole minutes (5), with the exact `durationInMilliseconds` (AAPS drops a
  treatment, or a page, with a fractional `duration`).

#### Temp basals

- Input: the U/hr rate as `rate` or `absolute`, the `percent`, or both.
- `percent`: Nightscout's change from the basal, as AAPS sends it (-50 for half, 20 for 1.2x, -100 for a zero temp,
  never under -100). Stored as a multiplier (0.5, 1.2, 0).
- The value not sent: from the default profile's or profile switch's basal at the start, in the profile's timezone.
  A rate over 11 times that basal keeps no percent.
- Output: always `rate` and `absolute` (both U/hr), and `percent` if it has one.
- Summary: the basal during a temp basal is its rate.

#### Temp targets and overrides

- Loop's `Temporary Override`: a temp target, with its correction range and insulin needs scale factor.
- Indefinite override: lasts a week. Has `durationType` `indefinite` while the override as sent is kept.
- Also indefinite, not ended as it starts: `indefinite` with a `duration` of 0, or a `Temporary Override` of 0
  minutes (the careportal sends one until cancelled).
- A `Temporary Target` of 0 minutes cancels the running one.

#### Profile switches

- No `profileJson` (careportal, T1DCarer): comes back with none. AAPS switches to its own profile of that name.
- `percentage`: a multiplier of the profile's insulin (120 as 1.2), returned as sent.
- `timeshift`: minutes, or ms when sent with an `originalProfileName` (AAPS). Returned as sent.

#### Event types

| Sent                                                                     | Stored as                               |
| ------------------------------------------------------------------------ | --------------------------------------- |
| `eventType` `<none>` (careportal, no type) with insulin and carbs        | Meal bolus                              |
| `eventType` `<none>` with carbs only                                     | Carb correction                         |
| `eventType` `<none>` with insulin only                                   | Correction bolus                        |
| `eventType` `<none>` with no insulin and no carbs                        | Note                                    |
| AAPS `eventType` `SettingsExport` (settings exported)                    | Note                                    |
| `eventType` `Carbs`                                                      | Carb correction                         |
| `glucoseType` with the reading after a colon (e.g. `Finger: 6.7 mmol/L`) | Finger prick                            |
| Loop and Trio `Suspend Pump`, `Resume Pump`                              | Kept                                    |
| Trio and iAPS `SMB`, `External Insulin`                                  | Insulin dose, with its event type       |
| xDrip+ `Sensor Stop`                                                     | Event group, with its event type        |
| Sensor, site or pump part change (e.g. `Sensor Change`, `Site Change`)   | Event group of its type, with no events |

- xDrip+ sends each bolus, carbs and note as `<none>`. Juggluco sends its amounts as `<none>`.
- LoopFollow and NightscoutKit read insulin and carbs only from boluses and carb corrections.
- A treatment sent as `Note` stays a Note, also with carbs or insulin.
- xDrip+ treatments after a `Sensor Stop` are kept. The site can add a Sensor Stop event group.
- v2 properties ages (`sage`, `cage`, `iage`, `bage`) and the summary use change event groups.
- Combo bolus `splitNow` and `splitExt`: fractions now and over time (70 as 0.7), returned as sent.

#### xDrip+ uploads

- A treatment is found by `find[uuid]`, with the `uuid` sent as the `_id` or as xDrip+ makes its `_id`.
- An entry batch (see `docs/clients/xdrip.md`):
  - Calibration (`type` `cal`): left out, its readings kept.
  - Reading out of range (0 or less, or over 1000 mg/dL): left out, the rest stored.
  - Only readings out of range: 200, `[]`, nothing stored. One such reading sent alone: refused.
  - Up to 10 entries with another problem (e.g. unknown `type`): left out, the rest stored.
  - Nothing else to store: refused, with the problem.
- Device status: an uploader battery that is not a number is left out. `uploader.battery` is the uploader battery.

### Nightscout API: device statuses

- Kept:
  - Other pump fields (e.g. AAPS's `extended` with its profile and temp basal).
  - Pump clock, manufacturer, model.
  - Uploader battery, returned as both `uploader.battery` and `uploaderBattery`. With no `uploaderBattery`,
    `uploader.battery` is used.
  - Uploader `isCharging`, `name`, `timestamp`.
  - Loop's override in force.
- While the status as sent is kept: Trio's pump `bolusIncrement` (LoopFollow's remote bolus step),
  `status.bolusing`, `status.suspended`, and battery `string` and `display`.
- After the status as sent is cleared: Loop's pump `pumpID`, `suspended` and `bolusing`, as sent. A follower still
  sees an older status's delivery suspended.
- Out of range, not stored (the rest kept): uploader or pump battery below 0 or over 100% (e.g. Trio's
  `uploader.battery` -100 for an unknown phone battery), reservoir below 0 or over 10000 U.

### Nightscout API: profiles

- v1 search, newest start first: no `count` gives the newest 10 profile sets (as Nightscout), a `count` that number
  (maximum 1000).
- v3 and the socket, with no filters and no limit: the newest.
- Loop's `loopSettings.scheduleOverride` (override running when sent): kept as an override preset (its `duration` in
  seconds). Loop reads it back.
- Trio's remote control settings (bundle, device token, APNS environment, team, expiry) and override presets: kept, for
  LoopFollow's Trio Remote Control.

#### APNs device tokens

- Loop's `loopSettings.deviceToken` and `bundleIdentifier`, or Trio's at the profile's top: register that app's APNs
  token for pushes to the uploading identity.
  - Production, unless `isAPNSProduction` is false.
  - Not a token: not registered. The profile is still kept.
- `deviceToken`: kept and registered whole up to 256 characters (Trio's are 160 hex). Longer: kept empty, not
  registered.
- Device tokens come back (v1, v3, socket storage) only to an identity that can change the site's data (e.g.
  LoopFollow remote commands). Read only users and devices, and the socket's `dataUpdate` to all readers, get none.

### Nightscout API: food and activities

- Food: `type` `food` (AAPS loads only that type).
- Activity (xDrip+): `type` is the activity type, a heart rate's `bpm` its heart rate, a motion's `class` its name.

### Nightscout API: Socket.IO

#### Connection

- A polling client upgrades to a websocket at once: the probe ends its waiting poll. The JavaScript client (AAPS,
  Nightscout pages) does not reconnect each minute.
- No namespace joined in 45 seconds: closed.
- No access in 30 seconds: the socket and its connection close. Access comes from an `authorize` that can read or
  write, or a `/storage` or `/alarm` `subscribe` that succeeds.
- One IP (an IPv6 /64): maximum 10 of a site's live sessions with no access (`SOCKET_MAX_UNAUTH_SESSIONS_PER_IP`). The
  next gets 429.

#### Authorize

- Token that is no one's: disconnected, no answer, as in Nightscout.
- Wrong API secret: the answer says it can do nothing.
- A refused socket's connection closes a second later: at once for a token, after the third refusal for others.
- `authorize` or `subscribe` before access uses its IP's budget on the site (`SITE_UNAUTH_PER_MINUTE_LIMIT`), as a
  failed auth request does, not the site's budget. The site's pages and followers get no 429s from it.

#### Live updates and notifications

- Live updates use the site's own settings (e.g. time zone), glucose in mg/dL.
- To each `/alarm` socket that can read:
  - A new Announcement, its notes as the message (AAPS shows it as a notification).
  - Each other treatment from the last 10 minutes, except Temp Basal and Temporary Override, as a `notification`
    (Nightscout's treatment notify): its event type as the title, who entered it and its notes as the message.

### Nighttroop API and client

#### JSON API and Go client

- Every site page is also a JSON API (with the JWT from sign in).
- The Go client `clients/nighttroop`, generated from the OpenAPI schema, can:
  - Create a site, sign in and read the summary.
  - List, read, add, edit and delete glucose readings, insulin doses and food intakes.
  - Call every Nightscout API v1, v2 and v3 endpoint on the site's Nightscout host, with a device's Nightscout token
    or its JWT.

#### Summary periods

- `GET /data/summary-periods/{period}-{date}` reads the day, week, fortnight or month that contains the date.
  - Dates from today back one year, in the site's time zone and week start day.
  - Other ids and dates out of range: not found.
  - A period not kept is built from its kept days, plus today so far while it is not over. A day not kept is built
    from its data. Neither result is kept.
- Contents:
  - Days with data, time in range, daily dose, carbs, and glucose every 5 minutes.
  - AGP: the 5th, 25th, 50th, 75th and 95th percentiles of the 15 minutes around each 5 minutes from midnight. A day's
    AGP uses the 7 days up to it.
  - `glucose_readings`, and `meter_readings` from a meter.
  - `mean` (in the settings' units), `sd`, `cv` (`sd` as a percent of `mean`) and `gmi` (3.31 + 0.02392 × the mean in
    mg/dL).
  - Percent of readings, total 100: `time_very_low` (below 54 mg/dL), `time_low`, in range, `time_high` and
    `time_very_high` (above 250 mg/dL).
  - `cgm_active`: percent of its 5 minutes, up to now, with a CGM reading.
  - `total_daily_basal`: the basal in force through a day. The rest of the daily dose is boluses.
  - `loop_iob`, `loop_iob_basal`, `loop_iob_treatment` and `loop_cob` every 5 minutes from midnight.
- A week, fortnight or month:
  - Readings and times: all readings of its days.
  - Daily dose, basal and carbs: the mean of its days that are over, or of all its days with data if none is over.
  - Loop IOB and COB: the mean of its days at that time of day.
- A day's loop IOB and COB: the newest loop device report in that 5 minutes, else the 5 minutes before, else null.
  A kept day keeps them after truncation.

### Backups and limits

#### S3

- Without `S3_URL`: no restore on start and no backups.
- With `S3_URL`:
  - The server does not start if no test upload works for about 30 seconds.
  - Production and staging need `S3_ACCESS_KEY` and `S3_SECRET_KEY`, not `minioadmin`, to start.

#### Backups

- Site DBs go to S3 in batches, and the control DB separately. S3 keeps old versions for a retention period.
- A site with no write since its last backup is not copied again, and its DB is not opened.
- A batch with the same sites, none copied again, is not uploaded again until the server restarts.
- In the next round, a write made during a copy is copied, and a failed upload is tried again.
- A batch whose last site is deleted or archived is uploaded empty. A restore then gives an empty batch.
- A batch that never had a site is not uploaded.

#### Restore and start

- A missing control DB or site batch folder is restored from S3 on start. A failed or cut download leaves nothing,
  and the next start tries again.
- A start removes only the staging folders that a stopped restore or archive left beside the DBs and backups.
- A live site whose DB file is missing, but whose batch folder is there:
  - Is restored from its batch's latest S3 backup (one download a batch) and opened.
  - Gives the control admin a "Site DB missing" event: restored, or why not (S3 not working, not in the backup,
    restore failed).
- The start stops, with the site named in the error, if a live site:
  - Cannot be opened or migrated, also after a restore.
  - Has no DB file and was not restored. Its batch is not uploaded without it.
- The server serves nothing until every live site is open and migrated. Else `update.sh` rolls the release back.
- A site DB that cannot be opened keeps its file and backup. A fixed deploy opens it with its data.
- A site creation that fails part way leaves no site or identities.
- A DB (control, site, jobs, logs or live) with a failed migration, when next opened:
  1. Goes back to the version before that migration. A warning names the DB and both versions.
  2. Is migrated again. If this fails, the error names the DB, migration and reason.

#### Stored record data

- A site keeps records only as fields, never the JSON sent. It also keeps:
  - A hash of a treatment's JSON, to tell a changed resend from an unchanged one.
  - Loop's indefinite duration.
  - A device status's pump and uploader, for fields with no column that Nightguard shows (e.g. AAPS's extended).

#### Device status truncation

Every 10 minutes, the server truncates every site:

- Loop status, and pump as sent: kept only on each device's newest status with one. A newer status without one does
  not change this. The others keep:
  - Their IOB and COB, extracted on write. The Nightscout API returns them as `openaps` or `loop`.
  - Their pump's battery, reservoir, clock, manufacturer, model, `pumpID`, `suspended` and `bolusing`. The Nightscout
    API returns them as `pump`.
- Statuses kept for each device: all of the last hour, then the newest in each 10 minutes to 1 week old, each 20
  minutes to 2 weeks, and each hour to 30 days.
  - The device's newest and deleted statuses are kept.
  - All statuses over 30 days old are deleted, also newest and deleted ones.
- The job reads statuses of the last day, and thins again those that became 1 or 2 weeks old in that day.
- A site truncate reads all statuses. It also clears those older than a day (e.g. a device's newest from before it
  stopped reports for a day).

#### Rate limits

- Reads and writes are rate limited for each site, identity and IP. Live sessions are capped for each site.
- A Nightscout request that fails auth, or needs none, uses its IP's budget on the site
  (`SITE_UNAUTH_PER_MINUTE_LIMIT`). One with a valid token does not.
  - An IP with no budget left is refused before its token is checked.
- IP limits (requests, sign ins, recoveries and site creation) cover the whole IPv6 /64.
- A request's IP is its TCP peer.
  - Behind a trusted proxy (`TRUSTED_PROXIES`): the hop before it in `X-Forwarded-For`, or on Fly.io `Fly-Client-IP`.
  - Any other peer that sets these headers keeps its own IP's limits.

#### Server load

- Past `REQUESTS_IN_FLIGHT_PER_CPU` requests at once and a `REQUESTS_BACKLOG_PER_CPU` queue (up to 5 seconds):
  503 at once, with `Retry-After`.
  - A quarter of each is kept for Nightscout API uploads.
  - Never refused this way, known by route only: `/health`, `/socket.io/` and `/<section>/<collection>/events`.
  - Other requests with `Accept: text/event-stream` are still refused when busy, and time out after 30 seconds.
- Each DB serves many reads at once. One slow read does not stop the others. Writes take turns and do not fail.
- A busy DB keeps its connections open between reads. A quiet site's DB closes all of them.

#### Unknown sites

- A site that does not exist, or is disabled, archived or deleted:
  - Pages and API: 404 `{"status": 404, "message": "site does not exist"}`.
  - Nightscout API: 410 Gone `{"status": 410, "message": "site does not exist"}`.
- A site that cannot be looked up (e.g. the server is too busy): 503 with `Retry-After`.

#### Record and size limits

- Any request returns at most 1000 records: site pages and API, Nightscout API and live updates. A Nightscout API
  search of readings (entries) returns up to 2880.
- A v1 upload holds at most 5000 records. Each 100 past the first costs a write against the site's and identity's
  limits.
- A live update upload holds at most 100.
- A Nightscout API delete of a treatment or entry type deletes at most 1000, newest first, and says how many.
- A body over `MAX_REQUEST_BODY_BYTES`: 413. With no length given, reading stops at the limit and it is a bad body.
- Headers over 64 KiB: 431.

#### Security headers

- Every response, on every host: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`,
  `Content-Security-Policy` with `frame-ancestors 'none'`, and `Referrer-Policy: same-origin`.
- With `HOST_PROTOCOL=https`, also `Strict-Transport-Security: max-age=31536000; includeSubDomains`.

## Identity

### Roles

| Role | Signs in to the site | Nightscout API | Changes data | Manages identities and connectors |
| --- | --- | --- | --- | --- |
| Admin | Password and TOTP | No | Yes | Yes |
| User | Password | Yes | Yes | No |
| Read only user | Nightscout token | Yes | No | No |
| Clinician | Password | No | No | No |
| PWD phone (the loop's phone) | Nightscout token | Yes | Yes | No |
| Read and write device | Nightscout token | Yes | Yes | No |
| Read only device | Nightscout token | Yes | No | No |

- A new site has an admin, a PWD phone and a read only device.
- A site has at most `MAX_IDENTITIES_PER_USER` identities.
- An admin adds a clinician as a user with the Clinician role:
  - A password, but no TOTP or Nightscout token.
  - Any page opens the printable report, with Print and Sign out, and no header, menus or tabs.
  - All else is refused (site data, own settings, password, Nightscout API).

### Signing in

#### Password and TOTP

- Admin: password, then TOTP code. Users: password.
- The TOTP step accepts only the pre-auth token from the same site's password step.
- TOTP code fields (sign in, setup, a new site's code check) drop typed or pasted letters.
- A code that is not six digits is refused as wrong.
- A TOTP code signs an identity in once, on the site (page or API) or in control:
  - While it is still valid, it or an earlier code is refused as wrong.
  - Another identity with the same secret can use it.
  - A new site's or the setup page's code check does not use it up.
  - `123123`, the fixed code of dev and test servers, is not one time.

#### Passkeys

- "Log in with a passkey" on the login page signs in the identity whose passkey the browser has, with no name,
  password or TOTP code.
- A passkey works only on its own site's address.
- Stored: only its ID, public key, flags and sign count. Never the authenticator model or attestation.
- A disabled identity's passkey does not sign in.
- Passkey sign ins are rate limited for each IP, and for each site from all IPs at a much higher rate.
- Each passkey sign in or add, on a site or control, finishes once. Its challenge is kept 5 minutes, then removed.
  A repeated finish is refused.

#### Refused sign ins

- Sign in attempts are rate limited for each IP, and for each name from all IPs at a much higher rate
  (`SITE_LOGIN_NAME_PER_MINUTE_*`).
  - Wrong passwords, passkeys or backup keys from one IP use only that IP's attempts.
  - An attempt that one limit refuses uses none of the other.
  - Refused: "Too many login attempts. Wait a minute, then try again."
- A disabled identity cannot sign in.
- Same status and message as a wrong password or secret, and a password check all the same:
  - A site or control login with a name that does not exist.
  - A Nightscout token whose name or role has no identity, or that is for an admin or an identity with TOTP.
- Site login inputs: no placeholders. Marked not valid only after a failed attempt, until the next.
- Devices use their Nightscout token as a token, as an API secret, or for a JWT.

#### End of sessions

- A password or TOTP change signs out the identity's sessions.
- An admin's reset, on the site or in control, also removes its passkeys.
- Disable, delete, or a password or TOTP change also ends the identity's open connections:
  - Socket.IO writes are refused at once, and sockets disconnect within 1 minute.
  - Live update streams end within 30 seconds.
- A page left open after its sign in ends goes to the login page when its live updates reconnect. It does not retry
  them again and again.

### Users & Devices

#### Add users and devices

- An admin adds users and devices. Their credentials show one time, in a dialog:
  - Named for the identity, and takes the focus.
  - Each credential has a copy button. "Copy all" copies all of them.
  - Copy buttons, also in device setup steps, show only an icon, named for what it copies (e.g. "Copy Password") for
    screen readers and as tooltip.
  - Closes only after you confirm the credentials are saved. The confirmation warns they never show again. "Go Back"
    keeps it open.
  - Escape or a click beside it shows that warning, and does not close it. After a reset signed you out, only
    "Sign In Again" remains.
- Add User starts on the User role. The server refuses an identity with no role.
- A select keeps its role when you pick that role again. It does not become empty.
- On a phone, the Add User and Add Device forms wrap in their cards. All inputs and buttons show with no
  sideways scroll.
- Identity creation is rate limited.

#### Device clients

- Clients: AAPS, Trio, Loop, xDrip+, xDrip4iOS, Nightguard, SugarPixel, T1D Carer and Other.
- The client comes before the role. AAPS, Trio or Loop sets and locks the role to PWD Phone. The server refuses
  them with another role.
- Credentials show numbered setup steps for the client: menus and fields in bold, the Nightscout URL and token filled
  in, each with a copy button.
- Any client but Other sets the device's own units to mg/dL. For xDrip4iOS and AAPS, the Nightscout API also writes
  mg/dL as `mg/dl`.
- A device keeps its client (e.g. `loop`, `other`), returned as `client`. Users and a site's first devices have none.

#### Identity list

- A row's switch turns the identity off and on.
- A row's Manage menu has Settings, Reset password, Reset TOTP and Delete, each with its icon. Delete is red and
  apart.
- Actions that do not apply (e.g. to reset or delete yourself) are greyed out. An identity cannot lock itself out.
- Each identity lists the clients (browsers and apps) that used it. It shows as shared when several did.
- From a row, an admin changes the identity's own settings and alarms. A device's units go to followers in its
  Nightscout status.
- Another identity's glucose settings are labelled, shown, followed from the site's and typed in its units (else the
  site's), not the admin's.

#### Credential resets

- An admin resets a device's password. A new Nightscout token comes with it.
- An identity resets its own password on the settings page, the admin its own TOTP. Both sign it out.
- Each reset is an event, with who made it.
- The site's backup key, on the recover page, recovers a lost admin login.

### My Account

- My Account is in the user menu, for identities that sign in to the website.
- It shows the signed in identity and role, a link to the site setup guide (not on control), and the passkeys,
  each with when added and last used.
- "Add a passkey" makes one on the site's own address, with the browser's fingerprint, face or PIN check.
  - Named for the identity (e.g. admin). No name, email or device of the person is asked for or kept.
  - Not added again if the browser already has one for the identity. At most 10.
- "Remove", after a confirmation, removes a passkey. It no longer signs in.

### Usage

- Usage (`/usage`) is in the user menu, for an admin only. My Account and the General settings do not show it.
- As control's page of the site, it shows the DB size (measured on open) against its limit, and a bar and key of
  the largest tables with records, size and share, the rest as Other, as last measured.
- "Truncate", after a confirmation, does as control's Truncate:
  - Clears the loop status and pump as sent from all but each device's newest report.
  - Thins device reports, and deletes those over 30 days old.
  - Deletes loop settings that no profile or device status refers to, except each app's last seen.
  - Shrinks the DB, and keeps all else, as the card says.
  - Shows the storage again, a toast of the new DB size, and a "Site truncated" event with the admin.
- "Download Data" downloads `<site>-<date>.zip`, with a DB snapshot `<site>-<date>.sqlite`.
- Truncate and Download Data share the site's few downloads a minute.

### Own settings and alarms

#### Settings and alarms

- Each identity sets its own units, time zone, language and formats over the site's.
- Each identity has its own alarms. A snooze quiets them only for it.
- `/settings/identity/events` and `/settings/alarms/events` stream only the identity's own changes, and to an admin
  the site's too. To others, a site change comes as `{}`, and
  they read again.
- `/settings/site/events` and `/settings/identity-settings/events` (other identities') stream only to an admin.

#### Push registrations

- Each identity, of any role, registers the apps on its devices for pushes at `/settings/push-registrations` (JSON
  API), with:
  - App: `loop`, `trio`, `loopfollow` or `t1dcarer`.
  - Provider: `apns` or `fcm`.
  - Bundle ID, and if it is a production build.
  - Token: kept encrypted, never returned.
- A token registered again is updated, or moved from another identity of the site. Another site gets its own, and
  cannot take or change this one.
- An APNs token has 64 to 256 hex characters, in upper or lower case.
- At most 10 an identity, also from profiles uploaded at once. Registering one of them again always works.
- An identity sees, streams (`/settings/push-registrations/events`) and deletes only its own. An admin also gets and
  deletes any of the site's, but its search and stream show only its own.

#### Push Devices page

- Settings, Push Devices (`/settings/push-devices`) lists the registered apps: whose, app and service, bundle ID,
  development build or not, when registered, and last push and error. Never the token.
- An admin sees all of the site's. Others see their own, and get 404 for another's.
- "Test" sends a test push, then shows the new last push, or a toast with why it failed.
- "Delete", after a confirmation, removes the row with a "Push Device Deleted" toast, and shows an empty table after
  the last. Pushes stop until its app registers again.

#### Test pushes

- `POST /settings/push-registrations/{id}/test` tests an identity's own registration, or any of the site's for an
  admin.
- It uses the site's push credential for the app, else the server's. With none: refused, saying to add one.
- A token Apple says is gone, or Firebase says is `UNREGISTERED`: the registration is deleted (410), on its own site
  only.
- Another refusal: 502 with Apple's or Firebase's reason. If they cannot be reached: why. Never the device token.
- Each push, and its error, is recorded on the registration. Registering it again (e.g. each Loop or Trio profile
  upload) keeps them.
- A site sends at most `SITE_PUSH_PER_MINUTE_LIMIT` pushes a minute.

#### Loop remote commands

- `POST /api/v2/notifications/loop`, with its `otp`: `Temporary Override`, `Temporary Override Cancel`,
  `Remote Carbs Entry` or `Remote Bolus Entry`.
- Only an identity that can write treatments sends one. Its sender is that identity, not its IP address.
- Pushed as Nightscout does, to the Loop that last uploaded its profile, with the site's push credential for its
  bundle ID, expiring in 5 minutes. Answer: 200 `OK`.
- Remote carbs with a `foodType` (e.g. 🍕) are pushed with it as Loop's `food-type`.
- Refused (400), saying which:
  - A bolus over 30 U.
  - Carbs over 250 g, or absorption over 24 hours.
  - An override over a week, or with no name.
  - Any other event type.
  - A site with no Loop, or no credential for it.

#### FCM pushes

- An FCM push uses the Firebase service account's OAuth token:
  - Got from Google's token endpoint with a JWT its key signs, never from the JSON's `token_uri`.
  - Used again until five minutes before it expires.
- The key and tokens are never logged or shown.
- If Google refuses the service account: 502 with Google's reason.

