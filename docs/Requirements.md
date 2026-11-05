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
- The server prints a one-time setup code to its logs; the setup page asks for it, so only whoever runs the server can
  set it up. Wrong codes are rate limited.
- The setup page shows the control admin's username, password and TOTP secret (with a QR code) once, and saves the
  admin only when a code from that TOTP secret is entered.
- A staging server (`MODE=staging`) is never set up: it starts with the control admin `admin` of its
  `STAGING_ADMIN_PASSWORD` and `STAGING_ADMIN_TOTP_SECRET`, made again whenever its data is wiped, and will not start
  without them. Otherwise it runs as production does, its site admins getting their own random credentials.

### Limiting a server to some countries

- With `ALLOWED_COUNTRIES` set, only addresses in those countries can create sites, sign in (to sites and the control
  site) and recover; others are refused with "Nighttroop is not available in your country". Addresses whose country
  is not known are allowed.
- Signed in pages and the Nightscout API are not limited, so a phone keeps uploading while its person travels.

### Signing in

- A control admin signs in with a username, password and TOTP code, and can sign out. Tab goes from the username to the
  password to the code, skipping the button that shows the password.
- The control site's header, with its menu, and its status banner show only once the admin has signed in, not on the
  login or setup page.
- The server's global rate limit does not stop a control admin signing in, or using the control site once signed in,
  so they can fix whatever is using it up. The control site has its own limit at the same rate, and each address's and
  sign in attempts' limits still apply.
- A banner on every control page says when S3 backups or Sentry error reporting are not connected.
- Errors reported to Sentry are under the server's `DEPLOYMENT` (`development` unless set) as their environment, so each
  of several servers, e.g. staging ones, has its own; a staging server on Fly.io is named after its app.
- The banner also says when the server has 90% or more of the sites `MAX_SITES` allows, and when it has them all.
- Every control page names the browser's tab after itself, e.g. `Sites · Nighttroop Control`.

### Sites

- The dashboard lists every site, newest first, with its state and DB size.
- Sites can be ordered by how long since they were last written to (the first to archive) or by DB size.
- A site not written to or seen within `DISABLE_AFTER_INACTIVITY_DAYS` is marked inactive.
- Clicking a site shows its details.
- An admin can create a site at an ID they pick (8 lowercase letters), e.g. a vanity one; a taken ID is refused. Its
  credentials are shown once. An admin can create sites past `MAX_SITES` and while public site creation is off.
- An admin can turn public site creation off, so sites cannot be created from the home page, and back on.
- A site can be:
  - **disabled**, after a confirmation, so it is no longer served;
  - **archived** once disabled, uploading it to S3 (needs S3);
  - **hydrated** from the archived sites page, bringing it back live;
  - **truncated** when over its DB size limit, clearing old records' raw JSON.
- A site deleted for good, such as one never used after it was made, has its identities deleted, with their passwords,
  TOTP secrets and devices' push tokens, and its push credentials' secrets removed; an archived site keeps them, as it
  can be hydrated.

### A site's identities

- A site's details list its identities: who they are, whether they use TOTP, and when each was last seen. Secrets
  are never shown.
- An admin can reset an identity's password (shown once, with a device's new Nightscout token) or TOTP secret. Each
  reset signs out the identity's existing sessions and is recorded as an event.

### Background jobs

- The jobs page lists the background jobs, newest first, with their duration; clicking one shows its details and
  errors.
- The job stats page summarises each day's jobs, newest day first, with its totals and errors.
- A failed job, and each day's job stats, link to the logs of its failed runs (the day's last 16), narrowed to the
  run's `job_id`, to see why it failed.
- Every run of a cron or queued job logs its name and a `job_id` of its own (a queued one also its `queued_job_id`),
  and its work for a site also the `site_id`, so one run's logs can be found together.

### Sync sources

- Lists every site's sync sources, the most recently attempted first, with failed counts in red.
- Clicking a source shows its sync reports.

### Push devices

- Lists every site's apps registered for pushes (`/push-devices`), newest first, at most 500, or one site's with
  `?site=`: its site, identity, app and service, bundle ID, whether it is a development build, when it registered, and
  its last push and error, never its token.
- A control admin can send any of them a test push with its Test button (`POST /push-devices/{id}/test`), within its
  site's push limit, seeing its new last push or why it failed; a read only control identity sees no Test buttons and
  is refused (403).

### Logs

- Every site's and the server's events, newest first, with readable names, when each started and how long it took.
- Events can be filtered by type and searched (ignoring case); errors and logs have their own pages.
- Every log the server writes is kept in the logs too, written in batches, without IP addresses. Logs, events and
  errors keep the `request_id` or `job_id` they were for. Logs held for writing are capped at 1 MB; past that they are
  dropped, and the logs then say how many.
- Logs, events and errors can be narrowed to one site, request or background job run by its ID. An event's, error's
  or log's details link to the rows of its request and job.
- Every DB query is timed, until its rows are read, and one that takes `SLOW_QUERY_MS` (100 ms) or longer is logged as
  a `Slow query` with its DB, its query's name, how long it took and how many rows it gave, but never its values, and
  the request, site and identity it was for. A write transaction waiting for the write lock is logged as `BEGIN`.
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
  sites created, a site's and an identity's reads and writes, sync changes, identities created, downloads), refreshing
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

- Anyone can create a site in one click on the public home page, without an account. It is shown once: its URL, site
  ID, admin password, the phone's Nightscout URL, and the backup key, with a prominent button to copy them all for a
  password manager. The copy has every first identity's credentials: the admin's, and the phone's and the read only
  device's Nightscout tokens. The TOTP secret and the phone's token are not shown as text: the secret is in the copy,
  and the token has its own copy button.
- The new site's admin login is its own step: its username and password, its TOTP secret only as a QR code, and a
  code check. The login button is disabled until a code from the authenticator app matches; a wrong code says so. A
  code is checked as soon as its sixth digit is in, with no click. The check changes nothing, and a recovered site's
  credentials are checked the same way.
- A new site takes the time zone and language of the browser that created it, and its country's units and formats.
- Site creation is rate limited per IP and server wide, and refused once the server holds `MAX_SITES`.
- While a control admin has turned public site creation off, the home page says so instead of offering to create a
  site, and creating one is refused.
- The public **recover page** takes a site's ID and backup key and gives the admin a new password and TOTP and the
  site a new backup key. The key works once; a wrong key is refused without saying why.
- A site that is never used is deleted after `DELETE_AFTER_INACTIVITY_MINS`.
- Every page's footer links to contact, the terms of service, the privacy policy and the API docs, and each footer link
  opens in a new tab, so the page the user was on stays open.
- The API docs, at `/api-docs` on every host, show the OpenAPI schema, `/openapi.json`, with Redoc in the site's
  colours: a menu and search of every endpoint, each with how it is signed in to, its parameters, body, responses and
  samples of them, and a link to download the schema. The page loads no images from anywhere else.
- With `FEEDBACK_URL` set, e.g. to a Google Form or GitHub issues, every page's footer also links to "feedback", and the
  home page asks for it under its features, each opening the link in a new tab; unset, neither shows.
- The public pages (create, recover, terms and privacy) say above their footer that Nighttroop is not a medical device,
  gives no medical advice, and its data and alarms may be late, missing or wrong.
- The create site page, its medical note and footer included, and the site and control login pages fit a laptop's or
  a phone's window, so none of them needs scrolling.
- A staging server (`MODE=staging`) says it is a test site in a large banner at the top of every page, public, site and
  control, warning that its data is frequently wiped.
- Every page names the browser's tab after itself, e.g. `Glucose Readings · Nighttroop`, whether it is loaded in full or
  opened from a menu; a page with no title of its own, such as the login, is named `Nighttroop`.
- A page that shows estimated locations, such as the events pages, also links to DB-IP, whose free databases make
  the estimates, in its footer.

### Summary

- The home page summarises the site: the latest glucose with how long ago it was, the loop's IOB and COB, every
  treatment, and each device's latest status.
- A graph of glucose with carbs and insulin marked, over 6 hours to a week.
- Time in range, total daily dose, and the last sensor and site change for today, yesterday, last week and last
  month.
- Readings are coloured and labelled low, in range or high by the alarm settings.
- The summary says how long ago things were in words, to the nearest hour once over an hour, as "45 minutes ago" or
  "2 days 1 hour ago", and its tiles and devices show units smaller and grey beside their numbers.
- The summary's glucose readings link to the page of every reading, and each treatment's kind (e.g. Insulin) to the
  page of that kind.
- The summary updates live as new data arrives, and catches up after the page has been in the background. While it
  does, ● Live shows beside its title.
- A summary with no readings in its time links to the sync settings.

### Getting started

- A site with no glucose for over a month, or ever, shows a guide to setting it up in place of its empty summary, until
  readings arrive, when the summary shows them live.
- The guide's steps each open the page they are done on: an admin gets their glucose in (the site's Nightscout URL for
  AAPS, Trio or xDrip+, or a sync from Dexcom, LibreLinkUp or Nightscout) and adds people and devices; everyone makes
  their settings and sets up alarms.
- Someone who is not an admin is told the admin sets up the site's data and who can use it.
- The guide is always at Getting Started, first in the menu.

### Data

Each kind of data has a page with a table, newest first and paged, where records can be added, viewed, edited and
deleted (after a confirmation). On a phone, a table wider than the window scrolls sideways in its card, so the page
does not:

- **Glucose**: CGM and meter readings with trend arrows, in the settings' units (mmol/L or mg/dL), with lows and
  highs coloured as on the summary.
  - Adding a reading offers its type as Meter or Sensor and its trend as an arrow with its name (e.g. `↗ FortyFiveUp`),
    as the table and its details show them.
  - Above the first page, the last 10 days' glucose are small graphs on the same scale, oldest first, each with the
    range shaded and its time in range; the last is today so far, from its clean readings.
- **Insulin doses**, with a combo bolus's split as the fractions given at once and over time, 0.7 for 70%
- **Food intakes**
- **Temp basals**, with the rate given in U/hr and a multiplier of the scheduled basal, 0.5 for half, 0 a zero temp
- **Temp targets**, in the settings' units
- **Profile switches**, with their profile's schedules to read, and a multiplier of its insulin, 1.2 for 120%
- **Activities**
- **Site and sensor changes**
- **Notifications**
- **Event groups**: the treatments that came in together through the Nightscout API, with their events.
  - An event group can also be added on the site: the Add Event Group menu lists the Nightscout types (e.g. Meal Bolus),
    and each opens a modal for that type with a time and up to one event of each kind that type can have (e.g. a Meal
    Bolus can have glucose, insulin, food and a site or sensor change, an Exercise only an activity), picked from those
    within an hour of the time. Changing the time lists the events near it.
  - Adding or editing a group with a kind of event its type cannot have, or of an unknown type, is refused.
  - Clicking a group edits its type, time, notes and events; changing its type offers only the kinds of event the new
    type can have. An event taken out of a group is in no group.
  - An event is in at most one group: events in another group are not offered, and grouping one again is refused.
  - A group added or edited on the site is a Nightscout treatment, with its events.
- **Device statuses**: each device's latest 3 statuses, and every status a device sent.

Every value written to a site, from its pages, the Nightscout API or a sync, has a broad limit far outside real use,
e.g. a bolus of at most 500 units or 1000 g of carbs, and a date from 1970 to the year 9999. Insulin and food can be
negative, to take back too much entered. The pages are stricter, e.g. a dose of at most 100 units either way.

Adding a record on a data page opens a modal of all the fields a person can type in, each saying what it is, in what
units, whether it is required, with its limits, and a dropdown or suggestions for its types. Its time is typed in the
settings' timezone, and is now when left blank; other fields left blank are not set.

### Presets

- **Food presets**: added in a modal of all their fields, as data is, viewed, edited and deleted.
- **Profiles**: profile stores uploaded by a loop, with their schedules to read, renamed or deleted.

### Sync

- A site can pull data from **Dexcom** (with its region), **LibreLinkUp** (with its region) or another
  **Nightscout** site, or from **virtual sync**, which makes realistic test data.
- Saving does not check the details: the first sync does, and turns syncing off and says why if it fails.
- Typing the details of a Dexcom, LibreLinkUp or Nightscout source not yet saved ticks its Enable Sync, as it is easy
  to forget; a click on the box, before or after, is kept, and a saved source's box is never changed.
- A failing source syncs less often, and says when it failed and why.
- A working source can be backfilled with the days before its oldest reading.
- The sync page shows how each source has synced, how often, its backfill and its recent events.
- The sync page updates as each sync or backfill finishes, showing its new errors, stats and events without reloading,
  leaving a form being edited as it is and an opened history open. While it does, ● Live shows beside its title.
- Sync credentials are encrypted at rest.
- A Nightscout sync on this server's host, or on a private, local or shared (carrier-grade NAT or Tailscale) address,
  is refused when saved and as it connects.

### Alarms

- An alarms page with urgent low, low, predicted low, high, urgent high, fast rise, fast drop, no data, low device
  battery and low reservoir alarms, each off by default.
- The site's alarms are the default for each identity, who can change them for themselves.
- An active alarm shows in a banner under the header on every page, live, and can be snoozed.
- Nonsense settings are refused, e.g. a high under the low.

### Settings

- General settings: units, time zone, language, date and time format, week start, insulin and carb defaults.
- An admin sets the site's settings; each identity can override them for themselves.
- An admin can **download** a snapshot of the site's DB.
- An admin can **delete** the site, after a confirmation, which archives it to S3 and signs everyone out.
- An admin can add, list and delete the site's **push credentials** on Settings, Push Notifications
  (`/settings/push-credentials`): an Apple APNs key (team ID, key ID and the `.p8` file pasted in) or a Firebase
  service account (project ID and its JSON), the apps it pushes to (`loop`, `trio`, `loopfollow`, `t1dcarer`) ticked,
  and, optionally, the only bundle IDs it pushes to, separated by commas. A key that cannot be read, or an unknown
  app, is refused with the modal left open; its secret is kept encrypted and never shown or returned; a site has at
  most 5; only the site's admins can see or delete them, and a delete asks first.

### Logs

- An admin sees the site's events (e.g. created, alarms, recoveries, credential changes), errors and logs, filtered
  by type and searched, and its batch's backups.
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
  status, versions, verifyauth, slice, count, times and echo, and experiments/test.
- v1 `experiments/test` is `{"status":"ok"}` for an identity that can upload readings and a 401 for any other, so
  Loop's and xDrip4iOS's connection tests pass with a PWD phone's token.
- A v1 search by `find[_id]` that finds nothing, such as a deleted treatment, is an empty list, not a 404.
- A search with a filter on a field or with an operator Nighttroop does not know, a v1 `find[...]` (e.g.
  `find[notes]=x`) or a v3 `field$operator` (e.g. `notes$eq=x`), is an empty list, as nothing would match in
  Nightscout. Any other parameter, such as xDrip+'s `rr`, is not a filter and is left out.
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
- A delete moves a document's `srvModified` and the v3 lastModified forward, so a history since an app's last sync
  has the documents deleted after it.
- **Socket.IO** live updates and alarms over long polling or websockets.
- A Nightscout food has `type` `food`, as AAPS loads only foods of that type.
- The Nightscout host's root, which LoopFollow's Nightscout tab opens, goes to the site's own pages.
- A live update sent to a site's sockets is in the site's own settings, such as its units, as it goes to every device.
- A new Announcement is sent to every `/alarm` socket that can read, with its notes as the message, so AAPS shows it
  as a notification.
- Glucose is stored in mg/dL and returned and accepted in the settings' units.
- The v1 status has the settings' units in `settings.units`, as Nightscout writes them (`mg/dl` or `mmol`).
- The settings `units_mgdl_label` and `units_mmol_label`, set only through the settings endpoint, say how the
  Nightscout API writes each units in a document's `units`: mg/dL as `mg/dL` or `mg/dl`, and mmol/L as `mmol/L` or
  `mmol`, Nighttroop's way by default. Units are stored as Nighttroop's, and any of these is read.
- The v1 status has `authorized` with a JWT and the token's permissions, with `api:treatments:create` for one that
  can write, so LoopFollow shows a read token as "Read" and a write token as "Read & Write".
- The v3 version needs no token, so an app can check for v3 before it signs in.
- An entry has an `sgv` if it is a CGM reading and an `mbg` if it is a meter reading, never both.
- A date filter takes milliseconds with a fraction (e.g. `date$gte=1761234567890.25`), which is dropped.
- A `created_at` filter or sort is on when it happened, as `date` is, and `srvCreated` is on when it was uploaded.
- v3 filters take `field$eq` as well as `field`, e.g. `eventType$eq=Site Change`.
- A device status keeps its pump's other fields, e.g. AAPS's `extended` with its profile and temp basal.
- A profile from Loop, with `loopSettings.deviceToken` and `bundleIdentifier`, or from Trio, with them at its top,
  registers that app's APNs token for pushes for the identity that uploaded it, a production build unless
  `isAPNSProduction` is false; a token that is not one is not registered, and the profile is still kept.
- A profile switch's `percentage` is kept as a multiplier of the profile's insulin (120 as 1.2) and returned as sent.
- A combo bolus's `splitNow` and `splitExt` are kept as the fractions given at once and over time (70 as 0.7) and
  returned as sent.
- A temp basal is sent with its rate in U/hr as `rate` or `absolute`, its `percent`, or both. Its `percent` is
  Nightscout's change from the basal, as AAPS sends it (e.g. -50 for half, 20 for 1.2x, -100 a zero temp, never under
  -100), kept as a multiplier of the basal (0.5, 1.2, 0). The one not sent is set from the default profile's or
  profile switch's basal at its start, in the profile's timezone; a rate over 11 times it keeps no percent.
- A temp basal is always returned with its `rate` and `absolute`, both its U/hr, and its `percent` when it has one.
- The summary's basal during a temp basal is its rate.
- Duplicate uploads are ignored; values far outside real use are refused.
- An entry uploaded without an `_id` with the date, type and device of one stored takes its new value, direction and
  raw values, as xDrip4iOS smooths readings, and followers get it as changed; one sent with the same values changes
  nothing, not even when it was last modified.
- A `created_at` filter given a time with no zone, as LoopFollow sends it, reads it as UTC.
- Loop's and Trio's Suspend Pump and Resume Pump treatments are kept.
- An activity's `type`, as xDrip+ sends it, is its activity type, a heart rate's `bpm` its heart rate and a motion's
  `class` its name.
- A device status keeps its pump's clock, manufacturer and model, its uploader's battery, returned in both
  `uploader.battery` and `uploaderBattery`, and Loop's override in force.
- A device status keeps Trio's pump `bolusIncrement`, LoopFollow's remote bolus step, while the status as sent is kept.
- A Nightscout treatment comes back with its `created_at` as `timestamp`, a bolus with its insulin as `programmed`,
  and a temp basal with `temp` `absolute`, as Loop's NightscoutKit drops treatments without them; a `timestamp`
  sent is not read.
- A device status keeps its uploader's `isCharging`, and, while the status as sent is kept, its uploader's `name` and
  `timestamp` and Trio's pump `status.bolusing`, `status.suspended` and battery `string` and `display`.
- A profile keeps Trio's remote control settings (bundle, device token, APNS environment, team and expiry) and its
  override presets, for LoopFollow's Trio Remote Control.
- A v1 upload of a document with an `_id` replaces the one with that `_id`, or is created with it, as Loop sends an
  override again with its `_id` when it changes, and deletes it by that `_id`.
- Loop's `Temporary Override` is kept as a temp target, with its correction range and insulin needs scale factor; an
  indefinite one lasts a week, and is given back with `durationType` `indefinite` while the override as sent is kept.
- An identifier over 36 characters, like an Omnipod dose's `syncIdentifier`, is kept as a hash of it, so two that
  start the same are not taken as one.
- v1 `profile/current` is the newest profile set itself, not a list of one, with `enteredBy`, as Loop reads it.
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
- A device's Nightscout token is also taken as its SHA1 in the `api-secret` header, as xDrip+ sends the token in its
  URL; a password reset stops the old token's SHA1 working.
- An xDrip+ upload works as it sends it (see `docs/clients/xdrip.md`): a calibration (`type` `cal`) in a batch of
  entries is left out and the readings with it kept; an uploader battery that is not a number is left out of its
  device status; `uploader.battery` is the uploader's battery; and a treatment is found by its `uuid` (`find[uuid]`), sent as the `_id` or as xDrip+ makes its `_id`.
- Trio's and iAPS's `SMB` and `External Insulin` treatments are kept as insulin doses, with their event type.
- xDrip+'s `Sensor Stop` treatments are kept as a site or sensor change, with their event type, so the treatments it
  uploads after one are kept too; the site can add a Sensor Stop change and a Sensor Stop event group.
- A profile has `enteredBy`, the identity that uploaded it, as Trio needs it to import one.
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
- A `dateString` with Juggluco's `-03:-30` zone for half hours west of UTC is read as `-03:30`.

### Nighttroop API and client

- Every site page is also a JSON API, with a JWT from signing in.
- A Go client, `clients/nighttroop`, is generated from its OpenAPI schema. It can create a site, sign in, read the
  summary, and list, read, add, edit and delete glucose readings, insulin doses and food intakes.
- The client also calls every Nightscout API v1, v2 and v3 endpoint on a site's Nightscout host, with a device's
  Nightscout token or the JWT it is exchanged for.

### Backups and limits

- S3 is off without `S3_URL`, with nothing restored on start or backed up. With it, the server tests uploading to S3
  at startup, again over about half a minute, and does not start if it never works, so a server that cannot reach
  its backups never starts empty and backs that up over them.
- Every site's DB is backed up to S3 in batches, and the control DB on its own, with old versions kept for a
  retention period.
- A site not written to since its last backup is not copied again, so its DB is not opened, and a batch with the same
  sites, none copied again, is not uploaded again until the server restarts; a write during a copy is copied the next
  round, and a failed upload is tried again the next round.
- A server with no control DB or site batch folder restores it from S3 on start; a download that fails or is cut off
  leaves nothing behind, so the next start downloads it again.
- A start removes the staging folders a restore or archive killed part way left beside the DBs and backups, and nothing
  else there.
- A site whose creation fails part way, such as its request being cancelled, leaves no site or identities behind.
- A site's DB that cannot be opened at start, as when its migration fails and leaves it dirty, keeps its file and its
  backup, so a fixed deploy opens it with its data.
- A live site whose DB file is missing on start, though its batch's folder is there, is restored from its batch's
  latest backup in S3, downloaded once per batch, and opened; either way it is a "Site DB missing" event the control
  admin sees, saying whether it was restored or why not (S3 not working, not in the backup, or the restore failed).
  A site restored that cannot be opened, as when its migration fails, stops the start, as a site that is there and
  cannot be opened does, so `update.sh` rolls back a release that cannot open it rather than keeping it without the site.
- A site whose DB file is missing and not restored, such as one created after its batch was last backed up, is left
  out, so the server still starts and the rest of its batch is still backed up.
- Old records' raw JSON is cleared after `TRUNCATE_RECORDS_AFTER_DAYS`.
- Reads and writes are rate limited per site, identity and IP, and live sessions per site are capped.
- Past the requests the server can work on at once (`REQUESTS_IN_FLIGHT_PER_CPU`) and a short queue
  (`REQUESTS_BACKLOG_PER_CPU`, waiting up to 5 seconds), a request is refused at once with 503 and `Retry-After`, so a
  server that falls behind answers what it can rather than every request timing out. The health check, live updates
  and Socket.IO connections are never refused this way.
- No request returns more than 1000 records, however many it asks for, through the site's pages and API, the
  Nightscout API or its live updates, and an upload holds at most 100. Deleting a type of treatment through the
  Nightscout API deletes at most 1000 each time, saying how many.
- A DB serves many reads at once, so one slow read, such as a long search, does not hold up the others; writes to it
  take turns rather than fail. A busy DB keeps its connections open between reads, so requests at once do not each
  open a new one, while a quiet site's DB closes all of them.
- An IP's rate limits, on requests, sign ins, recoveries and site creation, cover its whole IPv6 /64, so stepping
  through the addresses of one /64 does not get fresh limits.
- A request body larger than `MAX_REQUEST_BODY_BYTES` is refused with 413, or, when it does not say its length,
  stops being read there and is refused as a bad body.

## Identity

### Roles

| Role | Signs in to the site | Nightscout API | Changes data | Manages identities and sync |
| --- | --- | --- | --- | --- |
| Admin | Password and TOTP | No | Yes | Yes |
| User | Password | Yes | Yes | No |
| Read only user | Nightscout token | Yes | No | No |
| PWD phone (the loop's phone) | Nightscout token | Yes | Yes | No |
| Read and write device | Nightscout token | Yes | Yes | No |
| Read only device | Nightscout token | Yes | No | No |

- A new site has an admin, a PWD phone and a read only device.
- A site has at most `MAX_IDENTITIES_PER_USER` identities.

### Signing in

- The admin signs in with a password, then a TOTP code; users with a password.
- The TOTP step only accepts the pre-auth token from this site's own password step.
- Every TOTP code field, on sign in, setup and the new site's code check, takes only digits; letters typed or pasted
  into it are dropped.
- The server refuses a TOTP code that is not six digits, like a wrong one.
- Sign-in attempts are rate limited per identity and IP; a disabled identity cannot sign in.
- Devices use their Nightscout token, as a token, API secret or for a JWT.
- Changing an identity's password or TOTP signs out its existing sessions.
- Turning an identity off, deleting it or changing its password or TOTP also ends its open connections: its Socket.IO
  writes are refused at once and its sockets are disconnected within a minute, and its live update streams end within
  30 seconds.
- A page left open after its sign in ends goes to the login page once its live updates reconnect, rather than asking
  for them again and again.

### Users & Devices

- An admin adds users and devices, whose credentials are shown once, each with its own copy button and one to copy
  them all. Saying they are saved warns that they are never shown again, and they close only once that is confirmed;
  Go Back keeps them open.
- A device is added for its client: AAPS, Trio, Loop, xDrip+, xDrip4iOS, Nightguard, SugarPixel or Other. Its
  credentials say how to set that client up, in numbered steps naming the client's menus and fields in bold, with its
  Nightscout URL and token filled in, each with its own copy button, and a device for any but Other has its own units
  set to mg/dL, which those clients need, whatever the site's. One for xDrip4iOS also has the Nightscout API write
  mg/dL as `mg/dl`, as it reads a BG check's glucose as mg/dL only when its units are exactly that.
- The client is picked before the role, as a loop's device (AAPS, Trio or Loop) must be the PWD phone: picking one of
  them sets the role to PWD Phone and locks it, and the server refuses such a device with any other role.
- A device keeps the client it was added for, e.g. `loop` or `other`, given back as its `client`, so its settings can
  later be changed for every device of a client; a user and a site's first devices have none.
- On a phone, the Add User and Add Device forms wrap in their cards, so every input and button can be seen and tapped
  without scrolling sideways.
- An admin can turn an identity off and on, or delete it; an identity cannot lock itself out.
- An admin can change another identity's own settings and alarms from its row, such as a device's units for a client
  that needs mg/dL on a site in mmol/L; the Nightscout API gives that device its own units.
- Each identity lists the clients (browsers and apps) that have used it, and is marked shared when several have.
- Identity creation is rate limited, as each hashes a password.

### Resetting credentials

- An admin resets a device's password, which comes with its new Nightscout token.
- An identity resets its own password from the settings page, and the admin its own TOTP, which signs them out.
- Each reset is recorded as an event, with who made it.
- A lost admin login is recovered with the site's backup key on the recover page.

### Own settings and alarms

- Each identity can set its own units, time zone, language and formats over the site's.
- Each identity has its own alarms, and snoozing one only quiets it for them.
- Each identity, a device or user of any role, registers the apps on its devices that pushes are sent to
  (`/settings/push-registrations`, JSON API): the app (`loop`, `trio`, `loopfollow`, `t1dcarer`), its provider (`apns`
  or `fcm`), bundle ID, whether it is a production build, and its token, which is kept encrypted and never returned.
  Registering a token again updates it, and from another of the site's identities moves it there; another site
  registering the same token has its own, so no site can take or change another's. An APNs token is the same in upper
  or lower case. An identity has at most 10, though it can always register one of them again, and sees, streams the
  changes to (`/settings/push-registrations/events`) and deletes only its own.
- Settings, Push Devices (`/settings/push-devices`) lists the apps registered for pushes, with whose they are, the app
  and its service, its bundle ID, whether it is a development build, when it registered, and its last push and error;
  never its token. An admin sees every identity's on the site, and anyone else only their own. Each has a Test button,
  which sends it a test push and shows its new last push, or a toast with why it failed.
- An identity can send a test push to one of its registrations, and an admin to any of the site's
  (`POST /settings/push-registrations/{id}/test`), with its site's push credential for the app, else the server's.
  Without one it is refused, saying to add one; a token Apple says is gone deletes the registration (410); another
  refusal says Apple's reason (502), and one that cannot reach Apple why, never with the device token. Each push is
  recorded on the registration, with its error, if it had one, and a token Apple says is gone is deleted from its own
  site alone. A site sends at most `SITE_PUSH_PER_MINUTE_LIMIT` pushes a minute. Only APNs pushes are sent so far.
