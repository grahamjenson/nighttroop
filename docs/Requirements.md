# Requirements

Everything Nighttroop does, in three parts:

- **[Control](#control)**: the server's own admin site, at `control.<host>`, for whoever runs the server.
- **[Site](#site)**: each person's site, at `<site>.<host>`, and its Nightscout API, at `<site>.ns.<host>`.
- **[Identity](#identity)**: who can use a site, how they sign in and what each may do.

Keep this up to date: a PR that adds, changes or removes a feature updates this file (see `AGENTS.md`).

## Control

### Setting up a new server

- A new server has no control admin, so every page goes to `control.<host>/setup` and the APIs answer that it is not
  set up, until the admin is made. Only the health check works.
- A server with any Admin, whatever it is named, starts set up: deleting the first admin, `admin`, once another Admin
  is made, and restarting the server does not put it back into setup or take its sites offline.
- The server prints a one-time setup code to its logs; the setup page asks for it, so only whoever runs the server can
  set it up. Wrong codes are rate limited per IP, and from every IP at a far higher rate. A wrong setup or TOTP code
  is said to be wrong in an error toast.
- The setup page shows the control admin's username, password and TOTP secret (with a QR code) once, and saves the
  admin only when a code from that TOTP secret is entered.
- Every TOTP QR code (the setup page, a new site's admin, a new or reset identity) scans into an authenticator app as an
  account: it holds an `otpauth://totp` URI issued by Nighttroop, for the name at its site's or control plane's host.
- Under each TOTP QR code it says to scan it with an authenticator app, linking in a new tab to Google Authenticator,
  Microsoft Authenticator and 2FAS.
- A new site's admin, a recovered site's and a new or reset identity's TOTP QR code is a link, so tapping it on a phone
  opens its authenticator app, and under it the secret is shown as text, in fours, with a button to copy it bare.
- The server will not start without a `MODE`, or with an unknown one, or with `test` outside `go test`, saying what
  to set; a `development` server, whose passwords are all `password` and TOTP `123123`, will not start over `https`.
- A staging server (`MODE=staging`) is never set up: it starts with the control admin `admin` of its
  `STAGING_ADMIN_PASSWORD` and `STAGING_ADMIN_TOTP_SECRET`, made again whenever its data is wiped, and will not start
  without them. Otherwise it runs as production does, its site admins getting their own random credentials.
- A production or staging server will not start with `KEY_SETS` holding any of the public keys from `.env.dev` and
  `.env.test`, saying to make new ones with `cmd/keysets/create_keysets.go`.
- A staging deploy wipes every DB, and their backups, when one cannot be migrated and prepared, or its schema is not
  the one the migrations make on a new DB, as after a migration's `CHECK` was edited, and keeps them otherwise.

### Limiting a server to some countries

- With `ALLOWED_COUNTRIES` set, only addresses in those countries can create sites, sign in (to sites and the control
  site) and recover; others are refused with "Nighttroop is not available in your country". Addresses whose country
  is not known are allowed.
- Signed in pages and the Nightscout API are not limited, so a phone keeps uploading while its person travels.

### Signing in

- A control admin signs in with a username, password and TOTP code, and can sign out. Tab goes from the username to the
  password to the code, skipping the button that shows the password.
- "Log in with a passkey" on the control login signs in the control admin whose passkey the browser has, with no
  name, password or code, as a site's login does. Control passkey sign ins are rate limited per IP, and from every IP
  at a far higher rate, and are sign in events, as a failed one is.
- Account, under Security in the control menu, says which control admin is signed in, and adds and removes its
  passkeys as a site's My Account does: on the control site's own address, so each works only there, at most 10, none
  the browser already has, and named for the admin, with no name, email or device of the person kept. Control admins'
  and site identities' passkeys are kept in one table, each only ever found, listed or removed as its own owner's.
- The control login is sent as soon as the code's sixth digit is typed or pasted, once the username and password are in,
  as the site's is; a code typed before them waits for Log in, and a wrong code is said to be wrong and not sent again.
- A failed control login says, in an error toast, "Wrong name, password or code. Codes change every 30 seconds, so use
  the newest one.", or that there were too many login attempts, and marks the inputs invalid until it is sent again.
  The inputs have no placeholders, and are not marked invalid before then.
- The control site's header, with its menu, and its status banner show only once the admin has signed in, not on the
  login or setup page.
- The server's global rate limit does not stop a control admin signing in, or using the control site once signed in,
  so they can fix whatever is using it up. The control site has its own limit at the same rate, and each address's and
  sign in attempts' limits still apply.
- A banner on every control page says when S3 backups or Sentry error reporting are not connected.
- Errors reported to Sentry are under the server's `DEPLOYMENT` (`development` unless set) as their environment, so each
  of several servers, e.g. staging ones, has its own; a staging server on Fly.io is named after its app.
- An error reported to Sentry can be searched and counted by what it was logged with: tagged with its site, identity,
  role, request ID, background job and run, connector type, and request method and route (never its path, which can carry a
  token); its identity is its user (`<site>/<id>`, or `control/<id>`, named by the identity's name); and its request's
  method and route are its transaction. Its stack trace ends at the code that logged it, without slog's frames.
- The banner also says when the server has 90% or more of the sites `MAX_SITES` allows, and when it has them all.
- Every control page names the browser's tab after itself, e.g. `Sites · Nighttroop Control`.
- Every control page shows its dates and times in UTC, e.g. `2026-10-26 12:00:00`, and `-` for one not set.

### Control admins

- Admins, under Security in the control menu, lists every control admin, newest first, with its role (Admin or Read
  only), when it was made and by whom; the admin signed in is marked "You". Secrets are never shown or returned.
- An Admin adds another admin with a name (1 to 32 characters, unique) and a role, Read only unless chosen, and its
  password and TOTP secret are shown once, in the same credentials dialog, with its "never shown again" warning on
  Escape or pressing beside it, as a site's Users & Devices. Control has at most 20 admins.
- An Admin changes another's role from its Role select, which says so in a toast; a refused change puts it back. The
  admin changed is signed out, as its sign in was for its old role.
- An Admin resets another's password or TOTP from its Manage menu, after confirming, and the new one is shown once in
  the credentials dialog. The reset removes the admin's passkeys and is recorded as a credentials changed event; the
  secret is never logged. The reset signs it out, refusing (401) its sign ins from before; the Admin who reset it stays
  signed in, and it signs in again with its new password or TOTP.
- An Admin deletes another from its Manage menu, after confirming in the confirm dialog, which removes its row and its
  passkeys and signs it out.
- Only an Admin can add, change, reset or delete admins; a Read only admin sees the list only. No admin can change,
  reset or delete themselves here, and the last Admin can never be deleted or made Read only, checked as it is written
  so two admins removing each other at once cannot leave none.

### Sites

- The dashboard lists every site, newest first, with its state and DB size in MB, to two places, as do the archived
  sites and a site's details, with its DB size limit, 20 MB for a new site.
- Sites can be ordered by how long since they were last written to (the first to archive) or by DB size.
- A site not written to or seen within `DISABLE_AFTER_INACTIVITY_DAYS` is marked inactive.
- Clicking a site opens its own page, `/sites/{id}`, laid out like a site's summary, and the browser's back button
  returns to the sites. An API client asking for `/sites/{id}` gets the site. The page shows:
  - its ID, a badge of its state, and Inactive or Over limit badges when they apply, with when it was created, its
    batch, and when it was disabled, archived or deleted;
  - an Open Site link, in a new tab, while it is live;
  - its DB size, measured on disk as the page opens and kept on the site's record, large, against its limit, with a bar
    of how much of the limit it uses: amber from 80%, red over it, with a warning icon. A DB not on disk is measured by
    its tables, and with neither it says it is not measured yet;
  - the bar is split into its five largest tables, indexes counted with their table, darkest first, and the rest as
    Other, with a key under it naming each one with its records in thousands, size in MB and share of the DB (as
    `7.9k · 6.11 MB · 54%`, Other's records its tables' together); hovering a part names it too. Its tables are as
    last measured, kept on the site's record: when it was made, by the resource check every 5 minutes once its DB's
    size changes, and when it is truncated, so opening the page does not read its whole DB;
  - tiles of when it was last seen and last written to, how many identities it has (people and devices), and how
    long since its last error;
  - its identities, and its latest 10 events and 10 errors, each opening its details, with a link to all of them
    filtered to the site; the events and errors leave out the site, as they are all its own.
- On a phone a site's page is no wider than the window, and every identity's actions can be tapped: its role, TOTP
  and when it was last seen are a line under its name rather than columns of their own, and its actions stack.
- A site's page has an Actions grid of what can be done to it: Open Site, and for an admin Download DB, Disable,
  Truncate and Archive. Each says what it does, and has its button when it can be done, or says when it would be
  offered (e.g. Download DB and Truncate while its DB is here, Archive once disabled). Its page shows the change once
  done.
- Download DB (`/sites/{id}/db`) downloads the same zip a site's admin gets from Download Data, made by the same code:
  `<site>-<date>.zip` of a copy of a live or disabled site's DB as it is now, `<site>-<date>.sqlite`, copied
  consistently while the site is written to. Only a control admin can, it shares the
  site's limit on downloads, and the site records a "Site DB downloaded by control" event, with the admin, which its
  admins see on its Events page. An archived site has no DB here to download.
- The sites page has the Create a site and Public site creation cards side by side, each saying what it does.
- On a phone the sites page is no wider than the window: the order buttons, the create form and the public site
  creation card wrap, and the table scrolls sideways in its card.
- An admin can create a site at an ID they pick (8 lowercase letters), e.g. a vanity one; a taken ID is refused. Its
  credentials are shown once, and the sites list behind them is read again, so it has the new site without reloading.
  An admin can create sites past `MAX_SITES` and while public site creation is off.
- The new site ID field only lets letters a to z in: capitals typed or pasted are lowered and anything else is dropped.
- An admin can turn public site creation off, so sites cannot be created from the home page, and back on.
- A site can be:
  - **disabled**, after a confirmation, so it is no longer served and its open Socket.IO connections are closed;
  - **archived** once disabled, uploading it to S3 (needs S3);
  - **hydrated** from the archived sites page, bringing it back live;
  - **truncated** while live or disabled, however small, which truncates all its device statuses as every 10 minutes
    does its newest, deletes its old loop settings no record refers to any more, then shrinks its DB.
- A site deleted for good, such as one never used after it was made, has its identities deleted, with their passwords,
  TOTP secrets and devices' push tokens, and its push credentials deleted; an archived site keeps them, as it
  can be hydrated.

### A site's identities

- A site's page has a table of every one of its identities: who they are, whether they use TOTP, and when each was
  last seen. Secrets are never shown.
- An admin can reset an identity's password (shown once, with a device's new Nightscout token) or TOTP secret. Each
  reset signs out the identity's existing sessions and is recorded as an event. Each identity's Reset password,
  Reset TOTP and Sign in as buttons line up in their own columns.
- An admin can sign in to a site as any of its identities with its Sign in as button, after confirming: a new tab opens
  the site signed in as it, with its role, for the hour a sign in lasts. Control sends the tab to a link to the site
  that works once, within a minute, on its own site only, which sets its sign in cookie. The site records a "Signed in by control" event, with the identity and the control admin, which its
  admins see on its Events page. Read-only control identities cannot.

### Background jobs

- The jobs page lists the background jobs, newest first, with their duration; clicking one shows its details and
  errors.
- The job stats page summarises each day's jobs, newest day first, with its totals and errors, and its jobs' mean and
  spread of durations in the units that suit them, as the jobs page's are.
- A failed job, and each day's job stats, link to the logs of its failed runs (the day's last 16), narrowed to the
  run's `job_id`, to see why it failed.
- Every run of a cron or queued job logs its name and a `job_id` of its own (a queued one also its `queued_job_id`),
  and its work for a site also the `site_id`, so one run's logs can be found together.

### Connectors

- Lists every site's connectors, the most recently attempted first, with failed counts in red.
- A connector turned off after failing says "disabled after failing" in amber.
- Counts and durations are whole numbers, e.g. 12345, never 1.234e+04, and the averages (EMAs) have one decimal place.
- Clicking a connector shows its sync reports.

### Push devices

- Lists every site's apps registered for pushes (`/push-devices`), newest first, at most 500, or one site's with
  `?site=`: its site, identity, app and service, bundle ID, whether it is a development build, when it registered, and
  its last push and error, never its token.
- A control admin can send any of them a test push with its Test button (`POST /push-devices/{id}/test`), within its
  site's push limit, seeing its new last push or why it failed; a read only control identity sees no Test buttons and
  is refused (403).

### Push credentials

- Push Credentials (`/push-credentials`, in the menu under Users) lists every site's push credentials and those for
  every site ("Every site"), newest first: its site, name, service, team or project and key ID, apps, bundle IDs and
  who added it, never its secret, in the page or the JSON API (`/push-credentials`, `/push-credentials/{id}`).
- A control admin can add one to a live site with Add Push Credential (`POST /push-credentials`, with `site_id`),
  checked and kept encrypted as the site's admin's is, within the site's 5, added by the control admin's name; the
  site's admin sees it on their Push Certificates page. A site that is not live, or a key that cannot be read, is
  refused with the modal left open.
- One for every site, with no `site_id`, is refused (400) until the server uses it only for tokens its app registered
  itself, requires its bundle IDs and limits pushes across the server.
- A control admin can delete any of them, a site's or one for every site, after a confirmation, removing its secret
  so pushes stop using it.
- A read only control identity sees them, but no Add Push Credential or Delete buttons, and is refused (403) opening
  the form to add one (`/push-credentials/new`), adding one or deleting one.

### Logs

- Every site's and the server's events, newest first, with readable names, when each started and how long it took.
- Events can be filtered by type and searched (ignoring case); errors and logs have their own pages.
- A request's log, and what Sentry is sent of it, has the `token`, `secret` and Socket.IO `sid` in its query
  `REDACTED`, as a Socket.IO session's polls and posts carry only its `sid` once it is authorized.
- Every log the server writes is kept in the logs too, written in batches, without IP addresses. Logs, events and
  errors keep the `request_id` or `job_id` they were for. Logs held for writing are capped at 1 MB; past that they are
  dropped, and the logs then say how many.
- A request refused for what it sent (a 400, 405, 413 or 422), from any API or page, is an error of kind `client`:
  its status, method and route pattern (never its path or query, which can carry a token), why it was refused as its
  response said, its request ID, and the identity, role and User-Agent that sent it. It is reported to Sentry as a
  warning, grouped by status, route and reason, with the role and User-Agent, and the site and identity as an
  error's are. The same
  one of a site is recorded at most once a minute, and at most 20 of a site's a minute, so a client repeating bad
  requests cannot fill the errors or Sentry. Missing credentials, permissions, records and rate limits are not.
- A request refused for what it sent, or that failed with a 5xx, keeps the first 8 KB of what it sent on its error, and
  sends it to Sentry with any error reported while handling it, as `request_body`: its JSON or form, with credentials
  and connector logins (tokens, secrets, passwords, keys, TOTP codes, usernames, emails and URLs) `REDACTED`, to debug
  it by. A body that is neither, or was cut off at 8 KB (such as JSON that did not parse), keeps only its size and
  content type.
- Logs, events and errors can be narrowed to one site, request or background job run by its ID. An event's, error's
  or log's details link to the rows of its request and job.
- A control admin can download the whole logs DB, as one SQLite file `logs-<time>.sqlite`, with Download Logs DB on
  the events, errors and logs pages (`GET /logs-db`), e.g. to turn staging's captured requests into tests. It is a
  consistent copy while logs are written, made beside the DBs' backups in `DATA_DIR` and deleted once sent, and is not
  cut off by the 30 second request timeout however large it is. A read only control identity sees no
  button and is refused (403), one signed out is refused (401), and a download while another is being sent is
  refused (429).
- Every DB query is timed, until its rows are read, and one that takes `SLOW_QUERY_MS` (100 ms) or longer is logged as
  a `Slow query` with its DB, its query's name, how long it took and how many rows it gave, but never its values, and
  the request, site and identity it was for. A write transaction waiting for the write lock is logged as `BEGIN`.
- A staging server with `CAPTURE_REQUESTS=true` keeps each request and its response as a `Captured <method> <path>`
  log, with its site, identity, IP address, status, how long it took, and its headers and bodies, each body up to
  64 KB. Authorization, cookies, API secrets, tokens, Socket.IO session IDs, passwords, keys, TOTP codes, usernames,
  emails and URLs are `REDACTED` in its headers, query, path and JSON or form bodies. Any other body, such as a page, text or one past
  64 KB, keeps only its size and content type, so a page showing a new site's credentials is not kept. A page's
  styles, scripts, images and fonts are not
  kept, nor the bodies of event streams, Socket.IO connections and downloads of the logs DB. Any other server keeps
  none.
- Each sign in to the control site, and each refused one, is an event with the country (and city, where the server
  has DB-IP's City Lite) its IP address is estimated to be from. The IP address itself is not kept.

### Usage

- Shows active sites and users, use since start, the server's resources and the most used routes, refreshing every
  10 seconds.
- Counts the sites created today, this week and this month.
- Counts the push notifications sent today and this week, since midnight and Monday UTC, and how many of them failed,
  in memory since the server started.
- Its memory is all the server's, as the operating system counts it (htop's RES), with SQLite's, which Go's own figures
  leave out, and says how much of it is Go's heap.
- Its CPU is the server's in the last 5 seconds as htop shows it, 100% for each core used, so up to 400% on 4 cores.
- Only signed in requests are counted.
- Shows the 25 slowest DB queries, by the total time of the `QUERY_SAMPLE_PERCENT` (10%) of queries sampled since the
  server started, each by its DB and name, with how many runs were sampled and their mean, 95th percentile, longest and
  total milliseconds.

### Rate limits

- The rate limits page lists each kind of rate limiter (every request, the control site's signed out, an IP's, logins,
  sites created, a site's and an identity's reads and writes, connector changes, identities created, downloads), refreshing
  every 10 seconds.
- Each shows what one is kept for (server, IP, site, identity, login name or IP on a site), its rate a minute and burst,
  how many are kept, how much of its burst the most used one has spent, and how many would refuse a request now.
- The most used turns amber once it has spent half its burst, and red while any would refuse.
- Reading them spends nothing, and names no site, identity or IP.
- Above them, each rate limit shows its env variables and their default, and its rate and burst now. A control admin can
  change any and save: every limiter of its kind, those kept and those made after, uses it at once, it says "set here",
  and the server uses it again from each start. One set back to its env default, or left blank, uses the env's from
  then on. A rate of 0 or less, or a burst under 1, is refused, and other control identities only see the rates.

### Resource alerts

- Every 5 minutes the server checks its memory (of its container's limit, or else the machine's), the space used on
  its DBs' disk, and how long its CPU stayed at 100% of every core, sampled every 5 seconds.
- The same check measures each site's DB file and write-ahead log on disk, so a site's DB size counts what its syncs
  write as well as what its own requests do, and is never smaller than its tables added together.
- Memory or disk at or over 60% (`MEMORY_ALERT_PERCENT`, `DISK_ALERT_PERCENT`), or the CPU at 100% for 20 seconds or
  more (`CPU_PEGGED_ALERT_SECONDS`), shows in the banner on every control page until a check finds it back under.
- A disk with 100 GB or more free (`DISK_ALERT_FREE_GB`) never warns, however full it is.
- The banner says the memory used and the disk free in bytes, of their totals, e.g. "Disk high: 12.0 GB free of 954.0
  GB.", and the CPU's seconds at 100%.
- Going over is a "Resource high" event of the server's, and coming back under a "Resource recovered" one.
- The last 3 days of checks are kept while the server runs, and start again when it restarts. The usage page graphs
  them: mean CPU, up to 100% for each core as htop shows it, and memory and DB disk used, with a dashed line at the
  level each warns at.

## Site

### Creating and recovering a site

- Anyone can create a site in one click on the public home page, without an account. It opens the new site's own
  address, `/welcome`, which shows its credentials once: its URL, site ID, admin password, the phone's Nightscout URL,
  and the backup key, with a prominent button to copy them all for a password manager. The copy has every first
  identity's credentials: the admin's, and the phone's and the read only device's Nightscout tokens. The phone's token
  is not shown as text, only in the copy and by its own copy button.
- The credentials wait on the server, only in memory, for 10 minutes, under an ID after the welcome page's `#` that only
  the browser that created (or recovered) the site has, so it is never in a request's address or a log. The page drops
  the ID from its address, and opening it again, after that or on another site, says they have been shown and links to
  the recover page.
- The new site's admin login is its own step: its username and password, its TOTP secret as a QR code and text, and a
  code check. The Log in button, which opens the login in a new tab with the admin's name filled in so the credentials stay open, is disabled until a code from the authenticator app matches, saying under it to check a code in step 2 first; a wrong code says so. Its buttons are the shape of its Copy all credentials, not round pills. A
  code is checked as soon as its sixth digit is in, with no click. The check changes nothing, and a recovered site's
  credentials are checked the same way.
- Beside Log in, a new or recovered site's welcome page offers "Log in with a passkey", with no authenticator app: it
  adds the admin a passkey on the site's own address, with the device's fingerprint, face or PIN, and signs the admin in with it. Only
  the browser the credentials were shown to can, for 15 minutes, once: sending its cookie again adds nothing once the admin
  has a passkey. After that the passkey is added on My Account.
- A new site takes the time zone and language of the browser that created it, and its country's units and formats.
- Site creation is rate limited per IP and server wide, and refused once the server holds `MAX_SITES`.
- While a control admin has turned public site creation off, the home page says so instead of offering to create a
  site, and creating one is refused.
- The public **recover page** takes a site's ID and backup key and gives the admin a new password and TOTP and the
  site a new backup key, and removes the admin's passkeys, so a lost device's passkey signs in no more. The key works once; a wrong key is refused without saying why. Attempts are rate limited per IP, and a site's from every IP at a far higher rate.
- A recovery opens the site's own address, `/welcome`, which shows the new credentials once, as a new site's are, less
  the phone's and the read only device's tokens, which are left as they were; the recover page never shows them.
- A site that is never used is deleted after `DELETE_AFTER_INACTIVITY_MINS`.
- Every page's footer links to contact, the terms of service, the privacy policy and the API docs, and each footer link
  opens in a new tab, so the page the user was on stays open.
- The terms of service and privacy policy name the server's operator, `OPERATOR_NAME`, and link its `SUPPORT_EMAIL`,
  with the governing law where the operator is based; a server run by someone else never names T1DNZ or New Zealand.
- A production or staging server will not start without `OPERATOR_NAME` and `SUPPORT_EMAIL`; development and test
  default to T1DNZ and `nighttroop@t1d.nz`.
- The API docs, at `/api-docs` on every host, show the OpenAPI schema, `/openapi.json`, with Redoc in the site's
  colours: a menu and search of every endpoint, grouped into Site, Data, Settings and Nightscout and each named by its
  summary, each with how it is signed in to, its parameters, body, responses and samples of them, and a link to
  download the schema. The page loads no images from anywhere else.
- Each endpoint in the API docs, and in the schema served, gives the host it is on, on this server: `POST /site` the
  public host (e.g. `https://staging.nighttroop.nz`), the Nightscout API a site's Nightscout host
  (`https://{site_id}.ns.staging.nighttroop.nz`), and the rest a site's host (`https://{site_id}.staging.nighttroop.nz`).
- With `FEEDBACK_URL` set, e.g. to a Google Form or GitHub issues, every page's footer also links to "feedback", and the
  home page asks for it under its features, each opening the link in a new tab; unset, neither shows.
- The public pages (create, recover, terms and privacy) say above their footer that Nighttroop is not a medical device,
  gives no medical advice, and its data and alarms may be late, missing or wrong.
- The create site page, its medical note and footer included, and the site and control login pages fit a laptop's or
  a phone's window, so none of them needs scrolling.
- The site login, control login and control setup pages each show their white card in the middle of a grey page that
  fills the window, above the control site's footer.
- A TOTP code's six boxes fit inside their card on a phone, the last one included: they shrink to fit a card narrower
  than them, as the site and control login cards are on a phone.
- A staging server (`MODE=staging`) says it is a test site in a large banner at the top of every page, public, site and
  control, warning that its data is frequently wiped.
- Toasts show at the top of every page, site and control, newest on top. At most three take space at once: a fourth
  hides the oldest until it times out, so the newest is always seen and the stack never grows further over the page.
- Every page names the browser's tab after itself, e.g. `Glucose Readings · Nighttroop`, whether it is loaded in full or
  opened from a link or tab; a page with no title of its own, such as the login, is named `Nighttroop`.
- A page that shows estimated locations, such as the events pages, also links to DB-IP, whose free databases make
  the estimates, in its footer.

### Navigation

- The header has the logo and a home icon, both opening the summary (`/`), and links to the Data and Reports hubs, each
  opening its hub's first page; there is no Summary or Settings link. The hub of the page shown is marked, after a page
  opened from a link or tab as well as one loaded in full, and the home icon on the summary.
- Data has side tabs in groups, like Settings: Readings & treatments (Glucose Readings, Insulin Doses, Food Intakes,
  Temp Targets, Temp Basals, Profile Switches), Events (Activities, Notifications, Event Groups),
  Devices (Device Status) and Presets (Food Presets, Profiles, Loop Settings).
- Reports has a tab for each report: Summaries, Overview, AGP and TDD, and opens on Summaries.
- Settings, from the user menu, has side tabs in groups: This site (General), People & devices (Users & Devices) and
  Push (Push Devices, Push Certificates). Another identity's settings count as Users & Devices.
- Connectors, from the user menu, is a page of its own (`/settings/connectors`), with no tabs.
- Alarms, from the user menu, is a page of its own (`/settings/alarms`), with no tabs.
- Logs & Events, from the user menu, has a row of tabs, Events, Errors and Logs, and opens on Events.
- A hub's tab for the page shown is marked; each tab keeps its page's own URL, so links, back and forward still work.
- Only the tabs and menu items an identity may open are shown: someone who is not an admin has no Users & Devices
  or Push Certificates tab, and no Connectors, Logs & Events or Usage in the user menu.
- The header's search finds a page by its name, group or hub as it is typed, whatever the case, listing them under
  their hub (Data, Reports, Settings, Connectors, Alarms, Logs & Events, or Usage), only those the identity may open, or "No pages match".
- The search also finds a page by the headings and words on it, showing the one it was found by next to the page's
  name: Nightscout or Dexcom finds Connectors, low finds Alarms (Low Alarm), timezone finds General settings.
- The search also finds a page by what people also call it: BGL or blood sugar finds Glucose Readings, bolus finds
  Insulin Doses, carbs finds Food Intakes and Food Presets, exercise finds Activities.
- In the search, the arrow keys move through the results, Enter or a click opens one, Escape closes them, and Ctrl+K
  (⌘K on a Mac) goes to the search from anywhere.
- On a phone the header fits, its hubs' links as icons and its search an icon opening a box across the page; Reports'
  and Logs & Events' tabs scroll sideways on their own, and Data's and Settings' tabs wrap above the page, so no page
  gets wider. The user menu is the same as on a computer.
- The user menu has who is signed in, My Account, Getting Started, then Settings, Connectors (for admins), Alarms, Logs
  & Events, and Usage (for admins), each with its icon and marked when its page is shown, and Logout, all shown without
  scrolling on a computer.
- Every link and button opening a page in place of this one, as the tabs, pagers and filters do, keeps its URL, so back
  returns, and clicking several quickly shows the last one clicked.

### Summary

- The home page summarises the site: the latest glucose with how long ago it was, the loop's IOB and COB, the
  treatments, and each device's latest status.
- Under the graph, an Insulin & Carbs table lists each dose and meal's carbs, newest first, with how long ago it was,
  in an Insulin and a Carbs column; a loop's SMBs, Trio's `SMB` doses and Loop's automatic boluses, are marked
  SMB. Beside it, Other Treatments lists the temp targets and temp basals.
- A graph of glucose with carbs and insulin marked, over 6 hours to a week, drawn on one grid as Carer draws it. A
  site opens on the last 6 hours, and its range buttons choose 6 hours, 24 hours, 48 hours or a week:
  - The glucose axis is always 0 to 16 mmol/L (288 mg/dL), marked at each quarter, whatever the readings: a reading
    above 16 is drawn past the axis's top, not stretching it, and its tooltip gives the real reading.
  - The urgent range and the range shaded across it, with the low and high as dashed lines.
  - The targets in force shaded over them, edged by darker lines of the low and high target, stepping as they change.
  - Readings as dots, coloured low, in range or high, joined by a thin line broken where readings are over 15
    minutes apart.
  - The basal in force along the bottom, its highest rate reaching 3 mmol/L up the glucose axis and labelled on the
    right. A profile sent without its own timezone has its schedules read in the site's timezone, whoever views it.
  - Carbs as triangles pointing up above insulin as triangles pointing down, each labelled, the latest kept where
    labels overlap; a loop's SMBs, Trio's `SMB` doses and Loop's automatic boluses, are smaller, fainter and not
    labelled.
  - A legend of insulin and SMB (triangles pointing down), carbs (pointing up), basal (an outlined area) and target
    (a line), each drawn as on the graph.
  - Pointing anywhere on the graph shows what was then: the time, the reading nearest it within 5 minutes and when
    it was, the basal and the targets in force, and any treatment within 5 minutes.
- Each minute's targets and basal are a temp target's or temp basal's while one is in force, otherwise the profile
  switch's in force, otherwise the default profile's in force. A default profile or switch from before the summary's
  hours still counts, and once a temp or a timed switch ends the one under it shows again.
- Time in range, total daily dose, and the last sensor and site change for today, yesterday, last week and last
  month. A period with no data shows "–" for its time in range and daily dose, not 0; one with data and no insulin
  shows 0 U.
- A time in range reads the same on the summary, the Summaries page, overviews, AGP reports and the readings' glucose
  days: a whole percent, "<1%" for one above 0 that rounds to 0, and "–" for none.
- Readings are coloured and labelled low, in range or high by the alarm settings; high text is a darker amber than
  the graph's, so it can be read.
- A latest glucose older than the no data alarm's minutes (20 by default) is grey, with an amber badge saying how
  long there has been no reading, as "No reading for 45 min".
- An open summary greys its latest glucose and shows that badge when it turns old, without a change or a reload of
  the page, reading the summary again once then rather than every few seconds.
- The summary says how long ago things were in words, to the nearest hour once over an hour, as "45 minutes ago" or
  "2 days 1 hour ago", and its tiles and devices show units smaller and grey beside their numbers.
- The summary's latest glucose links to the page of every reading, and each other treatment's kind (e.g. Temp Basal)
  to the page of that kind.
- The summary updates live as new data arrives, and catches up after the page has been in the background. While it
  does, ● Live shows beside its title.
- Moving to another page closes the page's live updates, such as the alarms', without an error in the browser's
  console.
- A page's live updates stay connected for up to an hour, not cut off by the 30 second request timeout, so an open
  summary is not read again every 30 seconds.
- A summary with no readings in its time links to the Connectors page.

### Reports

- The reports draw what is in range in green, as other diabetes reports do, not the brand's colour the glucose pages
  keep for it: the time in range bar, the best day, and the AGP's percentiles and median within the range.
- Every report's days, weeks and months are the site's: from midnight in the site's timezone, and dated, listed and
  linked in it, whatever the timezone of whoever views them.
- A report's week and fortnight, on the Overview and AGP pages, are the 7 and 14 days to its date, as Dexcom Clarity's
  are, not the calendar week or fortnight a summary period is, which may have only just begun; a day is the day and a
  month the month with the date in it. A week or fortnight to today has today so far in it.
- The Reports hub has Overview, at `/reports/overview`, which shows the 7 days to yesterday at first, and the
  overview of any day, week, fortnight or month from today back a year is at `/reports/overview/{period}-{date}`, with
  the same buttons to other periods as the AGP report. Any other id is not found.
- The Overview and AGP say, in a blue notice above the report, that it is being updated while data in its days before
  today changed since they were worked out, as a backfill or late sync writes, or a record moves or is deleted, until
  they are built again; refreshed once they are, it is gone.
- An overview shows its glucose's average and standard deviation, in the settings' units, and its GMI and coefficient
  of variation, each "–" with no readings.
- An overview's time in range is a bar of its very high (above 250 mg/dL), high, in range, low and very low (below
  54 mg/dL) readings stacked, very high at the top, each beside it with its percent ("<1%" for one under a half), and
  the target range below it.
- An overview's sensor usage is its days with data of its days so far (as 6/9), the percent of its time the CGM read,
  and its meter readings a day.
- An overview and an AGP report say under their average glucose, GMI, coefficient of variation, time CGM active and
  time in range how each rose or fell (▲ or ▼, by how much, or No change) on the report before: the previous 7 or 14
  days, the day before or the previous month. They say nothing when either has none, as when it has no data or is
  over a year ago.
- As the international consensus on CGM metrics has it, an overview cautions under its GMI that it is reliable from 14
  days of data when it has fewer, and under its time CGM active that its numbers may not be typical when the CGM read
  for under 70% of the time.
- A week's, fortnight's or month's overview names its best whole day, the one with the most time in range (the later
  of two the same, and of the days the CGM read for at least 70% of when there are any, so a part day of a few hours
  is not taken for it), with a small graph of its glucose through the day, which opens that day's overview; today is left out, and a day's overview has none.
- An overview draws its days averaged into one from midnight: a bar of each 15 minutes from its 25th to its 75th
  percentile, the part above the range high and below the range low in their colours, a dot at each's mean, and the
  range's bounds as lines labelled with their values, with hours in the settings' time format. A day's is of the 7
  days to it. One with no glucose says so, and that a day's readings show the day after.
- An overview shows its daily dose and daily carbs, with a link to the TDD report of the month to its last day, and
  under them how much of the daily dose is basal and how much boluses, as a bar of the two and each's units and
  percent.
- The Reports hub has AGP, the AGP report, at `/reports/agp`, which shows the 7 days to yesterday at first, and a
  report of any day, week, fortnight or month from today back a year is at `/reports/agp/{period}-{date}`.
- The Reports hub's Summaries page, at `/reports` (and `/reports/summaries`), is of a month, this one unless it is
  given as `?month=2026-10`, and has a table of each, the longest first: the kept months of its year, the fortnights
  and weeks with a day in it, and its days: each one's days with data, time below range (low and very low), time in range, time above range
  (high and very high), daily dose and daily carbs,
  newest first, each opening its AGP report. Arrows beside the month go to the month before, back a year, and after, up to this one; a
  month outside those is refused. A past month's empty table says it has none in that month, or year for months.
- Data written for a day before today, as a backfill or a late sync writes, or changed or deleted, builds that day's
  summary again, and the day it moved from when a change moves it to another day, and the whole weeks, fortnights and months with it in, without building the rest: once the site's
  data has gone 2 minutes without more, or after 15 minutes of it at most, so a backfill's many records are built
  once. Glucose, insulin, food, temp basals and targets, profile switches and device statuses count; today is built
  as it is read.
- The day a record is moved away from, through the site or Nightscout API, or deleted from for good, through the
  Nightscout API, is marked stale in the site's DB as it is written, so a server stopped before building it again
  builds it within a minute of starting.
- What was written is found from the data's `updated_at` since the site was last checked, kept on the site as
  `summaries_checked_at`, so data written just before a server stopped, to a day built or never built, is built within
  a minute of it starting too.
- A connector's finished backfill has the days it wrote built on the next run, within a minute, without waiting for the
  site's data to go quiet; and data from the last 15 minutes, as a phone uploads as it goes, does not hold back the
  building of older days written to, so a site with a looping phone has them built 2 minutes after they stop too.
- An admin can regenerate the summaries from below the Summaries page's tables, which says it is for summaries that look wrong, as after changing the timezone, rather than as the page's main action, after confirming: a background job builds the last
  year of days with data, and yesterday, and their weeks, fortnights and months, again from the site's data. It is
  recorded as an event, and asking again before it runs asks once. Others cannot.
- An AGP report says which days it covers, and has its time in ranges against the goals for type 1 and type 2
  diabetes: a bar of very high, high, in range, low and very low stacked, each with its percent, very high (goal <5%)
  and high bracketed with their total (goal <25%), in range on its own (goal >70%), and low and very low (goal <1%)
  bracketed with theirs (goal <4%), with the target range and the very high and very low bounds in the settings' units.
- An AGP report's glucose metrics are its average glucose (goal under 154 mg/dL), GMI (goal <7%), coefficient of
  variation (goal ≤36%) and time CGM active (goal >70%), each against its goal, and its days with data, daily dose and
  daily carbs, each "–" when it has no days with data.
- Its ambulatory glucose profile (AGP) over a day from midnight shades the 5th to 95th and 25th to 75th percentiles,
  with the median as a line, each in the high colour above the range and the low colour below it, with very low, the
  range's low and high, very high and its top marked with their values, each percentile named at its line's end (kept
  apart when close), and hours in the settings' time format. A day's report draws the day's glucose over the AGP of the
  7 days to it. A report with no glucose says so instead, and that reports are made from whole days, so a day's
  readings show the day after; the summaries' empty tables say the same.
- Pointing at or tapping a time on the AGP, or focusing it and moving with the arrow keys (Home and End go to its first
  and last), shows a tooltip beside that time, or the nearest with glucose, of the time in the settings' time format and
  each percentile's glucose then, named 95% to 5%, in the site's units, and marks them on the graph; it goes as the
  pointer leaves or the graph loses focus. A note under the legend says so.
- An AGP report's daily glucose profiles are its days as a calendar under their weekdays, a month's from the site's
  week start day and a week's or fortnight's from its first day, each with its glucose from midnight over the range shaded, coloured above and below it, opening that day's AGP
  report; a day's report has only its own.
- Day, Week, Fortnight and Month open the report of that kind ending on the report's last day, or today; arrow buttons
  either side of its dates (‹ and ›), named Previous and Next for a screen reader, open the one before and after (a
  week's or fortnight's the 7 or 14 days before and after, or to today when fewer are over), disabled past today or a
  year ago. Back and forward move between reports.
- Clicking around the reports quickly, their tabs, periods and arrows, shows the last one clicked, logging no error for
  the pages it replaced.
- The Reports hub has TDD: the total daily dose report at `/reports/tdd`, of the 30 days to today, today so far. It
  says which days it covers, dated in the site's timezone, with arrow buttons either side (‹ and ›) to the 30 days
  before and after (`?end=` the last day, as `2026-01-31`); the one after is disabled for the latest month, and the one
  before past a year ago. A last day after today or more than a year before yesterday is refused. Back and forward move
  between them.
- It draws a bar of each day's total daily dose, oldest first, on an axis of every day it covers, the bars a couple of
  pixels apart, its basal (in gold) under its boluses, with a legend of the two; a day with no insulin has no bar and is
  not counted. Hovering a bar shows its date and its basal or boluses in U. Today's bar is paler, labelled "so far" to
  its right, and its tooltip's date says so far.
- It shows how many of its days before today have insulin and their mean, lowest and highest daily dose, today left
  out as it is not over, and says when none have insulin.
- Under the bars, a table lists every day it covers, newest first, with its basal, boluses and total daily dose in U,
  or "–" for a day with no insulin; today's date is marked "so far", with insulin or not.
- Under the bars, a note says basal is estimated from the basal rate in force each minute, the profile's scheduled
  basal or a temp basal, the rest of the daily dose being boluses; and, when today is in it, that today, paler and
  marked so far, is left out of the days with insulin and their daily dose, lowest and highest.
- The Summaries page has a Printable report button, opening `/reports/print` in a new tab: this month so far and the 3
  months before it, newest first, each month's overview and AGP report on a sheet of its own, with none of the site's
  header, links or period buttons.
- Each page of the printable report is laid out as Dexcom Clarity's: its report's name large, over how many days it is
  of and its first and last (this month's to today), "Generated by Nighttroop" beside a small logo at its right, its
  report flat under ruled headings, and at its foot "Nighttroop", its number ("1 of 8") and the date and time it was
  generated in the site's timezone.
- On a screen the printable report's pages are white A4 sheets on grey, under a Print button; printed, each is an A4
  page of its own, in colour, without the grey or the Print button, its sections not split across pages.
- A site can build 5 printable reports a minute, with a burst of 10, and is told to wait after that. It is a page for
  people, so asking for JSON is not found.

### Getting started

- A site with no glucose for over a month, or ever, shows a guide to setting it up in place of its empty summary, until
  readings arrive, when the summary shows them live.
- The guide's steps each open the page they are done on: an admin gets their glucose in (the site's Nightscout URL for
  AAPS, Trio or xDrip+, or a sync from Dexcom, LibreLinkUp or Nightscout) and adds people and devices; everyone makes
  their settings and sets up alarms.
- Someone who is not an admin is told the admin sets up the site's data and who can use it.
- The guide is always in the user menu, as Getting Started, and on My Account, which says what it is and opens it.
- The guide ends saying where the rest is: readings, treatments and presets under Data, reports under Reports, and the
  site's settings, alarms, and logs and events in the user menu.

### Data

Each kind of data has a page with a table, newest first and paged, where records can be added, viewed, edited and
deleted (after a confirmation). Every confirmation, on the site and control, is a dialog in the site's own style rather
than the browser's: its question, Cancel, and a red Delete for a delete, including deleting the site, a red Remove
for removing a passkey, or Continue otherwise; Cancel, Escape or pressing beside it does nothing. Its Add button, the
page's main action, is at the right of its title (below it on a
phone), as on the presets and push credentials pages. On a phone, a table wider than the window scrolls sideways in its card, so the page
does not, and its cells stay on one line, so a time or a value and its unit never break apart; only notes wrap.
Adding or deleting a record updates the table's "Showing" count straight away, without a reload, and deleting the
last one shows the table is empty again, as deleting the last push credential does. Each page shows 50 records at a
time, with Next while there are more after them, up to the page starting at the 32001st, the furthest a page skips;
a page asked for further back starts there. A box at the top of each page says how many records there are over
every page, as "1 temp target" or "2 temp targets", and the days of the earliest and latest, without their times, as
Today, Yesterday, "Sun 8 Nov" this year or "8 Nov 2025" before, each with how long ago it was (a device's history
counts that device's statuses); with none it says 0, with a dash for each day. On a phone its three figures stay side
by side. The site API's search gives the same: its `items`, `skip` (at most 32000, a negative one 0), `limit` (50
unless asked for, at most 1000) and, for these kinds, `total`, `earliest_at` and `latest_at`:

- **Glucose**: CGM and meter readings with trend arrows, in the settings' units (mmol/L or mg/dL), with lows and
  highs coloured as on the summary and marked L or H, read out as Low or High, so they are not shown by colour alone.
  - Adding a reading offers its type as Meter or Sensor and its trend as an arrow with its name (e.g. `↗ FortyFiveUp`),
    as the table and its details show them.
  - Above the first page, the last 10 days' glucose are small graphs on the same scale, oldest first, each with the
    range shaded and its time in range; the last is today so far, from its clean readings. They are titled "Last N
    days", or Today on a new site's first day.
- **Insulin doses**, with a combo bolus's split as the fractions given at once and over time, 0.7 for 70%, typed in
  and shown on the site as percents, 0 to 100
  - A dose's `automatic` is whether a loop gave it, true or false as Loop uploaded it and null for apps that do not
    say; the site API gives it but does not change it, ignoring it on a create or update.
- **Food intakes**
- **Temp basals**, with the rate given in U/hr and a multiplier of the scheduled basal, 0.5 for half, 0 a zero temp;
  the site types it in and shows it as a percent, from 1% to 1100%
- **Temp targets**, in the settings' units
- **Profile switches**, with their profile's schedules to read, and a multiplier of its insulin, 1.2 for 120%; the
  site types it in and shows it as a percent, from 10% to 500%
- Every multiplier is labelled with its unit, (%), and explained the same way: "A percent above 0: 100% is as it is,
  120% is 1.2 times." The API keeps and returns the fraction.
  - A profile switch's details show its profile's basal, carb ratio, sensitivity and target schedules as tables, as
    profiles do, with sensitivities and targets in the settings' units and the units in the headings.
- **Activities**
- **Notifications**
- **Event groups**: the treatments that came in together through the Nightscout API, with their events.
  - An event group can also be added on the site: the Add Event Group menu lists the Nightscout types (e.g. Meal Bolus),
    and each opens a modal for that type with a time and up to one event of each kind that type can have (e.g. a Meal
    Bolus can have glucose, insulin and food, an Exercise only an activity), picked from those within an hour of the
    time. Changing the time lists the events near it. The events'
    times, and the window they are picked from, are in the time format, as other times are.
  - Adding or editing a group with a kind of event its type cannot have, or of an unknown type, is refused, as is one
    with no events, unless it is a change.
  - Sensor, site and pump part changes (Sensor Start, Sensor Stop, Sensor Change, Site Change, Cannula Fill, Prime,
    Insulin Change, Reservoir Change, Pump Battery Change) are event groups with no events: the menu lists them, and
    their modal says there are no events to pick, with just a time, reason and notes, and for a sensor's start, stop
    or change its sensor code and transmitter ID. There is no separate page for them.
  - Clicking a group edits its type, time, notes and events; changing its type offers only the kinds of event the new
    type can have. An event taken out of a group is in no group.
  - An event is in at most one group: events in another group are not offered, and grouping one again is refused.
  - A group added or edited on the site is a Nightscout treatment, with its events.
  - A group's details show the pump its app named (ID, serial and type), as AAPS does on every treatment, and a sensor
    change's sensor code and transmitter ID, which its row's events list too.
- **Device statuses**: each device's latest 3 statuses, and every status a device sent, with each column's unit in its
  header, e.g. Phone battery (%), Reservoir (U), IOB (U) and COB (g).

Every value written to a site, from its pages, the Nightscout API or a sync, has a broad limit far outside real use,
e.g. a bolus of at most 500 units or 1000 g of carbs, and a date from 1970 to the year 9999. Insulin and food can be
negative, to take back too much entered. The pages are stricter, e.g. a dose of at most 100 units either way.

Adding a record on a data page opens a modal of all the fields a person can type in, each saying what it is, in what
units, whether it is required, with its limits, and a dropdown or suggestions for its types. Its time is typed in the
settings' timezone, and is now when left blank; other fields left blank are not set.

A record's details, on a data or presets page, edit it in a form whose every input is labelled, which wraps onto more
lines on a phone so each input can be seen and tapped. It has every field its add modal has, with the same labels,
help and limits, filled in with the record's, but its time and those an edit does not change (a reading's type and
method, a temp target's insulin needs). Emptying its notes or reason clears them; a number left blank stays as
it was. Its Save (Rename for a profile) is at the modal's foot beside Close, as an add modal's Add is, and Enter in
an input saves too. Each field is shown once: one its form edits (e.g. an insulin dose's split, type, notes and
reason, or a reading's trend) is not listed again in the details below the form, which keep only what it does not
edit (e.g. a reading's type, a profile switch's DIA), and end with the same Record section on every page: its ID,
UTC offset, app, device, who created and last updated it and when, each left out when it is not set.

Without a mouse, Tab reaches each record's row in a table, here and on the control pages, outlined in gold, and Enter
opens its details; Enter on its delete button deletes it without opening them. A row's delete icon, here and in
settings, is grey, and red only while it is pointed at or focused, so a table is not a column of red. A modal takes the focus when it opens,
so Tab goes on to its fields and a screen reader reads its title; Tab past its Close button goes round to its first
field, and Shift+Tab from its first field to Close, so the focus never reaches the page behind it. Closing it (Escape,
Close or saving) gives the focus back to the row or button that opened it.

### Presets

- Food Presets, Profiles and Loop Settings are in the Data hub under Presets, with the data they are used to enter.

- **Food presets**: added in a modal of all their fields, as data is, viewed, edited and deleted.
- **Profiles**: profile stores uploaded by a loop, with their schedules to read, renamed or deleted.
- **Loop Settings**: each different set of settings a loop app sent, kept once however often it was sent: Loop's from
  its profiles, Trio's remote settings from its profiles, and AAPS's configuration from its device statuses (an empty
  one is none sent). Each shows its app, what it came in, its device and when it was first and last seen, newest
  first; clicking one shows the settings as sent, indented to inspect.
  - A read only identity reading them from the site's API gets them without Loop's or Trio's `deviceToken`, and no
    change to one.
  - Clicking a set also shows what changed from the same app's set last seen before it: each setting changed, added
    or removed, by its name (a preset's by the preset's name, as `overridePresets.Running.targetRange`), old then new
    ("not sent" when it was not), glucose in the site's units; an app's first set says there is no set before it. The
    API gives these as `changes`, with the set before's `previous_id` and the `units`.
  - Deleting a set keeps the profiles and device statuses sent with it; the Nightscout API reads them back without it.
  - The Nightscout API reads each profile and device status back with the settings it was sent with.

### Connectors

Connectors are a site's third party integrations, each syncing from one service.

- A site can pull data from **Dexcom** (with its region), **LibreLinkUp** (with its region) or another
  **Nightscout** site, or from the **virtual connector**, which makes realistic test data.
- The virtual connector writes its readings, treatments and device statuses on the 5 minute grid, whatever window it
  syncs or backfills, so windows that overlap write the same records at the same times and none is written twice.
- The virtual connector's glucose is the same for the same time on every site: a 6 hour wave, with a meal as it rises
  and a correction at its peak, over a 25 hour one, so days differ. It reaches every range, about 2% very low (below
  54 mg/dL), 4% low, 63% in range, 27% high and 4% very high (above 250 mg/dL), and never reads below 40 mg/dL. It
  also has a finger prick meter reading twice a day, at 07:02 and 19:02 UTC, 6 mg/dL above the CGM then.
- The virtual connector's first sync writes the last 30 minutes, always 6 CGM readings and no meter reading, whatever
  time it runs; a backfill writes a meter reading in them.
- The virtual connector is only on a development, test or staging server (`MODE`). In production it is not offered on the
  Connectors page, saving or turning on one is refused, and one already saved is never synced or backfilled.
- Saving does not check the details: the first sync does, and turns syncing off and says why if it fails.
- A site's saves of its connectors, adding a source as well as changing one, share one rate limit
  (`SITE_UPDATE_CONNECTORS_PER_MINUTE_LIMIT`); one past it is refused (429).
- Typing the details of a Dexcom, LibreLinkUp or Nightscout source not yet saved ticks its Enable Connector box, as it is easy
  to forget; a click on the box, before or after, is kept, and a saved source's box is never changed.
- A failing source syncs less often, and says when it failed and why. Why, as stored, logged and sent to Sentry, never
  has the Nightscout's URL or host, the Dexcom session, or the followed LibreLinkUp patient's ID, as a timeout's would.
- A record a sync or backfill cannot read, such as a Dexcom or LibreLinkUp reading in a shape it does not know, or that
  the site refuses, such as a Nightscout treatment with an eventType it does not know, is skipped. The first of each
  sync is recorded as an error of the site's connector with the record (up to 8 KB), to debug it by. It is logged and
  sent to Sentry with only the names and types of the record's fields, never their values.
- A working source can be backfilled with the days before its oldest reading, from midnight in the site's timezone,
  so its data starts at the start of a day. A Nightscout or virtual backfill also
  fetches the last 2 days again, to fill gaps the site missed while it was down, as for a deploy, and its toast says so;
  Dexcom and LibreLinkUp backfills fetch all they keep, the last 24 and 12 hours. A Nightscout backfill keeps the
  device statuses it adds without their loop status (`openaps` or `loop`) or `pump` as sent, which are of no use once old
  and most of a looping site's DB, but with their IOB and COB, which the summaries show, and their pump's battery,
  reservoir, clock, manufacturer, model, `pumpID`, `suspended` and `bolusing`; a status the site already had is left
  as it is.
- A Nightscout or virtual backfill copies no device status older than 30 days, though a Nightscout server returns one,
  as an old one's IOB and COB are of little use and device statuses are most of a site's DB; its other records are
  copied however old.
- A connector not saved yet has only Save. Once saved, without a reload, it also has Backfill, and Delete as a red
  outline button, apart from Save, so the destructive action is not the boldest on the form.
- The Connectors page shows how each source has synced, how often, the records it has pulled in, in all, from its syncs
  and backfills, and in its last sync, its backfill and its recent events. It shows the date its last backfill fetched
  back to, and how long ago that is, so how far back its backfills have gone; - until it has been backfilled.
- The Connectors page updates as each sync or backfill finishes, showing its new errors, stats and events without reloading,
  leaving a form being edited as it is and an opened history open, even one opened as it updates. While it does,
  ● Live shows beside its title.
- Connector credentials are encrypted at rest.
- Each secret kept encrypted (connector credentials, TOTP secrets, push secrets and tokens, Nightscout tokens and the
  welcome's passkey cookie) is bound to what it is and its site, and each backup to its object key without its host,
  so one copied in as another site's, another kind of secret, or another batch's backup does not decrypt, nor one
  encrypted without them. Backups copied to a bucket for another host still restore.
- A connector's password, API secret or token, and its username, email or URL, are not offered to the browser or a
  password manager to save or fill, as they are not this site's sign in: the password is a masked text box, not a
  password box, which Chrome offers to save whatever its autocomplete, and its eye still shows it.
- A Nightscout connector copies the newest readings, treatments and device statuses, and the current profile store
  whenever the Nightscout's profiles change, however long ago it started; an edit to it in Nightscout's profile
  editor replaces the copy, whoever saved it there or whatever identifier its uploader, such as AAPS, gave it, and a
  new one is added beside it.
- A Nightscout connector on this server's host, or on a private, local or shared (carrier-grade NAT or Tailscale) address,
  is refused when saved and as it connects.
- A Nightscout connector on this server's host written with a trailing dot or capitals, such as
  `https://abcdefgh.ns.HOST./`, is refused, and a URL with a trailing dot is saved without it.
- A Nightscout connector on a multicast, reserved, broadcast, benchmarking or `0.x.x.x` address, or on an IPv4-mapped,
  NAT64 or 6to4 IPv6 address that reaches a private or local IPv4 one, is refused when saved and as it connects.
- Pasting a Nightscout link with its token, such as `https://site/api/v1?token=...`, `?api_secret=...` or
  `https://secret@site/api/v1`, into the Nightscout connector's URL splits it into the base URL and the API Secret or
  Token box.

### Alarms

- An alarms page with urgent low, low, predicted low, high, urgent high, fast rise, fast drop, no data, low device
  battery and low reservoir alarms, each off by default.
- The site's alarms are the default for each identity, who can change them for themselves.
- An active alarm shows in a banner under the header on every page, live, and can be snoozed.
- Nonsense settings are refused, e.g. a high under the low.

### Settings

- General settings: units, time zone, language, date and time format, week start, insulin and carb defaults.
- An admin sets the site's settings; each identity can override them for themselves.
- Every page's text is in the language setting, the identity's, then the site's, then the browser's, and in English
  where there is no translation; the control site and public pages are in the browser's language. Only English is
  written so far.
- Every page shows dates in the date format, the identity's, then the site's, then the default (YYYY-MM-DD), times in
  the time format chosen the same way (24-hour HH:mm:ss by default, or 12-hour, or without seconds), and both in the
  timezone chosen the same way; the logs pages too.
- On a phone the settings, alarms and another identity's settings are one column: each setting's label, then its Site
  and Yours controls, each named, and the save buttons fit in the window.
- An admin can **delete** the site, after a confirmation, which archives it to S3 and signs everyone out.
- An admin can add, list and delete the site's **push credentials** on Settings, Push Certificates
  (`/settings/push-credentials`): an Apple APNs key (team ID, key ID and the `.p8` file pasted in) or a Firebase
  service account (project ID and its JSON), the apps it pushes to (`loop`, `trio`, `loopfollow`, `t1dcarer`) ticked,
  and, optionally, the only bundle IDs it pushes to, separated by commas. A key that cannot be read, or an unknown
  app, is refused with the modal left open; its secret is kept encrypted and never shown or returned; a site has at
  most 5; only the site's admins can see or delete them, and a delete asks first, then deletes it for good, so adding
  and deleting them again and again cannot grow the control DB.

### Logs

- An admin sees the site's events (e.g. created, alarms, recoveries, credential changes), errors and logs, filtered
  by type and searched, and its batch's backups. Its errors include its requests refused for what they sent (kind
  `client`), with why, and which identity and app sent them, so the admin can see what a client gets wrong.
- A toast never shows the server's own words: a server error, or a refusal in the database's words, says "Something
  went wrong. It has been recorded on the Errors page.", linking an admin there, or that it is recorded for the site's
  admins to anyone else. The Errors page keeps what went wrong, in full. A refusal written for users shows its reason,
  without the "Bad or missing request body" or "Constraint check failed" start the API's message has.
- Toasts show above an open modal, so an error from a form in one can be read and its link clicked.
- A search of the site's events, errors or logs, or a filter by error kind, looks through only its newest 10,000 of
  each, so no search reads every row a site has; paging without one reaches every row.
- Each of the site's own events, errors and logs shows the request or background job run it was for; its batch's
  events do not.
- Each sign in to the site, and each refused one, is an event with the country (and city, where the server has
  DB-IP's City Lite) its IP address is estimated to be from. A refused one does not keep the name it tried.

### Nightscout API

- Each site has a Nightscout compatible API at `<site>.ns.<host>` for apps like AAPS, Loop, Trio and xDrip, and
  followers.
- **v1**: entries, treatments, device statuses, profiles, food and activities, with Nightscout's filters, plus
  status, versions, verifyauth, slice, count, times and echo, experiments/test, notifications/ack and adminnotifies.
- v1 `echo` shows the request's method, URL, query and headers with credentials (Authorization, Cookie, api-secret,
  token) `REDACTED`, and how its search is understood.
- `/pebble` has the newest `count` readings (1 by default, at most 1000) from the last two days for Pebble and other
  watch faces, each glucose a string in the `units` asked for (`mmol` or `mg/dl`) or the settings' units, with its
  trend number, and the newest with the delta, the uploader's battery and IOB, as Nightscout's has them.
- v1 `notifications/ack?level=&group=&time=` snoozes the identity's alarms in that group up to that level for `time`
  milliseconds, as the `/alarm` socket's ack does, and answers `OK`; any identity that can read may ack its own.
- v1 `adminnotifies` has no notices, as Nighttroop has none of Nightscout's settings to warn its admin about.
- v1 `alexa` and `googlehome` answer Nightscout's Alexa skill and Google Home action, for an identity that can read:
  `NSStatus`, `UploaderBattery` and `MetricNow` of bg, delta, iob, cob and uploader battery, said as Nightscout says
  them in the settings' units. An Alexa launch is welcomed and keeps listening, and an unknown intent or metric is
  answered that it is not known, with Google Home listening for another.
- v1 `experiments/test` is `{"status":"ok"}` for an identity that can upload readings and a 401 for any other, so
  Loop's and xDrip4iOS's connection tests pass with a PWD phone's token.
- A v1 search by `find[_id]` that finds nothing, such as a deleted treatment, is an empty list, not a 404.
- A search with a filter on a field or with an operator Nighttroop does not know, a v1 `find[...]` (e.g.
  `find[notes]=x`) or a v3 `field$operator` (e.g. `notes$eq=x`), is an empty list, as nothing would match in
  Nightscout. Any other parameter, such as xDrip+'s `rr` or `raw`, is not a filter and is left out, so a document is
  always as Nighttroop keeps it, never the JSON it was sent as.
- A v1 `DELETE /api/v1/entries.json` with `find[type]` and `find[date][$gte]`, `find[date][$lte]` or
  `find[date][$in][]`, as xDrip4iOS deletes readings, soft deletes the entries found, at most 1000 each time, saying
  how many in `deletedCount`; any other filter, or none on the date, is a 400, and it needs a token that can delete.
- **v2**: authorization (a token for a JWT, and a check of one of its permissions), properties and summary.
- **v3**: create, read, update, patch, delete, search and history of each collection, plus version, status,
  lastModified and settings.
- A v3 replace or patch with `If-Unmodified-Since` of a document changed in a later second is refused with a 412 and
  changes nothing, checked in the same transaction as the change, so a client cannot overwrite a change it has not seen.
- The v3 status shows what the token may do to each collection, e.g. `crud` for an uploader and `r` for a read only
  token, so apps like AAPS only upload with a token that may write.
- A v3 history includes deleted documents, with `isValid` false, so apps learn of deletions; a search leaves them
  out.
- A v3 history is the documents with a `srvModified` at or after its time, the oldest first, then by date, as in
  Nightscout, at most its `limit`, and its `ETag` is the newest `srvModified`, so an app paging from the `ETag`, as
  AAPS does, gets every change, and the last profile store in a profile history is the newest changed.
- A delete moves a document's `srvModified` and the v3 lastModified forward, so a history since an app's last sync
  has the documents deleted after it.
- **Socket.IO** live updates and alarms over long polling or websockets.
- A Socket.IO client polling upgrades to a websocket at once: the probe ends its waiting poll, as the JavaScript client
  that AAPS and Nightscout's pages use waits for it before upgrading, rather than reconnecting every minute.
- A Nightscout food has `type` `food`, as AAPS loads only foods of that type.
- The Nightscout host's root, which LoopFollow's Nightscout tab opens, goes to the site's own pages.
- A live update sent to a site's sockets is in the site's own settings, such as its time zone, as it goes to every
  device, with glucose in mg/dL.
- A new Announcement is sent to every `/alarm` socket that can read, with its notes as the message, so AAPS shows it
  as a notification.
- Any other new treatment from the last 10 minutes, except a Temp Basal or Temporary Override, is sent to every
  `/alarm` socket that can read as a `notification`, titled with its event type and saying who entered it and its
  notes, as Nightscout's treatment notify does.
- A Socket.IO `authorize` with a token that is no one's is disconnected unanswered, as Nightscout does; a wrong API
  secret is answered that it can do nothing. Either way, once a socket has been refused, at once for a token and on the
  third refusal otherwise, its connection is closed a second later.
- A Socket.IO `authorize` or `subscribe` before its socket is given access spends its IP's budget on the site
  (`SITE_UNAUTH_PER_MINUTE_LIMIT`), as a request that fails auth does, not the site's own, so one IP's failed sign ins
  cannot leave the site's pages and followers with 429s.
- A Socket.IO connection that joins no namespace within 45 seconds is closed, and a socket not given access within 30
  seconds (by an `authorize` that can read or write, or a `/storage` or `/alarm` `subscribe` that succeeds) is closed
  with its connection, so connections that never sign in cannot hold a site's live sessions from its followers. One IP
  (an IPv6 /64) may hold at most 10 of a site's live sessions not yet given access
  (`SOCKET_MAX_UNAUTH_SESSIONS_PER_IP`); its next is refused with 429, so it cannot take them all.
- Glucose is stored in mg/dL and always returned in mg/dL, whatever the identity's or site's units, as Nightscout's
  clients read `sgv`, `mbg`, a treatment's `glucose`, a temp target's `targetTop` and `targetBottom` and a profile's
  targets as mg/dL: entries, treatments, profiles and their `units`, over HTTP and Socket.IO. It is accepted in the
  units sent with it, or guessed from its value.
- A profile, or a profile switch's `profileJson`, sent in mmol/L comes back in mg/dL in every glucose: its `sens`,
  `target_low` and `target_high`, and Loop's `loopSettings` `minimumBGGuard`, `preMealTargetRange` and each override
  preset's `targetRange`, so a reader never sees an mmol/L sensitivity or suspend threshold labelled mg/dL. Every treatment comes back with `units`, a temp target and a therapy
  event (e.g. a Note or Sensor Change) too, so a follower on an mmol/L site does not read a target of 90 as mmol/L.
- The v1 status has the identity's own units in `settings.units`, as Nightscout writes them (`mg/dl` or `mmol`), so
  followers show glucose in them, and `/pebble` and the virtual assistants say it in them; the v3 settings are the
  identity's own, units included, and each, a default too, has its `srvCreated` and `date` in milliseconds.
- The v3 settings also keep AAPS's own documents as sent, shared by every identity of the site: its running config
  (`aaps`, `aaps-state`) and the messages AAPS and AAPSClient pair and control it with (`aaps_clientcontrol_...`).
  Each is created, replaced, patched, read, searched, deleted and sent to `/storage` sockets as any document is,
  and the settings' last modified is when one last changed. A document over 256 KiB, or a new one once the site
  keeps 100, is refused; client control messages left unchanged a week are deleted to make room.
- The settings `units_mgdl_label` and `units_mmol_label`, set only through the settings endpoint, say how the
  Nightscout API writes each units in a document's `units`: mg/dL as `mg/dL` or `mg/dl`, and mmol/L as `mmol/L` or
  `mmol`, Nighttroop's way by default, though only mg/dL is written. Units are stored as Nighttroop's, and any of these
  is read.
- A v3 replace or patch of a setting answers with the setting as stored; a read, replace, patch or delete of an
  identifier that is not one of Nighttroop's settings is a 404, and a create of one a 400, rather than an empty setting.
- The v1 status has `authorized` with a JWT and the token's permissions, with `api:treatments:create` for one that
  can write, so LoopFollow shows a read token as "Read" and a write token as "Read & Write".
- The v3 version needs no token, so an app can check for v3 before it signs in.
- An entry has an `sgv` if it is a CGM reading and an `mbg` if it is a meter reading, never both.
- An `mbg` entry keeps AAPS's `isCalibration` and `sensorMgdlAtPairing` (in mg/dL) and returns them as sent through
  v1 and v3, so AAPS's followers read a calibration as one; an update without them keeps them.
- A date filter takes milliseconds with a fraction (e.g. `date$gte=1761234567890.25`), which is dropped.
- A document sent with a date but no `utcOffset`, as AAPS sends a device status, takes the site timezone's offset at
  that date; a replace sent with none keeps the offset of the document it replaces.
- A `created_at` filter or sort is on when it happened, as `date` is, and `srvCreated` is on when it was uploaded.
- A document's `created_at` is written as Nightscout writes it, as JavaScript's `toISOString` does, in UTC with three
  fractional digits: `…23.940Z` sent comes back as is, not `…23.94Z`, and `…23Z` comes back as `…23.000Z`. So are a
  treatment's `timestamp`, an entry's `dateString` and a profile's `startDate`, in their UTC offset, so an entry is
  found and deleted by the `.SSS` `dateString` sent, as Trio's deleteGlucose does. Dates kept as sent, such as a
  device status's `pump.clock`, which LoopFollow reads without fractional seconds, are not changed.
- A v1 or v3 search with `sort` on the date (e.g. `sort=date`, `sort=created_at`) is the oldest first, paged with its
  `limit` and `skip`, so an app paging forward from the newest date it got, as AAPS does, gets every document; any
  other search is the newest first. With a filter other than the date, a page is those that match of the page the
  dates alone would give, so it can have fewer than its `limit`, still in order.
- A v1 or v3 search's `skip` is at most 32000, enough for Nightguard's 32 pages of 1000, and a negative one is 0.
- v3 filters take `field$eq` as well as `field`, e.g. `eventType$eq=Site Change`.
- A device status keeps its pump's other fields, e.g. AAPS's `extended` with its profile and temp basal.
- A profile from Loop, with `loopSettings.deviceToken` and `bundleIdentifier`, or from Trio, with them at its top,
  registers that app's APNs token for pushes for the identity that uploaded it, a production build unless
  `isAPNSProduction` is false; a token that is not one is not registered, and the profile is still kept.
- A profile keeps Loop's and Trio's `deviceToken` whole, up to 256 characters, as Apple's are of no fixed length
  (Trio's are 160 hex characters), and registers it whole; a longer one is kept empty and not registered.
- A profile's Loop `loopSettings.deviceToken` and Trio `deviceToken` are only given back, in v1, v3 and the socket's
  storage, to an identity that can change the site's data, as LoopFollow sending remote commands is; a read only
  user or device gets the profile without them, and the socket's `dataUpdate`, sent to every reader, has none.
- A profile switch's `percentage` is kept as a multiplier of the profile's insulin (120 as 1.2) and returned as sent.
- A profile switch, or AAPS's effective profile switch (a Note with a `profileJson`), keeps the switch it was made
  from, `originalProfileName`, `originalCustomizedName`, `originalTimeshift`, `originalPercentage`,
  `originalDuration` and `originalEnd`, and returns them as sent through v1 and v3, so AAPS's followers read an
  effective profile switch as one; an update without one keeps it.
- AAPS's running mode, an `OpenAPS Offline`, keeps its `mode` (e.g. `DISCONNECTED_PUMP`), `autoForced`, `reasons` and
  `originalDuration` (ms) and returns them as sent through v1 and v3, so AAPS's followers read its mode; an update
  without one keeps it.
- A combo bolus's `splitNow` and `splitExt` are kept as the fractions given at once and over time (70 as 0.7) and
  returned as sent.
- A temp basal is sent with its rate in U/hr as `rate` or `absolute`, its `percent`, or both. Its `percent` is
  Nightscout's change from the basal, as AAPS sends it (e.g. -50 for half, 20 for 1.2x, -100 a zero temp, never under
  -100), kept as a multiplier of the basal (0.5, 1.2, 0). The one not sent is set from the default profile's or
  profile switch's basal at its start, in the profile's timezone; a rate over 11 times it keeps no percent.
- A temp basal is always returned with its `rate` and `absolute`, both its U/hr, and its `percent` when it has one.
- A temp basal keeps AAPS's `type`, e.g. `NORMAL` or `PUMP_SUSPEND` for a suspended pump, and returns it as sent
  through v1 and v3, so AAPS's followers show a suspend as one; an update without it keeps it.
- The summary's basal during a temp basal is its rate.
- Duplicate uploads are ignored; values far outside real use are refused.
- A value sent that means none is not stored, and comes back left out, as one not sent does: an empty string, a
  CGM's `noise` or `rssi` of 0, and a heart rate of 0. An update with one keeps the value stored. `false`, and other
  zeros, such as a combo bolus's split, are kept.
- An entry's `filtered` and `unfiltered` come back as sent, 0 too, as AAPS sends 0 with every reading.
- A treatment uploaded again with the date and type of one stored, or its identifier and date, is merged into it,
  keeping its `_id` and what the upload leaves out, as Loop sends a temp basal or bolus again when it ends, and Trio a
  temp target or override run; followers get it as changed, and one sent again as it was stored changes nothing.
- A treatment with only carbs at the date and type of one with only insulin, or the other way round, is its own
  treatment, as AAPS's Bolus Wizard sends its bolus then its carbs as two Meal Boluses and deletes either on its own.
- An entry uploaded without an `_id` with the date, type and device of one stored takes its new value, direction and
  raw values, as xDrip4iOS smooths readings, and followers get it as changed; one sent with the same values changes
  nothing, not even when it was last modified.
- A meter reading (`mbg`) uploaded with the date of one stored is the same finger prick whatever its device, so a
  careportal BG Check that xDrip+ downloads and uploads again as its own is stored once.
- A `created_at` filter given a time with no zone, as LoopFollow sends it, reads it as UTC. So does one given only to
  the minute, such as `find[created_at][$gte]=2026-11-06T04:46`, as some clients send it; to the minute with a zone,
  it is that time in its zone.
- Loop's and Trio's Suspend Pump and Resume Pump treatments are kept.
- The pump AAPS names on every treatment, its `pumpId`, `pumpSerial` and `pumpType`, is kept with the treatment and
  returned as sent.
- An AAPS bolus's `isBasalInsulin`, and the insulin a bolus or profile switch was given with, its `icfg`
  (`insulinLabel`, `insulinEndTime` and `insulinPeakTime` in ms, `concentration`), are kept and returned as sent; other
  fields of an `icfg` are left out.
- An AAPS Bolus Wizard's `bolusCalculatorResult`, the JSON text of its calculation, is kept with the treatment and
  returned exactly as sent, as AAPS restores the record from it; one that is not JSON or is over 4096 characters is
  left out, and the rest of the treatment kept.
- A treatment's `sensorCode` and `transmitterId`, as xDrip+ sends with a sensor change, are kept with it and returned
  as sent. A sensor, site or pump part change (e.g. `Sensor Change`, `Site Change`) is an event group with no events, of
  its type, which the Nightscout v2 properties' ages (`sage`, `cage`, `iage`, `bage`) and the summary read.
- An activity's `type`, as xDrip+ sends it, is its activity type, a heart rate's `bpm` its heart rate and a motion's
  `class` its name.
- A device status keeps its pump's clock, manufacturer and model, its uploader's battery, returned in both
  `uploader.battery` and `uploaderBattery`, and Loop's override in force.
- A device status keeps Loop's pump `pumpID`, `suspended` and `bolusing`, as sent, once the status as sent is cleared,
  so a follower still sees an older status's insulin delivery suspended.
- A device status keeps Trio's pump `bolusIncrement`, LoopFollow's remote bolus step, while the status as sent is kept.
- A Nightscout treatment comes back with its `created_at` as `timestamp`, a bolus with its insulin as `programmed`,
  and a temp basal with `temp` `absolute`, as Loop's NightscoutKit drops treatments without them; a `timestamp`
  sent is not read. The `timestamp` is only given through API v1 and its sockets: API v3 and its storage socket leave
  it out, as AAPS reads it as a number and drops a treatment, or a page of them, whose `timestamp` is a date.
- A treatment's `durationInMilliseconds` is kept, created or updated, and given back as sent, as AAPS ends a temp
  target early to the millisecond and sends extended carbs and therapy events with it, while it is within a minute of
  the `duration` kept, so not once that is changed on the site; one sent with no `duration` is in minutes too, the
  nearest. A treatment with a duration is always given back with both `duration` and `durationInMilliseconds`, the one
  not kept worked out from the other.
- A device status keeps its uploader's `isCharging`, `name` and `timestamp`, and, while the status as sent is kept,
  Trio's pump `status.bolusing`, `status.suspended` and battery `string` and `display`.
- A profile keeps Trio's remote control settings (bundle, device token, APNS environment, team and expiry) and its
  override presets, for LoopFollow's Trio Remote Control.
- A v1 upload of a document with an `_id` replaces the one with that `_id`, or is created with it, as Loop sends an
  override again with its `_id` when it changes, and deletes it by that `_id`.
- Loop's `Temporary Override` is kept as a temp target, with its correction range and insulin needs scale factor; an
  indefinite one lasts a week, and is given back with `durationType` `indefinite` while the override as sent is kept.
- An identifier over 36 characters, like an Omnipod dose's `syncIdentifier`, is kept as a hash of it, so two that
  start the same are not taken as one.
- v1 `profile/current` is the newest profile set itself, not a list of one, with `enteredBy`, as Loop reads it, and
  `null` on a site with none, as in Nightscout.
- A profile search with no filters is the newest profile set, or with a `count` that many, newest first by when each
  starts.
- A device's Nightscout token is also taken as its SHA1 in the `api-secret` header, as xDrip+ and Loop send it; a
  password reset stops the old token's SHA1 working.
- A treatment's `eventTime`, as xDrip4iOS and Nightscout's careportal send it, is its `created_at`; an eventType of
  `Carbs` is a carb correction, and a glucoseType with the reading after a colon, like `Finger: 6.7 mmol/L`, is a
  finger prick.
- Entries of type `cal` (a CGM calibration's slope and intercept) are accepted and not stored, so the finger pricks
  sent with them are kept.
- A device status's `uploader.battery` is its uploader's battery when it has no `uploaderBattery`.
- A device status with a battery or reservoir level out of range (an uploader or pump battery below 0 or over 100%, a
  reservoir below 0 or over 10000 U), like the `uploader.battery` -100 Trio sends when the phone does not know its
  battery, is stored without that level, and the rest of it kept.
- A device's Nightscout token is also taken as its SHA1 in the `api-secret` header, as xDrip+ sends the token in its
  URL; a password reset stops the old token's SHA1 working.
- An xDrip+ upload works as it sends it (see `docs/clients/xdrip.md`): a calibration (`type` `cal`) in a batch of
  entries is left out and the readings with it kept; a reading out of range (0 or less, or over 1000 mg/dL) in a
  batch is left out and the rest stored, and a batch with only readings out of range is refused; an uploader battery
  that is not a number is left out of its device status; `uploader.battery` is the uploader's battery; and a treatment is found by its `uuid` (`find[uuid]`), sent as the `_id` or as xDrip+ makes its `_id`.
- Trio's and iAPS's `SMB` and `External Insulin` treatments are kept as insulin doses, with their event type.
- A Loop bolus's `automatic` comes back as sent, true or false, through API v1 and v3, so LoopFollow shows an
  automatic bolus as an SMB and leaves it out of its bolus stats.
- A Loop treatment's `syncIdentifier` comes back as sent, up to 256 characters, through API v1 and v3, as Loop finds
  its own by it, though a dose's `_id` is a hash of an Omnipod one; so do its `insulinType`, a carb entry's
  `userEnteredAt` (in UTC) and a bolus's `unabsorbed`.
- xDrip+'s `Sensor Stop` treatments are kept as an event group, with their event type, so the treatments it uploads
  after one are kept too; the site can add a Sensor Stop event group.
- A profile's `enteredBy` comes back as sent, e.g. `Loop` or `Trio`, as a treatment's does, so a tool picking a loop's
  profiles by it finds them, else the identity that uploaded it, as Trio needs one to import a profile.
- A treatment's `enteredBy` comes back as sent, e.g. `loop://iPhone` or `Trio`, else the identity that uploaded it.
- A treatment's `uuid` comes back as sent, as xDrip+ matches its own by it; one sent with no `_id` is its identifier.
- Treatments can be found by whether they have `carbs` or a `duration` (`find[carbs][$exists]`), by `enteredBy`, and
  without some `enteredBy`s (`find[enteredBy][$ne]`, repeated or under `find[$and][n]`), as Trio and iAPS fetch them.
- A treatment's or entry's `id`, as Trio sends it, is its identifier when it has no `_id`.
- A v1 `DELETE /api/v1/treatments.json` or `entries.json` deletes the same way by Trio's and iAPS's filters:
  `find[id][$eq]`, `find[created_at][$eq]` with `find[eventType][$eq]`, `find[dateString][$eq]` and
  `find[carbs][$exists]`, and each of up to 10 `find[$or]` alternatives, every one of which needs a date or an id.
- A create that names its `identifier` keeps it, and only a document with that identifier is its duplicate.
- A create naming a deleted document's identifier, or a replace of one, replaces it whatever changed, as Juggluco
  changes a treatment by deleting it and sending it again.
- A treatment with `eventType` `<none>`, Nightscout careportal's treatment with no type, is a Note keeping its
  insulin and carbs.
- A treatment with AAPS's `eventType` `SettingsExport`, sent when its settings are exported, is a Note.
- A `dateString` with Juggluco's `-03:-30` zone for half hours west of UTC is read as `-03:30`.
- An entry's `date` is kept to the millisecond, a fraction of one dropped, though its `dateString`, as Loop's does,
  has only whole seconds; the `dateString` gives its UTC offset.

### Nighttroop API and client

- Every site page is also a JSON API, with a JWT from signing in.
- A Go client, `clients/nighttroop`, is generated from its OpenAPI schema. It can create a site, sign in, read the
  summary, and list, read, add, edit and delete glucose readings, insulin doses and food intakes.
- `GET /data/summary-periods/{period}-{date}` reads the day, week, fortnight or month with that date in it, from today
  back a year, in the site's timezone and from its week start day: its days with data, time in range, daily dose and
  carbs, glucose every 5 minutes, and AGP (the 5th, 25th, 50th, 75th and 95th percentiles of the 15 minutes around
  each 5 minutes from midnight). A day's AGP is of the 7 days to it. A period not kept is built from the kept days in it, and today so
  far while it is not over; a day not kept is built from its data; and neither is kept. Any other id, or a date out of range, is not found.
- A summary period also has its glucose readings (`glucose_readings`) and those from a meter (`meter_readings`), their
  mean (`mean`, in the settings' units), standard deviation (`sd`), coefficient of variation (`cv`, the standard
  deviation as a percent of the mean) and GMI (`gmi`, 3.31 + 0.02392 × the mean in mg/dL), and the percent of them very
  low (`time_very_low`, below 54 mg/dL), low (`time_low`), in range, high (`time_high`) and very high
  (`time_very_high`, above 250 mg/dL), which add up to 100. A week's, fortnight's or month's readings are all its days',
  and its times are of all of them, so a day of few readings, such as today so far, counts for little. Its daily dose
  and carbs are the means of its days that are over, as today has had only some of its insulin and food, or of all its
  days with data when none is over. `cgm_active` is the percent of its 5 minutes, up to now, with a CGM reading.
- A summary period also has its daily dose's basal (`total_daily_basal`), the basal in force through a day, and a
  period's the mean of its days', as its daily dose is; the rest of the daily dose is boluses.
- A summary period also has the loop's IOB (`loop_iob`, `loop_iob_basal`, `loop_iob_treatment`) and COB (`loop_cob`)
  every 5 minutes from midnight: a day's newest loop device report in each 5 minutes, or else in the 5 minutes before,
  as one is kept only every 10 minutes once an hour old, null when there was none, and a
  week's, fortnight's or month's the mean of its days' at that time of day. A kept day keeps them after its device
  reports are truncated.
- The client also calls every Nightscout API v1, v2 and v3 endpoint on a site's Nightscout host, with a device's
  Nightscout token or the JWT it is exchanged for.

### Backups and limits

- S3 is off without `S3_URL`, with nothing restored on start or backed up. With it, the server tests uploading to S3
  at startup, again over about half a minute, and does not start if it never works, so a server that cannot reach
  its backups never starts empty and backs that up over them.
- A production or staging server with `S3_URL` will not start unless `S3_ACCESS_KEY` and `S3_SECRET_KEY` are set, and
  not to the development MinIO's `minioadmin`.
- Every site's DB is backed up to S3 in batches, and the control DB on its own, with old versions kept for a
  retention period.
- A site not written to since its last backup is not copied again, so its DB is not opened, and a batch with the same
  sites, none copied again, is not uploaded again until the server restarts; a write during a copy is copied the next
  round, and a failed upload is tried again the next round.
- A batch whose last site is deleted or archived is uploaded empty, so S3's latest backup of it no longer has the
  site's data and a restore brings back an empty batch; a batch that never had a site is not uploaded.
- A server with no control DB or site batch folder restores it from S3 on start; a download that fails or is cut off
  leaves nothing behind, so the next start downloads it again.
- A start removes the staging folders a restore or archive killed part way left beside the DBs and backups, and nothing
  else there.
- A site whose creation fails part way, such as its request being cancelled, leaves no site or identities behind.
- A site's DB that cannot be opened at start, as when its migration keeps failing, keeps its file and its backup, so
  a fixed deploy opens it with its data.
- A DB (control, site, jobs, logs or live) a migration failed on is forced back to the version before that migration
  when next opened, logged as a warning naming the DB and both versions, and migrated again; if the migration fails
  again, the error names the DB, the migration and why it failed.
- A live site whose DB file is missing on start, though its batch's folder is there, is restored from its batch's
  latest backup in S3, downloaded once per batch, and opened; either way it is a "Site DB missing" event the control
  admin sees, saying whether it was restored or why not (S3 not working, not in the backup, or the restore failed).
  A site restored that cannot be opened, as when its migration fails, stops the start, as a site that is there and
  cannot be opened does, so `update.sh` rolls back a release that cannot open it rather than keeping it without the site.
- A live site whose DB file is missing and not restored, such as one created after its batch was last backed up,
  stops the start, naming the site in the error, as a site that cannot be opened or migrated does: the server never
  serves, so takes no data, until every live site is opened and migrated, and `update.sh` rolls back a release that
  cannot. Its batch is not uploaded without it.
- A site keeps what each record was sent as only in its fields, never the JSON itself. What it needs of the JSON is
  kept with the record: a hash of a treatment's, to tell one sent again with changes (as Loop sends a dose again with
  what was delivered) from one sent the same, Loop's indefinite duration, and a device status's pump and uploader, for
  the fields without a column that Nightguard shows (AAPS's extended).
- Every 10 minutes every site is truncated: a device status's loop status, most of a site's DB with AAPS uploading
  every 5 minutes, is kept only on its device's newest status with one, which followers (LoopFollow, the Nightscout
  web app) draw their forecast and reason from; a newer status without one, as a pump's own, leaves it so. Once a
  newer one comes in, the others are cleared: their IOB and COB, extracted when they were written, are kept, and the
  Nightscout API gives them back as their `openaps` or `loop` object. The job looks at the statuses written in the last
  day; truncating a site looks at them all, so it clears those the job no longer looks at, as a device's newest before
  it went a day without reporting.
- A device status's `pump` as sent, the most of it after its loop status, is truncated the same way: kept only on its
  device's newest status with one, the others keeping their pump's battery, reservoir, clock, manufacturer, model,
  `pumpID`, `suspended` and `bolusing`, which the Nightscout API gives back as their `pump`.
- Truncating also thins each device's statuses: all of the last hour are kept, then only the newest in each 10
  minutes until a week old, in each 20 minutes until two weeks old, and in each hour until 30 days old, which keeps its
  newest. A deleted status is not removed, so a client syncing deletions sees it. Every status over 30 days old, its
  device's newest or deleted too, is deleted. The job also thins again the statuses that turned a week or two weeks old
  in the last day.
- Reads and writes are rate limited per site, identity and IP, and live sessions per site are capped.
- A Nightscout request that fails auth, or needs none, spends its IP's budget on the site
  (`SITE_UNAUTH_PER_MINUTE_LIMIT`); one with a valid token does not, so a device uploading often from one IP is limited
  only by its site and identity. An IP that has spent it is refused before its token is checked.
- Past the requests the server can work on at once (`REQUESTS_IN_FLIGHT_PER_CPU`) and a short queue
  (`REQUESTS_BACKLOG_PER_CPU`, waiting up to 5 seconds), a request is refused at once with 503 and `Retry-After`, so a
  server that falls behind answers what it can rather than every request timing out. The health check, live updates
  and Socket.IO connections are never refused this way. They are known by their route (`/health`, `/socket.io/`, a
  collection's `/<section>/<collection>/events`), so any other request sent with `Accept: text/event-stream` is still
  refused when the server is busy and still times out after 30 seconds. A quarter of each is kept for uploads to the
  Nightscout API, so glucose keeps arriving while page loads and reads fill the rest.
- A request for a site there is not is 404 with `{"status": 404, "message": "site does not exist"}` on its pages and its
  API, and 410 Gone with `{"status": 410, "message": "site does not exist"}` on its Nightscout API, so an app can tell
  its user the site is gone rather than a path is wrong, but one whose site cannot be looked up, as when the server is
  too busy to answer in time, is 503 with `Retry-After`, so an uploader tries again rather than take its URL as wrong.
- A disabled, archived or deleted site is answered just as a site that never was, so the answer tells nothing about it.
- No request returns more than 1000 records, however many it asks for, through the site's pages and API, the
  Nightscout API or its live updates. A v1 upload holds at most 5000, as Juggluco and Loop send their backlog in one
  batch, each 100 past the first costing a write against the site's and identity's limits; a live update's upload
  holds at most 100. Deleting a type of treatment or entry
  through the Nightscout API deletes at most 1000 each time, the newest first, saying how many.
- A DB serves many reads at once, so one slow read, such as a long search, does not hold up the others; writes to it
  take turns rather than fail. A busy DB keeps its connections open between reads, so requests at once do not each
  open a new one, while a quiet site's DB closes all of them.
- An IP's rate limits, on requests, sign ins, recoveries and site creation, cover its whole IPv6 /64, so stepping
  through the addresses of one /64 does not get fresh limits.
- An IP is the request's TCP peer, unless that is a trusted proxy (`TRUSTED_PROXIES`): then it is the hop before it in
  `X-Forwarded-For` and, on Fly.io, its `Fly-Client-IP`. Any other peer setting either header keeps its own IP's limits.
- A request body larger than `MAX_REQUEST_BODY_BYTES` is refused with 413, or, when it does not say its length,
  stops being read there and is refused as a bad body.
- A request whose headers hold more than 64 KiB, far above a browser's cookies or an uploader's token, is refused with
  431, so no connection holds more of them.
- Every response, from every host, has `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, a
  `Content-Security-Policy` with `frame-ancestors 'none'`, so no page, another site's among them, can frame it, and
  `Referrer-Policy: same-origin`, so no other origin is sent its URL; with `HOST_PROTOCOL=https` it also has
  `Strict-Transport-Security: max-age=31536000; includeSubDomains`.

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
- An admin adds a clinician, such as the person with diabetes' doctor, as a user with the Clinician role. It has a
  password, but no TOTP or Nightscout token, and its only page is the printable report: signed in, it opens the report
  with Print and Sign out buttons, and no header, menus or tabs. Any other page opens the report instead, and anything
  else, as the site's data, its own settings, its password or the Nightscout API, is refused.
- A site has at most `MAX_IDENTITIES_PER_USER` identities.

### Signing in

- The admin signs in with a password, then a TOTP code; users with a password.
- "Log in with a passkey" on the login page signs in the identity whose passkey the browser has, with no name,
  password or TOTP code: a passkey is two factors, the device and the person's fingerprint, face or PIN. A passkey only
  works on its own site's address. It keeps only what checking a sign in needs: its ID, public key, flags and sign
  count, never the authenticator's model or attestation. A disabled identity's passkey does not sign in, and passkey
  sign ins are rate limited per IP, and a site's from every IP at a far higher rate.
- Each passkey sign in or passkey added, on a site or the control site, finishes once only: its challenge is kept on
  the server for 5 minutes and removed when it is finished, so sending the same finish again is refused.
- The TOTP step only accepts the pre-auth token from this site's own password step.
- Every TOTP code field, on sign in, setup and the new site's code check, takes only digits; letters typed or pasted
  into it are dropped.
- The server refuses a TOTP code that is not six digits, like a wrong one.
- A TOTP code signs an identity in once, on the site (page or API) or in control: the same code, or an earlier one,
  is refused like a wrong one while it would still be valid, even after its owner signed in with it, though another
  identity with the same secret signs in with it. Checking a new site's code, or the setup page's, does not use it up.
  A dev or test server's fixed code `123123` is not one time.
- Sign-in attempts are rate limited per IP, and per name from every IP at a far higher rate
  (`SITE_LOGIN_NAME_PER_MINUTE_*`), so one IP sending wrong passwords for a name, or wrong passkeys or backup keys to
  a site, uses up only its own attempts and cannot lock anyone else out; an attempt one limit refuses spends none of the
  other's. One refused says "Too many login attempts. Wait a minute, then try again."; a disabled identity cannot sign in.
- A site or control login with a name that does not exist is refused with the same status and message as a wrong
  password, and checks a password all the same, so neither the answer nor how long it takes tells which names exist.
- The site login's inputs have no placeholders, and are marked invalid only after a failed attempt, until the next one.
- Devices use their Nightscout token, as a token, API secret or for a JWT.
- A Nightscout token whose name or role has no identity, or is for an admin or an identity with TOTP, is refused with
  the same status and message as a wrong secret, and checks a password all the same, so neither tells what is on the site.
- Changing an identity's password or TOTP signs out its existing sessions, and an admin's reset of it, on the site or in
  control, also removes its passkeys, as for a lost device, so it signs in only with its new credentials.
- Turning an identity off, deleting it or changing its password or TOTP also ends its open connections: its Socket.IO
  writes are refused at once and its sockets are disconnected within a minute, and its live update streams end within
  30 seconds.
- A page left open after its sign in ends goes to the login page once its live updates reconnect, rather than asking
  for them again and again.

### Users & Devices

- An admin adds users and devices, whose credentials are shown once, each with its own copy button and one, Copy all,
  to copy them all. Each copy button, and those in a device's setup steps, is named for what it copies, such as Copy
  Password, for a screen reader and as its tooltip, as it shows only an icon. Saying they are saved warns that they are never shown again, and they close only once that is confirmed;
  Go Back keeps them open. They are a dialog named for the identity that takes the focus, and Escape or pressing beside
  it shows that same warning rather than closing them; once a reset has signed you out, neither closes them, leaving
  Sign In Again.
- Adding a user starts on the User role, so one added without choosing a role is a User, not an Admin; the server
  refuses an identity sent with no role.
- Picking the role a select already has keeps it, rather than leaving the select empty.
- A device is added for its client: AAPS, Trio, Loop, xDrip+, xDrip4iOS, Nightguard, SugarPixel, T1D Carer or Other.
  Its credentials say how to set that client up, in numbered steps naming the client's menus and fields in bold, with its
  Nightscout URL and token filled in, each with its own copy button, and a device for any but Other has its own units
  set to mg/dL. One for xDrip4iOS or AAPS also has the Nightscout API write mg/dL as `mg/dl`, as xDrip4iOS reads a BG
  check's glucose as mg/dL only when its units are exactly that, and AAPS writes them so.
- The client is picked before the role, as a loop's device (AAPS, Trio or Loop) must be the PWD phone: picking one of
  them sets the role to PWD Phone and locks it, and the server refuses such a device with any other role.
- A device keeps the client it was added for, e.g. `loop` or `other`, given back as its `client`, so its settings can
  later be changed for every device of a client; a user and a site's first devices have none.
- On a phone, the Add User and Add Device forms wrap in their cards, so every input and button can be seen and tapped
  without scrolling sideways.
- An admin can turn an identity off and on with its row's switch. Its other actions, Settings, Reset password, Reset
  TOTP and Delete, are in its row's Manage menu, each labelled with its icon, Delete in red apart from the rest; those
  that do not apply, as resetting or deleting yourself, are greyed out. An identity cannot lock itself out.
- An admin can change another identity's own settings and alarms from its row, such as a device's units, which its
  Nightscout status gives followers to show glucose in.
- Another identity's glucose settings are labelled, shown, followed from the site's and typed in its own units, or the
  site's when it has none, not the admin's.
- Each identity lists the clients (browsers and apps) that have used it, and is marked shared when several have.
- Identity creation is rate limited, as each hashes a password.

### Resetting credentials

- An admin resets a device's password, which comes with its new Nightscout token.
- An identity resets its own password from the settings page, and the admin its own TOTP, which signs them out.
- Each reset is recorded as an event, with who made it.
- A lost admin login is recovered with the site's backup key on the recover page.

### My Account

- My Account, in the user menu, says who is signed in and with which role, opens the guide to setting up the
  site (not on control's), and lists its
  passkeys: when each was added and last used. Only an identity that signs in to the website has it.
- "Add a passkey" makes one on the site's own address, with the browser's fingerprint, face or PIN check, named for
  the identity (e.g. admin), so no name, email or device of the person is asked for or kept. A passkey the browser
  already has for the identity is not added again, and an identity has at most 10.
- "Remove" (after asking) removes one of the identity's own passkeys, as for a lost device, so it signs in no more.

### Usage

- Usage (`/usage`), in the user menu, is an admin's page of the site's DB size, measured as it opens, against its
  limit, and where its data builds up: a bar and a key of its largest tables, each with its records, size and share,
  the rest as Other, as last measured, as the control site's page of the site does. Others have no Usage, and My
  Account and the General settings do not show it.
- Its Truncate, after asking, truncates the site as control's Truncate does: the loop's status and pump as sent are
  cleared from all but each device's newest report, device reports are thinned and those over 30 days old deleted,
  loop settings no profile or device status refers to any more are deleted, but each app's last seen, and the DB is
  shrunk. Everything else is kept, and the card says so. It shows the storage again, with a toast of
  the DB's new size, and records a "Site truncated" event with the admin. It shares the site's few downloads a minute.
- Its Download Data downloads the site's data: a zip, `<site>-<date>.zip`, of a snapshot of its DB,
  `<site>-<date>.sqlite`, so more files can be added beside it later. It shares the site's few downloads a minute.

### Own settings and alarms

- Each identity can set its own units, time zone, language and formats over the site's.
- Each identity has its own alarms, and snoozing one only quiets it for them.
- Streaming the changes to an identity's settings (`/settings/identity/events`) or alarm settings
  (`/settings/alarms/events`) sends only its own, and to an admin the site's too; to anyone else a change to the site's,
  which its own inherit, is sent as `{}`, so it reads them again without being shown the site's. The site's
  (`/settings/site/events`) and other identities' (`/settings/identity-settings/events`) stream only to an admin.
- Each identity, a device or user of any role, registers the apps on its devices that pushes are sent to
  (`/settings/push-registrations`, JSON API): the app (`loop`, `trio`, `loopfollow`, `t1dcarer`), its provider (`apns`
  or `fcm`), bundle ID, whether it is a production build, and its token, which is kept encrypted and never returned.
  Registering a token again updates it, and from another of the site's identities moves it there; another site
  registering the same token has its own, so no site can take or change another's. An APNs token is 64 to 256 hex
  characters, the same in upper or lower case. An identity has at most 10, even from profiles uploaded at once, though
  it can always register one of them again, and sees, streams the changes to (`/settings/push-registrations/events`)
  and deletes only its own; an admin can also get and delete any of the site's, but its search and stream are still
  only its own.
- Settings, Push Devices (`/settings/push-devices`) lists the apps registered for pushes, with whose they are, the app
  and its service, its bundle ID, whether it is a development build, when it registered, and its last push and error;
  never its token. An admin sees every identity's on the site, and anyone else only their own. Each has a Test button,
  which sends it a test push and shows its new last push, or a toast with why it failed, and a Delete button, which
  after a confirmation removes its row with a "Push Device Deleted" toast, showing the table is empty again after the
  last; pushes to it stop until its app registers again. Anyone but an admin is refused (404) another identity's.
- An identity can send a test push to one of its registrations, and an admin to any of the site's
  (`POST /settings/push-registrations/{id}/test`), with its site's push credential for the app, else the server's.
  Without one it is refused, saying to add one; a token Apple says is gone, or Firebase says is `UNREGISTERED`,
  deletes the registration (410); another refusal says Apple's or Firebase's reason (502), and one that cannot reach
  them why, never with the device token. Each push is
  recorded on the registration, with its error, if it had one, which the same registration again, as each upload of
  a Loop or Trio profile is, keeps, and a token Apple says is gone is deleted from its own site alone. A site sends at most `SITE_PUSH_PER_MINUTE_LIMIT` pushes a minute.
- A Loop remote command (`POST /api/v2/notifications/loop`: `Temporary Override`, `Temporary Override Cancel`,
  `Remote Carbs Entry` or `Remote Bolus Entry`, with its `otp`), from an identity that can write treatments, is pushed
  as Nightscout pushes it to the Loop that last uploaded its profile, with the site's push credential for its bundle
  ID, expiring in 5 minutes, and is answered 200 `OK`. Who sent it is its identity, not its IP address. A bolus over
  30 U, carbs over 250 g or absorbing over 24 hours, an override over a week or without a name, or another event type
  is refused (400), as is a command to a site with no Loop or no credential for it, saying which.
- An FCM push is sent with the Firebase service account's OAuth token, got with a JWT its key signs from Google's
  token endpoint, never the `token_uri` in its JSON, and used again until five minutes before it expires; the key and
  tokens are never logged or shown. A service account Google refuses says Google's reason (502).
