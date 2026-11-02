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

### Limiting a server to some countries

- With `ALLOWED_COUNTRIES` set, only addresses in those countries can create sites, sign in (to sites and the control
  site) and recover; others are refused with "Nighttroop is not available in your country". Addresses whose country
  is not known are allowed.
- Signed in pages and the Nightscout API are not limited, so a phone keeps uploading while its person travels.

### Signing in

- A control admin signs in with a username, password and TOTP code, and can sign out. Tab goes from the username to the
  password to the code, skipping the button that shows the password.
- The server's global rate limit does not stop a control admin signing in, or using the control site once signed in,
  so they can fix whatever is using it up. The control site has its own limit at the same rate, and each address's and
  sign in attempts' limits still apply.
- A banner on every control page says when S3 backups or Sentry error reporting are not connected.
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

### Logs

- Every site's and the server's events, newest first, with readable names, when each started and how long it took.
- Events can be filtered by type and searched (ignoring case); errors and logs have their own pages.
- Every log the server writes is kept in the logs too, written in batches, without IP addresses. Logs, events and
  errors keep the `request_id` or `job_id` they were for. Logs held for writing are capped at 1 MB; past that they are
  dropped, and the logs then say how many.
- Logs, events and errors can be narrowed to one site, request or background job run by its ID. An event's, error's
  or log's details link to the rows of its request and job.
- Each sign in to the control site, and each refused one, is an event with the country (and city, where the server
  has DB-IP's City Lite) its IP address is estimated to be from. The IP address itself is not kept.

### Usage

- Shows active sites and users, use since start, the server's resources and the most used routes, refreshing every
  10 seconds.
- Counts the sites created today, this week and this month.
- Its memory is all the server's, as the operating system counts it (htop's RES), with SQLite's, which Go's own figures
  leave out, and says how much of it is Go's heap.
- Its CPU is the server's in the last 5 seconds as htop shows it, 100% for each core used, so up to 400% on 4 cores.
- Only signed in requests are counted.

### Rate limits

- The rate limits page lists each kind of rate limiter (every request, the control site's signed out, an IP's, logins,
  sites created, a site's and an identity's reads and writes, sync changes, identities created, downloads), refreshing
  every 10 seconds.
- Each shows what one is kept for (server, IP, site, identity, login name or IP on a site), its rate a minute and burst,
  how many are kept, how much of its burst the most used one has spent, and how many would refuse a request now.
- The most used turns amber once it has spent half its burst, and red while any would refuse.
- Reading them spends nothing, and names no site, identity or IP.

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
  code check. The login button is disabled until a code from the authenticator app matches; a wrong code says so. The
  check changes nothing, and a recovered site's credentials are checked the same way.
- A new site takes the time zone and language of the browser that created it, and its country's units and formats.
- Site creation is rate limited per IP and server wide, and refused once the server holds `MAX_SITES`.
- While a control admin has turned public site creation off, the home page says so instead of offering to create a
  site, and creating one is refused.
- The public **recover page** takes a site's ID and backup key and gives the admin a new password and TOTP and the
  site a new backup key. The key works once; a wrong key is refused without saying why.
- A site that is never used is deleted after `DELETE_AFTER_INACTIVITY_MINS`.
- Every page's footer links to contact, the terms of service and the privacy policy.
- With `FEEDBACK_URL` set, e.g. to a Google Form or GitHub issues, every page's footer also links to "feedback", and the
  home page asks for it under its features, each opening the link in a new tab; unset, neither shows.
- The public pages (create, recover, terms and privacy) say above their footer that Nighttroop is not a medical device,
  gives no medical advice, and its data and alarms may be late, missing or wrong.
- A server set up as a test site (`TEST_SITE`) says so in a large banner at the top of every page, public, site and
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
- The summary's glucose readings link to the page of every reading, and each treatment's kind (e.g. Insulin) to the
  page of that kind.
- The summary updates live as new data arrives, and catches up after the page has been in the background.
- A new site's empty summary links to the sync settings.

### Data

Each kind of data has a page with a table, newest first and paged, where records can be added, viewed, edited and
deleted (after a confirmation):

- **Glucose**: CGM and meter readings with trend arrows, in the settings' units (mmol/L or mg/dL), with lows and
  highs coloured as on the summary.
  - Adding a reading offers its type as Meter or Sensor and its trend as an arrow with its name (e.g. `↗ FortyFiveUp`),
    as the table and its details show them.
- **Insulin doses**
- **Food intakes**
- **Temp basals**, with a percent as the percent of the scheduled basal, 50 for half, never negative
- **Temp targets**, in the settings' units
- **Profile switches**, with their profile's schedules to read
- **Activities**
- **Site and sensor changes**
- **Notifications**
- **Event groups**: the treatments that came in together through the Nightscout API, with their events.
  - An event group can also be added on the site: a Nightscout type (e.g. Meal Bolus), a time, and up to one event of
    each kind (glucose, insulin, food, temp basal, temp target, profile switch, site or sensor change, activity and
    notification), picked from those within an hour of the time. Changing the time lists the events near it.
  - Clicking a group edits its type, time, notes and events; an event taken out of a group is in no group.
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
- A failing source syncs less often, and says when it failed and why.
- A working source can be backfilled with the days before its oldest reading.
- The sync page shows how each source has synced, how often, its backfill and its recent events.
- Sync credentials are encrypted at rest.

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
- **v2**: authorization (a token for a JWT, and a check of one of its permissions), properties and summary.
- **v3**: create, read, update, patch, delete, search and history of each collection, plus version, status,
  lastModified and settings.
- The v3 status shows what the token may do to each collection, e.g. `crud` for an uploader and `r` for a read only
  token, so apps like AAPS only upload with a token that may write.
- A v3 history includes deleted documents, with `isValid` false, so apps learn of deletions; a search leaves them
  out.
- A delete moves a document's `srvModified` and the v3 lastModified forward, so a history since an app's last sync
  has the documents deleted after it.
- **Socket.IO** live updates and alarms over long polling or websockets.
- A new Announcement is sent to every `/alarm` socket that can read, with its notes as the message, so AAPS shows it
  as a notification.
- Glucose is stored in mg/dL and returned and accepted in the settings' units.
- The v1 status has the settings' units in `settings.units`, as Nightscout writes them (`mg/dl` or `mmol`).
- The v3 version needs no token, so an app can check for v3 before it signs in.
- An entry has an `sgv` if it is a CGM reading and an `mbg` if it is a meter reading, never both.
- A date filter takes milliseconds with a fraction (e.g. `date$gte=1761234567890.25`), which is dropped.
- A `created_at` filter or sort is on when it happened, as `date` is, and `srvCreated` is on when it was uploaded.
- v3 filters take `field$eq` as well as `field`, e.g. `eventType$eq=Site Change`.
- A device status keeps its pump's other fields, e.g. AAPS's `extended` with its profile and temp basal.
- A temp basal sent with only a `percent` of the basal (e.g. 50 for half, 150 for 1.5x) has its rate set from the
  default profile's or profile switch's basal at its start, in the profile's timezone, and is returned with that
  `absolute` and its `percent` as sent.
- Duplicate uploads are ignored; values far outside real use are refused.
- A v1 upload of a document with an `_id` replaces the one with that `_id`, or is created with it, as Loop sends an
  override again with its `_id` when it changes, and deletes it by that `_id`.
- Loop's `Temporary Override` is kept as a temp target, with its correction range and insulin needs scale factor; an
  indefinite one lasts a week.
- An identifier over 36 characters, like an Omnipod dose's `syncIdentifier`, is kept as a hash of it, so two that
  start the same are not taken as one.
- v1 `profile/current` is the newest profile set itself, not a list of one, with `enteredBy`, as Loop reads it.
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
- A profile has `enteredBy`, the identity that uploaded it, as Trio needs it to import one.
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
- A live site whose DB file is missing on start, though its batch's folder is there, is restored from its batch's
  latest backup in S3, downloaded once per batch, and opened; either way it is a "Site DB missing" event the control
  admin sees, saying whether it was restored or why not (S3 not working, not in the backup, or the restore failed).
- A site whose DB file is missing and not restored, such as one created after its batch was last backed up, is left
  out, so the server still starts and the rest of its batch is still backed up.
- Old records' raw JSON is cleared after `TRUNCATE_RECORDS_AFTER_DAYS`.
- Reads and writes are rate limited per site, identity and IP, and live sessions per site are capped.
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

- An admin adds users and devices, whose credentials are shown once.
- A device is added for its client: AAPS, Trio, Loop, xDrip+, Nightguard, SugarPixel or Other. Its credentials say how
  to set that client up, with its Nightscout URL and token filled in, and a device for any but Other has its own units
  set to mg/dL, which those clients need, whatever the site's.
- An admin can turn an identity off and on, or delete it; an identity cannot lock itself out.
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
