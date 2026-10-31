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

- A control admin signs in with a username, password and TOTP code, and can sign out.
- A banner on every control page says when S3 backups or Sentry error reporting are not connected.

### Sites

- The dashboard lists every site, newest first, with its state and DB size.
- Sites can be ordered by how long since they were last written to (the first to archive) or by DB size.
- A site not written to or seen within `DISABLE_AFTER_INACTIVITY_DAYS` is marked inactive.
- Clicking a site shows its details.
- An admin can create a site at an ID they pick (8 lowercase letters), e.g. a vanity one; a taken ID is refused. Its
  credentials are shown once.
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

### Sync sources

- Lists every site's sync sources, the most recently attempted first, with failed counts in red.
- Clicking a source shows its sync reports.

### Logs

- Every site's and the server's events, newest first, with readable names, when each started and how long it took.
- Events can be filtered by type and searched (ignoring case); errors and logs have their own pages.
- Each sign in to the control site, and each refused one, is an event with the country (and city, where the server
  has DB-IP's City Lite) its IP address is estimated to be from. The IP address itself is not kept.

### Usage

- Shows active sites and users, use since start, the server's resources and the most used routes, refreshing every
  10 seconds.
- Counts the sites created today, this week and this month.
- Only signed in requests are counted.

### Resource alerts

- Every 5 minutes the server checks its memory (of its container's limit, or else the machine's), the space used on
  its DBs' disk, and how long its CPU stayed at 100%, sampled every 5 seconds.
- Memory or disk at or over 60% (`MEMORY_ALERT_PERCENT`, `DISK_ALERT_PERCENT`), or the CPU at 100% for 20 seconds or
  more (`CPU_PEGGED_ALERT_SECONDS`), shows in the banner on every control page until a check finds it back under.
- Going over is a "Resource high" event of the server's, and coming back under a "Resource recovered" one.
- The last 3 days of checks are kept while the server runs, and start again when it restarts. The usage page graphs
  them: mean CPU, and memory and DB disk used, with a dashed line at the level each warns at.

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
- The public **recover page** takes a site's ID and backup key and gives the admin a new password and TOTP and the
  site a new backup key. The key works once; a wrong key is refused without saying why.
- A site that is never used is deleted after `DELETE_AFTER_INACTIVITY_MINS`.
- Every page's footer links to contact, the terms of service and the privacy policy.
- A page that shows estimated locations, such as the events pages, also links to DB-IP, whose free databases make
  the estimates, in its footer.

### Summary

- The home page summarises the site: the latest glucose with how long ago it was, the loop's IOB and COB, every
  treatment, and each device's latest status.
- A graph of glucose with carbs and insulin marked, over 6 hours to a week.
- Time in range, total daily dose, and the last sensor and site change for today, yesterday, last week and last
  month.
- Readings are coloured and labelled low, in range or high by the alarm settings.
- The summary updates live as new data arrives, and catches up after the page has been in the background.
- A new site's empty summary links to the sync settings.

### Data

Each kind of data has a page with a table, newest first and paged, where records can be added, viewed, edited and
deleted (after a confirmation):

- **Glucose**: CGM and meter readings with trend arrows, in the settings' units (mmol/L or mg/dL), with lows and
  highs coloured as on the summary.
- **Insulin doses**
- **Food intakes**
- **Temp basals**
- **Temp targets**, in the settings' units
- **Profile switches**, with their profile's schedules to read
- **Activities**
- **Site and sensor changes**
- **Notifications**
- **Event groups**: the treatments that came in together through the Nightscout API, with their events.
- **Device statuses**: each device's latest 3 statuses, and every status a device sent.

### Presets

- **Food presets**: added, viewed, edited and deleted.
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
- Each sign in to the site, and each refused one, is an event with the country (and city, where the server has
  DB-IP's City Lite) its IP address is estimated to be from. A refused one does not keep the name it tried.

### Nightscout API

- Each site has a Nightscout compatible API at `<site>.ns.<host>` for apps like AAPS, Loop, Trio and xDrip, and
  followers.
- **v1**: entries, treatments, device statuses, profiles, food and activities, with Nightscout's filters, plus
  status, versions, verifyauth, slice, count, times and echo.
- **v2**: authorization (a token for a JWT), properties and summary.
- **v3**: create, read, update, patch, delete, search and history of each collection, plus version, status,
  lastModified and settings.
- **Socket.IO** live updates and alarms over long polling or websockets.
- Glucose is stored in mg/dL and returned and accepted in the settings' units.
- Duplicate uploads are ignored; values far outside real use are refused.

### Nighttroop API and client

- Every site page is also a JSON API, with a JWT from signing in.
- A Go client, `clients/nighttroop`, is generated from its OpenAPI schema. It can create a site, sign in, read the
  summary, and list, read, add, edit and delete glucose readings, insulin doses and food intakes.
- The client also calls every Nightscout API v1, v2 and v3 endpoint on a site's Nightscout host, with a device's
  Nightscout token or the JWT it is exchanged for.

### Backups and limits

- Every site's DB is backed up to S3 in batches, and the control DB on its own, with old versions kept for a
  retention period.
- Old records' raw JSON is cleared after `TRUNCATE_RECORDS_AFTER_DAYS`.
- Reads and writes are rate limited per site, identity and IP, and live sessions per site are capped.

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
- Sign-in attempts are rate limited per identity and IP; a disabled identity cannot sign in.
- Devices use their Nightscout token, as a token, API secret or for a JWT.
- Changing an identity's password or TOTP signs out its existing sessions.

### Users & Devices

- An admin adds users and devices, whose credentials are shown once.
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
