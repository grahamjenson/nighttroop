# TODOs

## Before Alpha Testing

1. Deploy a staging server ([docs/deploy/hostinger.md](docs/deploy/hostinger.md)) with S3 backups and `SENTRY_DSN` set,
   and check a fresh server restores from those backups
1. Test end to end with real devices: Loop, Trio, AAPS and xDrip+ uploading, follower apps (xDrip+ follower, LoopFollow,
   Nightguard) reading, and Dexcom and Libre sync
1. A tester guide: creating a site, keeping its credentials safe (they cannot be recovered), and connecting uploaders,
   followers and sync

## Before Beta Testing

1. Reports/Derived Values:
    1. AGP
    1. Time below and above range, as well as time in range
    1. Small multiples Daily

1. Alarms:
    1. Push notifications to phones, as alarms now only reach an open site page or a Nightscout socket client

1. Bad Sites
    1. write entry compression
    1. write device status compression
    1. write treatments compressions

1. Control UI
    1. Show bad users
    1. Block IP addresses
    1. Restore one site from its backup
    1. P3: Manage the control admins: list them, reset another's password or TOTP, change their role or delete one

1. Accounts
    1. Recovering a site's admin without the initial backup key
    1. Recovering a lost control admin login, as setup only runs on a server with no control admin

1. Security
    1. Testing deploy with rolling to a new keyset, moving to a new encrypted keyset.
    1. P4: A change to the site's settings is not streamed to a non-admin's `/settings/identity/events`, though its
       inherited values change; send it `{}`, as the summary is sent other models' changes, rather than the site's
    1. P3: Refresh the built in DB-IP country database (`utils/geoip/dbip-country-lite.mmdb`) each month or two, as
       DB-IP updates it monthly; the Docker image's city database is fetched fresh on each build.

## Before First Release:

1. Think about Translations
1. Docs hosting
1. Fuzz testing beyond the API (`tests/api/fuzz_test.go`)
1. P3: Add event groups (list, get, add, edit, delete and the events to group) to `clients/nighttroop/openapi.json`
   and its client
1. P4: Tag each endpoint in `clients/nighttroop/openapi.json` (site, data, settings, Nightscout), so the API docs'
   menu groups its 51 endpoints, and give each a summary, as those without one are listed by their operation ID
1. Performance testing (cost estimate for running full node)
1. Accessibility:
    1. P3: Keep Tab inside an open modal (`modal.Modal`), as past its Close button it goes on to the page behind
    1. P4: Show an identity's new credentials (`credentialsModal`) in a `modal.Modal`, so it is a dialog that takes
       the focus too
1. P3: Make a Nightscout document's fields without JSON: `stored.Fields` (`nscollections/document.go`) marshals each
   record's model and reads it back as a map (`utils.ToMap`) before the response is marshalled again, about 40% of
   the CPU of a 288 entry `/api/v1/entries` search
1. P3: Open a site's DB when it is first used rather than every site's at start, as each still takes ~2ms to open
   and check migrations
1. P4: Keep which batches were uploaded, and with which sites, across restarts, as the first backup round after a
   restart uploads every batch again though their backups are current
1. Migrations (from the audit of what happens when one fails):
    1. P1: A live site whose DB file is missing, and not restored from its batch's backup, still lets the server come up
       without it and `/health` say ok, so `update.sh` keeps the release rather than rolling it back. Its batch is no
       longer uploaded without it, so S3's latest keeps the site
    1. P0: At the first production deploy, freeze the migrations: add `data/sql/MIGRATIONS_FROZEN`, so CI refuses a
       change to a migration on main, and add numbered migrations from then on (docs/migrations.md). Then add a test
       that the last release's DBs migrate to the current schema
    1. P1: A migration that fails leaves its DB dirty, which every open refuses. Each runs in a transaction, so a dirty
       version N had none of N applied: force it back to N-1 and try again, or report it plainly
    1. P1: A build cannot open a DB a newer one migrated, so rolling back needs `update.sh`'s copy of `./data`, losing
       what was written since. Keep migrations additive, and let a build open a DB one additive version newer
    1. P2: Document changing a table: SQLite rebuilds it, which needs `PRAGMA defer_foreign_keys = ON` in the
       migration's transaction, as `event_groups` references the event tables `ON DELETE RESTRICT`; test migrations
       on a DB with data, and have staging compare the whole schema rather than only preparing queries
1. Push notifications (the control DB's `push_credentials` and `push_registrations`):
    1. P2: A control collection for push credentials for every site, as the site's (`/settings/push-credentials`)
    1. P3: Add push credentials and registrations to `clients/nighttroop/openapi.json` and its client
    1. P3: A delete button on Push Devices (`/settings/push-devices`), for an identity's own registrations and an
       admin's for the site's; the push registrations collection's HTMX responses are empty until then
    1. P2: Send FCM pushes with the service account (an OAuth token from its key, then the v1 send API), as
       `push.Send` does APNs ones
    1. P2: Push alarms to each identity's registered apps when one goes off, through `push.Send`
    1. P2: A control page to add, list and delete push credentials for every site, with BDD tests
    1. P2: Before any server credential (one with no site) is added: use it only for tokens its app registered itself,
       require its bundle IDs, and limit pushes across the server, as any site could otherwise push to any token
    1. P2: Decide whether Loop's and Trio's device tokens stay in plaintext in their profiles, which every reader of
       the site gets from the Nightscout profile API, as LoopFollow needs them
    1. P3: Register a profile's app in one immediate transaction, as two uploads at once can pass the 10 limit
    1. P3: Cap a site's deleted push credentials too, or delete them for good, as only live ones are counted
1. Nightscout API:
    1. P4: The Alexa and Google Home intents have Nightscout's bg, delta, iob, cob and uploader battery, but not its
       basal, raw bg, forecast, pump reservoir and battery, CGM status, db size, LastLoop and InsulinRemaining, nor
       the temp basal in NSStatus (`nsv2/assistants.go`)
    1. P1: Nightscout clients (AAPS, Loop, LoopFollow, Nightguard, the xDrip+ follower) read `sgv` and `mbg`, and
       AAPS its temp targets and profiles, as mg/dL whatever the `units`, but the API gives them in the identity's
       units, then the site's. A device added for a client in Add Device is in mg/dL (#524) and an admin can set any
       identity's units (#523), but the site's first PWD phone, a device added as Other, and LoopFollow and an
       AAPSClient follower, which are not in Add Device's clients, still get the site's units. Add them to Add
       Device's clients
    1. P1: A Nightscout batch of over 100 records is refused with a 400 (`consts.MaxBatchRecords`), and Loop (up to
       1000 a request) and Juggluco on v1 (up to 5000) send the same batch again until it is accepted, so after a
       backlog (first set up, or about 8 hours offline) they never upload again. Take larger v1 batches in chunks, or
       raise the cap for them (docs/clients/loop.md, docs/clients/juggluco.md)
    1. P1: Honour `sort` in searches and sort v3 history by `srvModified` oldest first, as Nightscout does. Searches
       are always newest first, so AAPS's first load (`sort=date&date$gt=…&limit=500`, back 100 days) and its
       history paging (`limit=500`) skip all but the newest 500 when there are more, and its profile history takes
       the oldest store changed rather than the newest. Needs history queries ordered by `updated_at`, not a `CASE`
       in each search's `ORDER BY`, which would stop SQLite reading the newest rows from the index
    1. P2: Keep the fields AAPS's followers need: an effective profile switch's `original*` fields (else it is read
       as a Note), a running mode's `mode`, `autoForced`, `reasons` and `originalDuration`, a temp basal's `type`
       (e.g. `PUMP_SUSPEND`), and a calibration's `isCalibration` and `sensorMgdlAtPairing`
    1. P4: Return a food's Nightscout `portion`, `unit` and `gi` as sent; AAPS takes 0, g and none without them
    1. P4: Keep arbitrary API v3 `settings` documents and support v2 authorization roles and subjects, which AAPS 4's
       client control (pairing and remote commands) uses
    1. P4: Only write a treatment's events that changed, rather than every one on each update
    1. P4: Keep when a site's settings last changed, for API v3 lastModified
1. LoopFollow ([docs/clients/loopfollow.md](docs/clients/loopfollow.md)):
    1. P2: Keep Loop's `automatic` on boluses, so LoopFollow shows automatic boluses as SMBs and leaves them out of
       its bolus stats
1. Loop ([docs/clients/loop.md](docs/clients/loop.md)):
    1. P1: Replace a treatment uploaded again rather than ignore it, as Nightscout does: Loop posts a temp basal or
       bolus again when it ends early, with the same date and `syncIdentifier`, and the first upload's duration and
       amount are kept. Trio's ended temp targets and overrides need the same, as entries now have
       (`CollectionRefresher`)
    1. P2: Loop remote commands: `POST /api/v2/notifications/loop`, sent to Loop as an Apple push to its profile's
       `loopSettings.deviceToken`, once push notifications can be sent
    1. P4: Keep a Loop device status's pump `pumpID`, `suspended` and `bolusing`, and its uploader `name` and
       `timestamp`, once the status as sent is cleared; they are given back only while it is kept
1. xDrip4iOS ([docs/clients/xdrip4ios.md](docs/clients/xdrip4ios.md)):
1. xDrip+ ([docs/clients/xdrip.md](docs/clients/xdrip.md)):
    1. P2: An unknown `find[...]` filter is ignored, so the search gives everything, and a client that deletes the
       first match (as xDrip+ does by `find[uuid]`) deletes the wrong document. Give nothing for a filter on an unknown
       field

## Bugs

1. P4: A live site left out at start, its DB missing and not restored, is only a `site_db_missing` event; mark it on
   the control sites page too
1. P2: Under overload a site whose lookup times out is answered 404, as if it did not exist, which an uploader may take
   as a wrong URL; answer 503 instead
1. P2: A site's Nightscout API is limited to `SITE_UNAUTH_PER_MINUTE_LIMIT` (60 a minute) from one IP even with a
   valid token, as the unauthenticated limit runs before auth on every Nightscout request, not only on failed ones
1. P2: AAPS sends `SettingsExport` treatments, which are refused as an unknown event type (Sentry NIGHTTROOP-6)
1. P3: When the server is busy, uploads from devices are refused as often as page loads; give Nightscout writes their
   own share of the requests in flight, so glucose keeps arriving
1. P3: A cancelled or timed out request is logged as an error at each layer it passes, sending several Sentry events
   for one request; log context cancellation below error
