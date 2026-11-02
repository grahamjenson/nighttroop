# TODOs

## Before Alpha Testing

1. Deploy a staging server ([docs/deploy/hostinger.md](docs/deploy/hostinger.md)) with S3 backups and `SENTRY_DSN` set,
   and check a fresh server restores from those backups
1. Test end to end with real devices: Loop, Trio, AAPS and xDrip+ uploading, follower apps (xDrip+ follower, LoopFollow,
   Nightguard) reading, and Dexcom and Libre sync
1. A way for testers to report bugs and give feedback, linked from the site
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
1. Nightscout API:
    1. P1: AAPS reads an entry's `sgv` as mg/dL whatever its units, and only knows the units `mg/dl` and `mmol`, so
       with an identity in mmol/L (the default in mmol/L countries) its readings, temp targets and profiles are read
       wrong. Return mg/dL to Nightscout clients, or set device identities to mg/dL (docs/clients/aaps.md)
    1. P1: Honour `sort` in searches and sort v3 history by `srvModified` oldest first, as Nightscout does. Searches
       are always newest first, so AAPS's first load (`sort=date&date$gt=…&limit=500`, back 100 days) and its
       history paging (`limit=500`) skip all but the newest 500 when there are more, and its profile history takes
       the oldest store changed rather than the newest. Needs history queries ordered by `updated_at`, not a `CASE`
       in each search's `ORDER BY`, which would stop SQLite reading the newest rows from the index
    1. P2: A temp basal's `percent` is taken as the percent of the basal, but AAPS sends the change (`rate - 100`,
       so -50 is half, -100 a zero temp) with the rate in U/hr; its negative percents are refused with a 400, which
       AAPS drops for good (docs/clients/aaps.md). Decide whether AAPS's percent is read as a change
    1. P2: Keep the fields AAPS's followers need: an effective profile switch's `original*` fields (else it is read
       as a Note), a running mode's `mode`, `autoForced`, `reasons` and `originalDuration`, a temp basal's `type`
       (e.g. `PUMP_SUSPEND`), a device status's `pump.clock`, `pump.status` and `pump.extended` (without `pump.clock`
       AAPS skips the pump's data), and a calibration's `isCalibration` and `sensorMgdlAtPairing`
    1. P2: Return a food's Nightscout `type` (`food`), `portion`, `unit` and `gi`; AAPS keeps only foods of type
       `food`, so it loads none
    1. P3: Accept the SHA1 of a token as the API secret on Socket.IO `authorize`, as AAPS 3.4 and older send it,
       which needs a kept hash of each token
    1. P4: Keep arbitrary API v3 `settings` documents and support v2 authorization roles and subjects, which AAPS 4's
       client control (pairing and remote commands) uses
    1. P1: Accept an `api-secret` that is the SHA1 hex of a device's Nightscout token, e.g. by keeping each token's
       SHA1 when it is made or reset, as Juggluco sends on v1, its default and its only mode on Wear OS
       ([docs/clients/juggluco.md](docs/clients/juggluco.md))
    1. P2: Check If-Unmodified-Since on API v3 replace and patch, in the same query as the update
    1. P3: Juggluco on v1 posts up to 5000 readings in one `/api/v1/entries` request and sends the same batch again
       until it is accepted, so a batch over the 100 record cap never uploads
    1. P4: Only write a treatment's events that changed, rather than every one on each update
    1. P4: Keep when a site's settings last changed, for API v3 lastModified
    1. P1: API v1 entries give `sgv` in the identity's or site's units, but Nightscout always gives mg/dL and the
       xDrip+ follower reads it as a whole number of mg/dL, so on an mmol/L site it shows 6.5 mmol/L as 6 mg/dL. Give
       mg/dL to v1 clients, or have device identities default to mg/dL ([docs/clients/xdrip.md](docs/clients/xdrip.md))
    1. P1: Keep xDrip+'s `Sensor Stop` treatments. An unknown eventType is refused, and xDrip+ uploads its treatments
       one at a time, stops at a refused one and marks them all done, so the treatments queued after a Sensor Stop are
       never uploaded. Adding the type also adds it to the site's event group types, a UI change
    1. P2: An unknown `find[...]` filter is ignored, so the search gives everything, and a client that deletes the
       first match (as xDrip+ does by `find[uuid]`) deletes the wrong document. Give nothing for a filter on an unknown
       field
    1. P3: An `mbg` entry also has `sgv` in v1, so the xDrip+ follower, which reads every entry as a reading, shows a
       finger stick as a CGM reading. Leave `sgv` off `mbg` entries
    1. P4: Keep xDrip+'s heart rate (`type` `hr-bpm`, `bpm`) and motion (`motion-class`, `class`) activity, which are
       accepted with only their time
    1. P4: Give back a treatment's `enteredBy` and `uuid` as sent; xDrip+ looks for `enteredBy` starting `xdrip` to
       skip its own, and falls back to matching time and amounts
    1. Trio and iAPS ([docs/clients/trio.md](docs/clients/trio.md)):
        1. P0: Accept `api-secret` as the SHA1 hex of a device's token, as Trio and iAPS send only that (it is all
           Nightscout checks), e.g. by keeping an HMAC of each token's SHA1 to look the identity up; until then
           Trio cannot sign in
        1. P1: Filter v1 treatments on `find[carbs][$exists]` and `find[duration][$exists]`, and on `enteredBy`
           (`find[enteredBy]`, `find[enteredBy][$ne]` and `find[$and][n][enteredBy][$ne]`), together: Trio fails
           to decode a carbs fetch with any treatment without carbs, and once it gets only carbs it would import
           its own back as new ones and double them, as `enteredBy` comes back as the identity, not `Trio`; iAPS
           takes any announcement as a remote command, not only `enteredBy` `remote`
        1. P1: A treatment uploaded again with the same `created_at` and `eventType` but a new duration (Trio's
           temp target and override runs, when one ends early) replaces the first, as Nightscout's v1 upsert does;
           now the first is kept, so an ended override stays 30 days long
        1. P2: v1 delete by query (`DELETE /api/v1/treatments.json?find[id][$eq]=`, `find[created_at][$eq]` with
           `find[eventType][$eq]`, and entries' `find[$or]`) is a 405; Trio deletes carbs, insulin and glucose
           this way, and the override before uploading its run, so override runs are never uploaded. It needs
           Trio's `id` kept as the identifier, and must refuse any filter it does not know rather than ignore it
        1. P3: v1 `find[created_at]` filters on when a record was uploaded, not Nightscout's `created_at`, when
           it happened
        1. P3: Keep the device status `pump.clock`, `pump.status`, `pump.bolusIncrement`, the pump battery's
           voltage and state, and `uploader` (`battery`, `isCharging`) Trio sends, for its followers

## Bugs

1. P4: A live site left out at start, its DB missing and not restored, is only a `site_db_missing` event; mark it on
   the control sites page too

