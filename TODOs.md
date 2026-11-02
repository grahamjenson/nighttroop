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
    1. P2: Check If-Unmodified-Since on API v3 replace and patch, in the same query as the update
    1. P3: Juggluco on v1 posts up to 5000 readings in one `/api/v1/entries` request and sends the same batch again
       until it is accepted, so a batch over the 100 record cap never uploads
    1. P4: Only write a treatment's events that changed, rather than every one on each update
    1. P4: Keep when a site's settings last changed, for API v3 lastModified
    1. P1: Replace a treatment uploaded again rather than ignore it, as Nightscout does: Loop posts a temp basal or
       bolus again when it ends early, with the same date and `syncIdentifier`, and the first upload's duration and
       amount are kept ([docs/clients/loop.md](docs/clients/loop.md))
    1. P2: Keep a Loop device status's `override`, `uploader` and pump `clock`, `pumpID`, `model`, `suspended` and
       `bolusing`, which followers show; only the loop object, pump battery and reservoir are kept
    1. P2: Loop remote commands: `POST /api/v2/notifications/loop`, sent to Loop as an Apple push to its profile's
       `loopSettings.deviceToken`
    1. P2: The Nightscout API's `sgv` and `mbg` are in the site's units, but NightscoutKit and most followers read them
       as mg/dL, so on an mmol/L site they read 6.3 as 6.3 mg/dL
    1. P3: NightscoutKit's `fetchTreatments` drops treatments without a `timestamp`, a bolus without `programmed`
       and a temp basal without `temp`, which Nighttroop does not return
    1. Trio and iAPS ([docs/clients/trio.md](docs/clients/trio.md)):
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
        1. P3: Keep the device status `pump.clock`, `pump.status`, `pump.bolusIncrement`, the pump battery's
           voltage and state, and `uploader` (`battery`, `isCharging`) Trio sends, for its followers

## Bugs

1. P4: A live site left out at start, its DB missing and not restored, is only a `site_db_missing` event; mark it on
   the control sites page too

