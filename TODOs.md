# TODOs

## Before Alpha Testing

1. Deploy a staging server ([docs/deploy/hostinger.md](docs/deploy/hostinger.md)) with S3 backups and `SENTRY_DSN` set,
   and check a fresh server restores from those backups
1. Test end to end with real devices: Loop, Trio, AAPS and xDrip+ uploading, follower apps (xDrip+ follower, LoopFollow,
   Nightguard) reading, and Dexcom and Libre sync
1. Medical disclaimer: Nighttroop is not a medical device and its alarms are not to be relied on, shown when a site is
   created and on the alarm settings
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
    1. P4: Only write a treatment's events that changed, rather than every one on each update
    1. P4: Keep when a site's settings last changed, for API v3 lastModified
1. LoopFollow ([docs/clients/loopfollow.md](docs/clients/loopfollow.md)):
    1. P1: LoopFollow reads `sgv` as mg/dL, as Nightscout always sends it, so an identity in mmol/L shows 6.7 as
       7 mg/dL with urgent low alarms; send it mg/dL (its requests say `LoopFollow` in the User-Agent) or warn
    1. P2: Keep Loop's `automatic` on boluses, so LoopFollow shows automatic boluses as SMBs and leaves them out of
       its bolus stats
    1. P2: Keep Loop's `durationType: indefinite` on a Temporary Override; without it the override has duration 0,
       which LoopFollow does not show
    1. P3: LoopFollow asks for up to 5000 treatments and 1152 entries a day, above the 1000 a search returns, so a
       Trio day of SMBs and temp basals, or weeks of stats, are cut to the newest 1000
    1. P3: Add `authorized` (the JWT and permissionGroups) to status.json, with `api:treatments:create` for a
       write identity, so LoopFollow shows "Read & Write" rather than "Read"
    1. P4: Keep a treatment's `enteredBy`, which LoopFollow shows for overrides and temp targets, rather than the
       identity's name
    1. P4: Keep the pump's `pumpID`, so LoopFollow carries an Omnipod reservoir reading over statuses without one,
       and its `bolusIncrement`, LoopFollow's remote bolus step
    1. P4: A profile search with no filters gives only the latest store, so LoopFollow's remote diagnostics see no
       history
    1. P4: The Nightscout host's `/` is a 404, which LoopFollow's Nightscout tab shows; redirect it to the site

## Bugs

1. P4: A live site left out at start, its DB missing and not restored, is only a `site_db_missing` event; mark it on
   the control sites page too

