# TODOs

## Before Beta Testing

1. Entry/Treatments Compression/truncation. Not holding onto years of full data, keep only the necessary stuff.
    1. P2: Regenerate summaries rebuilds from site data, so it removes the summaries of days whose records the
       truncate of a site far over its size limit deleted: keep a day's summary when its records are gone
    1. P3: A site far over its size limit that keeps writing changes records on each truncate, so its DB is vacuumed
       every 10 minutes: vacuum only when its free pages are a large part of it
1. Push Notification testing. Alarms testing.
1. e2e restore/archiving/hydration testing
    1. P4: A BDD test of a site page's backup tiles once it is backed up: `RUN_BACKGROUND_JOBS` backs up each site
       written to, so its tiles can be checked after a backup, not only before one
1. e2e bakup_key testing. Control
1. Performance testing (cost estimate for running full node)

1. Accounts
    1. P4: Add the icons of SugarPixel and T1D Carer to `server/assets/static/logos`, with their makers' leave, as
       neither is open source: their tiles in the Add Device modal show their first letter
    1. P3: Start the Add Device modal from a device status or an identity used by more than one app, with that app
       picked, so a shared identity is one click from its own
    1. Recovering a site's admin without the initial backup key
    1. Recovering a lost control admin login, as setup only runs on a server with no control admin

1. Security
    1. P3: Capture each Socket.IO message on staging, as `CAPTURE_REQUESTS` keeps a websocket's upgrade but not its
       messages
    1. P2: Add a `script-src` Content-Security-Policy with nonces, as only `frame-ancestors` is set now: the templ and
       htmx pages use inline scripts and handlers (`onclick`, `hx-on`) that it would need auditing for first
    1. Abuse levels: the score and level (`controldb.AbuseLevel`) are shown to control admins only. Check the levels
       are right on real sites first, then act on them, each item its own PR:
        1. P2: Score a site's rate-limit refusals too: a count for each site that halves every hour, of refusals by the
           normal site and identity limits only (not the abuse throttle, nor a Blocked refusal, so they cannot feed
           themselves). Signal c ÷ (c + 50), at most 0.79 alone, and at most 0.79 with the connectors together
        1. P2: Over use: each write also spends 1 from `site_post_flagged` (each record of a batch); truncate the site
           harder (step 1, then step 2 if it is still over use). Replace `FlaggedForAbuse`'s 3 times and
           `hardTruncateSizeFactor`'s 1.5 times with the levels
        1. P2: Abusive: each write costs 4 from `site_post_flagged`; truncate step 2 at once
        1. P2: Blocked: refuse every write with 403 and say why, except glucose readings (the site's glucose create, every
           Nightscout `entries` write on v1, v3 and live updates), sign in and out, alarm snooze, Alexa and Google Home,
           and Truncate; its connectors sync glucose only. Reads are never refused
        1. P2: Delete a site's old records (truncate step 2) only while it is also over its size limit, as deleting does
           not help a site whose level comes from refusals or connectors
        1. P3: Record an event when a site's level changes, and sort the control sites list by abuse score
    1. P4: Keep passkey ceremonies' sessions per site or IP, as the server keeps at most 10,000 for all sites, and many
       begun at once push out others' before they finish

## Before Prod Deployment

1. Refresh the built in DB-IP country database
1. Test rolling a new keyset, moving to a new encrypted keyset.
1. Define maintenance tasks.
1. Define data processes
1. Define DR recovery processes
1. Freeze the migrations: add `data/sql/MIGRATIONS_FROZEN`, so CI refuses a
       change to a migration on main, and add numbered migrations from then on (docs/migrations.md). Then add a test
       that the last release's DBs migrate to the current schema, comparing each DB's `Schema` with a new one's as
       `cmd/try_migrate_or_delete` does

## Build these things

1. Control Page to verify backup_key, and regenerate a backup_key

1. Migrations (from the audit of what happens when one fails):
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


1. Nightscout API:
    1. P4: The Alexa and Google Home intents have Nightscout's bg, delta, iob, cob and uploader battery, but not its
       basal, raw bg, forecast, pump reservoir and battery, CGM status, db size, LastLoop and InsulinRemaining, nor
       the temp basal in NSStatus (`nsv2/assistants.go`)
    1. P3: Add the v2 `basal` property watch faces read: the profile's basal now (`sitemodels.BasalAt`) and the temp
       basal running, as `display` and `current` (`basal`, `tempbasal`, `totalbasal`, `treatment`)
    1. P4: The v2 `pump`, `openaps` and `loop` properties read only the newest loop device status of the last 20
       minutes. Nightscout reads 6 hours (2 for the pump), so it also has a pump with no loop, and an earlier
       enacted when the newest status has none
    1. P5: Temp targets, temp basals, profile switches and activities keep their duration in whole minutes, so the
       site's summaries end a temp target AAPS ended at 5.99 minutes at 5; only its event group keeps the
       milliseconds, for the Nightscout API's `durationInMilliseconds`
    1. P4: The Nightscout API gives glucose only in mg/dL, so the `units_mmol_label` setting no longer changes what it
       writes. Drop it from the settings, the `sites` and `site_identities` columns and their queries
    1. P4: A search sorted by date with a filter other than the date picks its page by the dates alone, then filters
       it, so a page can be short, and empty when none of it matches although later ones do. No app sends such a
       search yet; filtering in the oldest first query would need each search's filters there too
    1. P4: A v2 subject's `accessToken` is kept in memory for 10 minutes, so it is lost on a restart or on another
       server; AAPS pairing a client again after that gets no token. Keep it in the DB if this is a problem
    1. P3: Count an AAPS or Juggluco bolus with `isBasalInsulin` (long acting insulin logged by an MDI user) as basal
       in the summaries, reports and IOB, as it is kept and given back but still counted as a bolus
    1. P4: Food, insulin and temp targets added on the site are not Nightscout treatments until grouped, so AAPS and
       other clients do not see them. Decide whether adding one on its page also adds its group
    1. P4: Only write a treatment's events that changed, rather than every one on each update
    1. P4: Keep when a site's settings last changed, for API v3 lastModified
    1. P5: Let the site clear a sensor change's sensor code or transmitter ID, as an empty one keeps what it had
    1. P4: Refuse a treatment's other values past their limits (e.g. fat, protein, a temp basal's rate, a duration)
       with their range and what was sent, as insulin and carbs are; the database refuses them saying only the column

1. Backups:
    1. P2: Keep a dated daily copy of each written site (`<host>/sites/<site>/daily/<date>`), so the hourly versions
       of its backup can be kept a day or two
    1. P3: Delete a deleted site's backup (`<host>/sites/<site>.tar.enc`) from S3; only never used sites are deleted now
    1. P3: Delete the old batch backups (`<host>/backup/<batch>.tar.enc`) from the bucket by hand, once every site has a
       backup of its own; the server no longer reads them
    1. P4: Remove the restore's fallback to the old `.tar.lz4.enc` backups (`s3utils.LegacyKey`). Lifecycle rules
       expire only replaced versions (30 days for sites, 90 for control), and nothing replaces an object at an old
       key, so it stays. First, once every batch and the control DB has a `.tar.enc` backup, delete their old keys.
       Then upload each old archive again as `.tar.enc`, or keep the fallback for archives.
    1. P5: A backup killed while it writes leaves its temp file (`.<site>.sqlite.lz4-*`) in `data/backup`; remove
       these on start, as the restore staging folders are

1. Database:
    1. P3: CAST the bare values that a column equals in the other DBs' queries, as the site DB's are (data/sql/README.md),
       as SQLite prepares them again on each run: `controldb`'s `AllSiteIdentities` and its push
       queries, `jobsdb`'s `MarkJobsForConsumer`, `livedb`'s `IdentityAlarms`, and `logsdb`'s site and connector
       searches
    1. P4: `data/sql/gen/limits.go` reads a table's constraint written with no space, `UNIQUE(site_id, type)`, as a
       column, so `controldb`'s limits have a `"UNIQUE(siteId,"` field
    1. P5: Check the limits of the columns only an update writes (who updated a record, a site's and an identity's
       settings, a connector's sync), which `TestSiteDataLimits` and `TestControlDataLimits` leave out
