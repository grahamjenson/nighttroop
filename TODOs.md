# TODOs

## Before Beta Testing

1. Entry/Treatments Compression/truncation. Not holding onto years of full data, keep only the necessary stuff.
    1. P4: A site's DB size on its pages, and for its write limits (3 and 4 times its size limit), counts the free
       pages a truncate leaves, under 10% of it: measure it without them, as the truncate of a site far over its
       limit does
1. e2e restore and archive testing
    1. P2: Refuse to archive a site while S3 is off, as the site admin's delete does
1. Performance testing (cost estimate for running full node)
    1. P3: A Nightscout token (`?token=`, as LoopFollow, xDrip4iOS and Nightguard send on each request) is checked
       with bcrypt again 5 minutes after it was first checked (`utils/security/password_verifier_cache.go`), about
       70 ms each time and most of what such a client costs: decide a longer time, as a JWT from it lasts an hour
    1. P3: Each commit syncs the DB's write-ahead log to disk (`synchronous` is `FULL`, as the DSN in
       `data/sql/gen/db.go` sets none), so a v1 upload of N records is N syncs: decide `NORMAL`, which can lose the
       last commits when the power fails, for the site DBs or only the logs, jobs and live DBs
1. Register with Resend and PagerDuty and .p8 push notifications to test all.
1. Accounts
    1. P4: Add the icons of SugarPixel, Dexcom and FreeStyle Libre to `server/assets/static/logos`, and put them on the right pages.
    1. P2: Recover a site through its verified email, the recovery request must give the email address exactly.
    1. P4: Sent Emails lists the last 50 emails only, with no next page
    1. P4: Sent Emails should be able to identify the site by the subject line referencing it in

1. Security
    1. P2: A verification email's wait of an hour is checked before the email is sent, and kept after it
       (`mail.Links`: `Wait`, then `Sent`), so requests at once each send an email, each to its own address, from a
       site's or a control admin's Account: take the hour before the send, as control's email to a site does
    1. P4: The FCM and APNs token caches (`utils/push`) keep an entry for each credential ID, and a deleted
       credential's stays, so a site that adds, tests and deletes push credentials grows them: keep them in an LRU
    1. P2: Add a `script-src` Content-Security-Policy with nonces, as only `frame-ancestors` is set now: the templ and
       htmx pages use inline scripts and handlers (`onclick`, `hx-on`) that would need auditing
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

## Before Prod Deployment

1. P4: Refresh the built in DB-IP country database
1. P1: Test rolling a new keyset, moving to a new encrypted keyset.
1. P3: Define maintenance tasks.
1. P3: Define data processes
1. P2: Define DR recovery processes
1. P1: Freeze the migrations: add `data/sql/MIGRATIONS_FROZEN`, so CI refuses a
       change to a migration on main, and add numbered migrations from then on (docs/migrations.md). Then add a test
       that the last release's DBs migrate to the current schema, comparing each DB's `Schema` with a new one's as
       `cmd/try_migrate_or_delete` does

## Build these things

1. Control pages:
    1. P3: Control Page to verify backup_key, and regenerate a backup_key

1. Site usability:
    1. P4: While an urgent alarm is snoozed, the banner shows the alarm it quiets (a high under its urgent high) as
       "Snoozed: High", though that one has no snooze: leave it out, or say what quiets it
    1. P5: The dose and carbs boxes of the add and details forms open a phone's full keyboard, as they take a
       negative value and a phone's number keys have no minus sign: decide a way to take back a value with number keys

1. Migrations (from the audit of what happens when one fails):
    1. P1: Decide the deploy tools from the gaps in `docs/deploy.md`, which lists each way a deploy can fail and what
       fixes it, starting with an unavailable site (seeing why, trying a fix on a copy, and restoring, forcing back,
       disabling or deleting it) and a missing control DB starting a new server in setup


1. Tests:
    1. P4: A browser test cannot check a text with a time of day in it that is not now, as the end of an alarm's
       snooze 30 minutes on: `{{ now }}` takes no minutes to add, so a test checks only the words before the time
    1. P4: The tests back sites up into a `.dbs` folder beside each test package, kept from run to run (`tests/api`:
       about 5 MB and 170 files a run): give the test config a data folder of the test's own
    1. Steps of the test scripts (`tests/dsl/flow`) that pass when they cannot check, none of which hides a failure in
       a test file now:
        1. P3: `GOTO` ignores the error of an assert after it, so `ASSERT_HELD` with nothing held, or a `SCREENSHOT`
           over its size limit, passes there: fail the step, as the other steps do
        1. P4: `ASSERT_EQUALS ".a.id" "@result.id"` passes when neither value is there, as both are then null: refuse
           an expected value that is looked up and not there
        1. P4: `ASSERT_NOT_VISIBLE` and `ASSERT_COUNT` do not fill in `{{ }}`, so one with a value in its locator finds
           nothing and can pass


1. Push notifications (the control DB's `push_credentials` and `push_registrations`):
    1. P3: Add push credentials and registrations to `clients/nighttroop/openapi.json` and its client
    1. P2: Push alarms to each identity's registered apps when one goes off, through `push.Send`
    1. P2: Before any server credential (one with no site) is added: use it only for tokens its app registered itself,
       require its bundle IDs, and limit pushes across the server, as any site could otherwise push to any token. Then
       let the control site's Push Credentials (`POST /push-credentials`) add one with no `site_id`, which it refuses
       until then


1. Nightscout API:
    1. P3: Add the v2 `basal` property watch faces read: the profile's basal now (`sitemodels.Profiles.BasalAt`) and
       the temp basal running, as `display` and `current` (`basal`, `tempbasal`, `totalbasal`, `treatment`)
    1. P4: The v2 `pump`, `openaps` and `loop` properties read only the newest loop device status of the last 20
       minutes. Nightscout reads 6 hours (2 for the pump), so it also has a pump with no loop, and an earlier
       enacted when the newest status has none
    1. P4: A suspended pump (`PUMP_SUSPEND`) of 0 minutes is a basal of 0 in the summaries until the next temp basal;
       also end it at a `Resume Pump` event, kept as an event group, for a Loop that does not send the suspend again
       with its duration
    1. P4: Food, insulin and temp targets added on the site are not Nightscout treatments until grouped, so AAPS and
       other clients do not see them. Decide whether adding one on its page also adds its group
    1. P4: Only write a treatment's events that changed, rather than every one on each update
    1. P4: Keep when a site's settings last changed, for API v3 lastModified
    1. P5: Let the site clear a sensor change's sensor code or transmitter ID, as an empty one keeps what it had
    1. P4: Refuse a treatment's other values past their limits (e.g. fat, protein, a temp basal's rate, a duration)
       with their range and what was sent, as insulin and carbs are; the database refuses them saying only the column
    1. P4: A temp basal sent as a percent with no absolute (AAPS on a percent pump) is given back with no `absolute`
       or `temp`, as Nightscout gives it, so Loop's NightscoutKit drops it. Check what LoopFollow and Loop Caregiver
       show for one
    1. P5: A zero temp sent as a percent (-100, a multiplier of 0) cannot be saved in its edit form, as the Basal (%)
       field takes at least 1%: let it be edited, or show its basal percent as read only

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
    1. P3: Restoring an archived site reads only its copy here if there is one, so a damaged copy fails each time
       though S3 has a good one: download it when the copy does not decompress. Also refuse a restore of a site that
       is unavailable, as a second restore can delete the DB kept for it

1. Database:
    1. P4: Make `go generate` fail for a bare value that an indexed column equals in a query with an `ORDER BY`, as
       SQLite prepares such a query again on each run (data/sql/README.md); it fails now only for a bare `LIMIT`
    1. P4: `data/sql/gen/limits.go` reads a table's constraint written with no space, `UNIQUE(site_id, type)`, as a
       column, so `controldb`'s limits have a `"UNIQUE(siteId,"` field
    1. P5: Check the limits of the columns only an update writes (who updated a record, a site's and an identity's
       settings, a connector's sync), which `TestSiteDataLimits` and `TestControlDataLimits` leave out

1. Site pages:
    1. P3: Narrow a site's Events, Errors and Logs pages to one identity: the API takes `identity_id`, the pages
       have no field for it, and Manage's **Its events** and **Its errors** still search by name
    1. P4: The details of a profile, a device status and a loop's settings have no Delete button, as they have no
       `addmodal` edit form: add one, so they too are deleted on a phone with no scroll of the table to the row's icon
