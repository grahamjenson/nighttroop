# TODOs

## Alpha Testing

1. Point real deivces at Loop, Trio, AAPS and xDrip+ uploading, follower apps (xDrip+ follower, LoopFollow,
   Nightguard) reading, and Dexcom and Libre connectors
1. Automated Testing of the staging, create site, do a bunch of stuff
1. FINISH announcement post
1. e2e code review. and flatten.

## Before Beta Testing

1. Reports/Derived Values:
    1. Time below and above range, as well as time in range
    1. Small multiples Daily
    1. P3: Keep the days waiting to be built again (`summary_daily.Builder.stale`) in the site's DB, as they are held in
       memory and lost when the server stops or deploys before it builds them; and mark the day a changed record moved
       from, as only the day it moved to is marked

1. Alarms:
    1. Push notifications to phones, as alarms now only reach an open site page or a Nightscout socket client

1. Bad Sites
    1. write entry compression
    1. write treatments compressions

1. Control UI
    1. Show bad users
    1. Block IP addresses
    1. Restore one site from its backup
    1. P4: Sign out a control admin whose password or TOTP another resets, as a reset site identity is; its control
       sign in, which lasts an hour, keeps working until then
    1. P4: Measure each site's DB tables in the resource check and keep them, as a site's page reads every page of its DB
       (SQLite's `dbstat`) to split its DB size bar, which a DB of hundreds of MB takes a while to do

1. Accounts
    1. Recovering a site's admin without the initial backup key
    1. Recovering a lost control admin login, as setup only runs on a server with no control admin

1. Security
    1. Testing deploy with rolling to a new keyset, moving to a new encrypted keyset.
    1. P3: Capture each Socket.IO message on staging, as `CAPTURE_REQUESTS` keeps a websocket's upgrade but not its
       messages
    1. P3: Refresh the built in DB-IP country database (`utils/geoip/dbip-country-lite.mmdb`) each month or two, as
       DB-IP updates it monthly; the Docker image's city database is fetched fresh on each build.
    1. P2: Add a `script-src` Content-Security-Policy with nonces, as only `frame-ancestors` is set now: the templ and
       htmx pages use inline scripts and handlers (`onclick`, `hx-on`) that it would need auditing for first
    1. P2: Stop many sites making the server poll one third party Nightscout, as only each site's connector saves are
       rate limited and its credentials' ciphertexts differ: keep a keyed HMAC (`SecurityConfig.mac`, the MAC key set
       in `KEY_SETS`) of the normalised Nightscout host in a new `site_connectors` column with a partial unique index,
       set on create and update and cleared on delete, and refuse a second site syncing that host with a 409 saying it
       is synced by another site. First decide how a caregiver's own site may still follow a Nightscout another site
       syncs, which this would refuse

## Before First Release:

1. Translations
1. Docs hosting
1. P3: Add event groups (list, get, add, edit, delete and the events to group) to `clients/nighttroop/openapi.json`
   and its client
1. Performance testing (cost estimate for running full node)
1. P4: Mark the user menu's button in the header while a page of Settings, Alarms, or Logs & Events is shown, as the
   header then marks no hub and only the open menu marks it
1. P4: Marshal an entry without `Entry.MarshalJSON` (`models/nightscoutmodels/entries.go`) marshalling it again as
   an alias and making a `time.FixedZone` for its `dateString`, about a quarter of the CPU of a 288 entry
   `/api/v1/entries` search now its fields are not read back as a map (`BenchmarkWriteRecordsV1Entries`)
1. P3: Open a site's DB when it is first used rather than every site's at start, as each still takes ~2ms to open
   and check migrations
1. P4: Keep which batches were uploaded, and with which sites, across restarts, as the first backup round after a
   restart uploads every batch again though their backups are current
1. Migrations (from the audit of what happens when one fails):
    1. P0: At the first production deploy, freeze the migrations: add `data/sql/MIGRATIONS_FROZEN`, so CI refuses a
       change to a migration on main, and add numbered migrations from then on (docs/migrations.md). Then add a test
       that the last release's DBs migrate to the current schema, comparing each DB's `Schema` with a new one's as
       `cmd/try_migrate_or_delete` does
    1. P1: Decide the deploy tools from the gaps in `docs/deploy.md`, which lists each way a deploy can fail and what
       fixes it, starting with a site that stops the start (seeing why, trying a fix on a copy, and restoring,
       forcing back, disabling or deleting it) and a missing control DB starting a new server in setup
1. Push notifications (the control DB's `push_credentials` and `push_registrations`):
    1. P3: Add push credentials and registrations to `clients/nighttroop/openapi.json` and its client
    1. P2: Push alarms to each identity's registered apps when one goes off, through `push.Send`
    1. P2: Before any server credential (one with no site) is added: use it only for tokens its app registered itself,
       require its bundle IDs, and limit pushes across the server, as any site could otherwise push to any token. Then
       let the control site's Push Credentials (`POST /push-credentials`) add one with no `site_id`, which it refuses
       until then
    1. P3: Encrypt Loop's and Trio's device tokens at rest: they are in plaintext in `loop_settings`, though only
       identities that can change the site's data are given them back, and push registrations keep them encrypted
       `controlcollections.AdminOnly` in `server/control_router.go`, as its other admin only routes are, and drop
       their handlers' own admin checks
1. Nightscout API:
    1. P4: The Alexa and Google Home intents have Nightscout's bg, delta, iob, cob and uploader battery, but not its
       basal, raw bg, forecast, pump reservoir and battery, CGM status, db size, LastLoop and InsulinRemaining, nor
       the temp basal in NSStatus (`nsv2/assistants.go`)
    1. P4: The Nightscout API gives glucose only in mg/dL, so the `units_mmol_label` setting no longer changes what it
       writes. Drop it from the settings, the `sites` and `site_identities` columns and their queries
    1. P4: A search sorted by date with a filter other than the date picks its page by the dates alone, then filters
       it, so a page can be short, and empty when none of it matches although later ones do. No app sends such a
       search yet; filtering in the oldest first query would need each search's filters there too
    1. P4: Return a food's Nightscout `portion`, `unit` and `gi` as sent; AAPS takes 0, g and none without them
    1. P4: Keep arbitrary API v3 `settings` documents and support v2 authorization roles and subjects, which AAPS 4's
       client control (pairing and remote commands) uses
    1. P4: Only write a treatment's events that changed, rather than every one on each update
    1. P4: Keep when a site's settings last changed, for API v3 lastModified
    1. P5: Let the site clear a sensor change's sensor code or transmitter ID, as an empty one keeps what it had
1. Loop ([docs/clients/loop.md](docs/clients/loop.md)):
    1. P2: Loop remote commands: `POST /api/v2/notifications/loop`, sent to Loop as an Apple push to its profile's
       `loopSettings.deviceToken`, once push notifications can be sent
    1. P4: Keep a Loop device status's pump `pumpID`, `suspended` and `bolusing`, and its uploader `name` and
       `timestamp`, once the status as sent is cleared; they are given back only while it is kept

## Bugs

