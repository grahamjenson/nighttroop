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
    1. P3: Refresh the built in DB-IP country database (`utils/geoip/dbip-country-lite.mmdb`) each month or two, as
       DB-IP updates it monthly; the Docker image's city database is fetched fresh on each build.

## Before First Release:

1. Think about Translations
1. Docs hosting
1. Fuzz testing beyond the API (`tests/api/fuzz_test.go`)
1. P3: Add event groups (list, get, add, edit, delete and the events to group) to `clients/nighttroop/openapi.json`
   and its client
1. Performance testing (cost estimate for running full node)
1. P3: Open a site's DB when it is first used rather than every site's at start, as each still takes ~2ms to open
   and check migrations
1. P4: Keep which batches were uploaded, and with which sites, across restarts, as the first backup round after a
   restart uploads every batch again though their backups are current
1. Push notifications (the control DB's `push_credentials` and `push_registrations`):
    1. P2: Collections for push credentials, a site admin's for their site and a control admin's for every site,
       keeping the `.p8` key or service account encrypted and never returning it, with a cap on each site's
    1. P2: An identity registers an app's push token, kept encrypted with its SHA-256; Loop's comes from its profile's
       `loopSettings.deviceToken`, `bundleIdentifier` and `isAPNSProduction`
    1. P2: Send pushes with the credential `PushCredentialFor` finds: APNs with an ES256 token from the `.p8`, then FCM
       with the service account, rate limited per site, recording each push and deleting tokens the service says
       are gone
    1. P2: Delete a site's push credentials and registrations when the site is deleted
    1. P2: Site settings and control pages to add, list and delete push credentials, with BDD tests
1. Nightscout API:
    1. P1: Nightscout clients (AAPS, Loop, LoopFollow, xDrip4iOS, Nightguard, the xDrip+ follower) read `sgv` and
       `mbg`, and AAPS its temp targets and profiles, as mg/dL whatever the `units`, but the API gives them in the
       identity's units, then the site's. A device added for a client in Add Device is in mg/dL (#524) and an admin
       can set any identity's units (#523), but the site's first PWD phone, a device added as Other, and LoopFollow,
       xDrip4iOS and an AAPSClient follower, which are not in Add Device's clients, still get the site's units. Add
       them to Add Device's clients
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
    1. P2: Check If-Unmodified-Since on API v3 replace and patch, in the same query as the update
    1. P4: Only write a treatment's events that changed, rather than every one on each update
    1. P4: Keep when a site's settings last changed, for API v3 lastModified
1. LoopFollow ([docs/clients/loopfollow.md](docs/clients/loopfollow.md)):
    1. P2: Keep Loop's `automatic` on boluses, so LoopFollow shows automatic boluses as SMBs and leaves them out of
       its bolus stats
1. Loop ([docs/clients/loop.md](docs/clients/loop.md)):
    1. P1: Replace a treatment uploaded again rather than ignore it, as Nightscout does: Loop posts a temp basal or
       bolus again when it ends early, with the same date and `syncIdentifier`, and the first upload's duration and
       amount are kept. Trio's ended temp targets and overrides, and xDrip4iOS's re-sent readings, need the same
    1. P2: Loop remote commands: `POST /api/v2/notifications/loop`, sent to Loop as an Apple push to its profile's
       `loopSettings.deviceToken`, once push notifications can be sent
    1. P4: Keep a Loop device status's pump `pumpID`, `suspended` and `bolusing`, and its uploader `name` and
       `timestamp`, once the status as sent is cleared; they are given back only while it is kept
1. xDrip4iOS ([docs/clients/xdrip4ios.md](docs/clients/xdrip4ios.md)):
    1. P2: Return a BG Check's `glucose` as xDrip4iOS can match it: it reads it as mg/dL only when `units` is
       exactly `mg/dl`, and as mmol/L otherwise, and matches an upload by its exact value, so Nighttroop's
       `mg/dL` or rounded mmol/L makes it upload the BG check again each sync and add a copy in the app
    1. P2: Delete entries by a query, `DELETE /api/v1/entries.json?find[type]=sgv&find[date][$gte]=..&find[date][$lte]=..`
       and `find[date][$in][]=..`, which xDrip4iOS sends when a reading is deleted or readings are rebuilt
1. xDrip+ ([docs/clients/xdrip.md](docs/clients/xdrip.md)):
    1. P2: An unknown `find[...]` filter is ignored, so the search gives everything, and a client that deletes the
       first match (as xDrip+ does by `find[uuid]`) deletes the wrong document. Give nothing for a filter on an unknown
       field
    1. P4: Give back a treatment's `uuid` as sent; xDrip+ matches its own by `enteredBy` starting `xdrip`, kept, and
       falls back to matching time and amounts
1. Trio and iAPS ([docs/clients/trio.md](docs/clients/trio.md)):
    1. P2: v1 delete by query (`DELETE /api/v1/treatments.json?find[id][$eq]=`, `find[created_at][$eq]` with
       `find[eventType][$eq]`, and entries' `find[$or]`) is a 405; Trio deletes carbs, insulin and glucose
       this way, and the override before uploading its run, so override runs are never uploaded. It needs
       Trio's `id` kept as the identifier, and must refuse any filter it does not know rather than ignore it

## Bugs

1. P4: A live site left out at start, its DB missing and not restored, is only a `site_db_missing` event; mark it on
   the control sites page too
