# Deploying Nighttroop: what happens, what can fail, and what to do

This is the deploy strategy: every step a deploy takes, in order, what can go wrong at each one, and how to fix it.
It is written before more deploy tooling is built, so the tools can be chosen from it. Where nothing fixes a failure
yet, it says so under **Gap**, and [the gaps](#the-gaps) at the end lists them together.

There are two kinds of deploy:

- [**A production update**](#a-production-update): a VPS running `docs/deploy/hostinger` or `docs/deploy/voyager`,
  updated with [`update.sh`](deploy/update.sh) to an image built from a version tag.
- [**A staging deploy**](#a-staging-deploy): the Fly.io server in [`docs/deploy/staging_flyio`](deploy/staging_flyio),
  deployed on each push to `main`. Its data can be wiped, so it matters less.

Both run the same server, so [what the server does when it starts](#5-the-new-version-starts) and
[stops](#3-the-old-version-stops) is the same for both.

## What data there is, and where

Knowing what each deploy step puts at risk depends on what is kept where:

| Data | On the server | Backed up to S3 | Lost if missing |
| --- | --- | --- | --- |
| Control DB: sites, identities, the admin, settings | `data/db/controldb.sqlite` | Every 30 seconds (`CONTROL_BACKUP_INTERVAL_IN_SECONDS`), and on stop | Every site, as the server starts as a new one in setup |
| Site DBs: each site's readings, treatments and settings | `data/db/site/<batch>/<site>.sqlite`, a batch per first letter | Each batch every 5 minutes (`BATCH_BACKUP_INTERVAL_IN_SECONDS`), and on stop | The start stops until it is restored (see [5.8](#58-restore-missing-sites)) |
| Jobs DB: queued and past background jobs | `data/db/jobsdb.sqlite` | No | Queued jobs and job history; they start again |
| Logs DB: events, errors, logs | `data/db/logsdb.sqlite` | No | The history the control plane shows |
| Live DB: what the server keeps only while running, like alarms | `data/db/livedb.sqlite` | No | Alarms going off now, made again as readings come in |
| Local backups of the control and site DBs | `data/backup/` | They are what is uploaded | Nothing, while the DBs are there |
| `update.sh`'s copies of `./data` | `backups/data-<time>.tar.gz`, the newest 5 | No | The way back to before a release |
| `KEY_SETS` | `nighttroop.env` | No: keep it in a password manager | Every encrypted credential, and every S3 backup |

S3 backups are encrypted with `KEY_SETS`, and S3 keeps each replaced backup for 90 days (control) or 30 days (sites),
when the bucket has versioning.

## A production update

### 0. A version is released

1. A tag such as `v1.2.3` is pushed on a commit on `main`.
2. The **Docker image** workflow (`.github/workflows/docker.yml`) builds the `Dockerfile`'s default target and pushes
   `ghcr.io/grahamjenson/nighttroop:v1.2.3` and `:latest`.
   - The build downloads DB-IP's city database. If it cannot, the image is still built, and only knows countries.

| What goes wrong | What happens | Fix |
| --- | --- | --- |
| The build fails | No image is pushed; servers keep running what they run | Fix it on `main`, then tag a new version. Don't move a tag that was pushed. |
| The tag is on a commit CI failed on, or not on `main` | The image is built anyway | Don't deploy it; tag a fixed commit. **Gap:** nothing stops it being built. |
| `:latest` moves to a version not yet tried | `update.sh` with no image argument deploys it | Always deploy a version tag, never `:latest`. |
| The city database could not be downloaded | Sign-ins show only a country | Nothing to fix; the next build tries again. |

### 1. `update.sh` checks it can run, and pulls the image

On the VPS, `sudo ./update.sh ghcr.io/grahamjenson/nighttroop:v1.2.3`:

1. It checks it runs as root, so `./data` is copied with its owner.
2. It checks Nighttroop is running.
3. It pulls the new image, while the old version keeps serving.
4. If that image is the one running, it stops with exit code 0.

| What goes wrong | What happens | Fix |
| --- | --- | --- |
| Not run as root | Stops, changing nothing | Run it with `sudo`. |
| Nighttroop is not running | Stops, changing nothing | Find out why it stopped (`docker compose logs nighttroop`) before updating; start it with `docker compose up -d`. |
| The pull fails: a wrong tag, the registry down, no network | Stops, changing nothing | Check the tag on the repository's Packages page, and try again. |

Nothing has changed yet, so every failure here is safe.

### 2. The running image is kept

It tags the running image `nighttroop-rollback:<time>`, so it can start it again whatever its tag points to by then.

### 3. The old version stops

`docker compose stop` sends the server `SIGTERM`, and kills it after the compose file's `stop_grace_period`, 2
minutes. The server then, in order:

1. stops taking HTTP requests, giving those already in flight 5 seconds;
2. stops the background jobs (syncs, alarms, backups), giving those running up to a minute to finish;
3. backs the control DB up locally and uploads it to S3;
4. backs up each batch of site DBs whose sites changed since its last upload, and uploads it;
5. closes every DB.

**Nighttroop is down from here until the new version is healthy in [step 6](#6-updatesh-waits-for-it-to-be-healthy)**,
usually under a minute. Caddy keeps running and answers with a `502`.

| What goes wrong | What happens | Fix |
| --- | --- | --- |
| S3 is unreachable | The last uploads fail and are logged; the DBs on disk are whole | Nothing for this server: its data is on disk. S3 is behind until the new version's first backups. Don't move to a new VPS until it has caught up. |
| A job or the uploads take longer than 2 minutes | The server is killed; SQLite's write-ahead log keeps the DBs whole | As above. If it is often slow, raise `stop_grace_period`. **Gap:** nothing records that the last backup didn't finish. |
| A phone or app uploads while it is down | Its upload fails | Loop, AAPS, Trio and xDrip+ keep what they could not upload and send it again. **Gap:** which apps retry, and how far back Dexcom, Libre and Nightscout syncs fetch what was missed, isn't written down or tested. |

### 4. `./data` is copied

With nothing writing to it, it copies `./data` to `./backups/data-<time>.tar.gz`. This copy is what a rollback puts
back.

| What goes wrong | What happens | Fix |
| --- | --- | --- |
| The disk is full, or the copy fails | **The script stops with Nighttroop stopped**, still on the old image, as `.env` is not changed yet | Free space (`du -sh /opt/nighttroop/*`, old `data-failed-*` folders, `docker system prune`), then `docker compose up -d nighttroop` starts the old version again. **Gap:** the script doesn't start it again itself, or check for space first. |

### 5. The new version starts

The script writes the new image into `.env` and runs `docker compose up -d`. The server then starts in this order.
Each step that fails ends the process with its error, and Docker starts it again (`restart: unless-stopped`), so a
start that always fails restarts until `update.sh` gives up in [step 6](#6-updatesh-waits-for-it-to-be-healthy) and
rolls back.

#### 5.1 Read the settings

The settings come from `nighttroop.env` and the compose file ([docs/config.md](config.md)).

- **Fails** when a setting has a bad value, or one it needs, such as `KEY_SETS`, is missing: `cannot create server
  config`. The public `KEY_SETS` of `.env.dev`, and `minioadmin` S3 keys, are bad values here.
- **Fix:** correct `nighttroop.env`, and update again. A new version that needs a new setting must say so in its
  release.

#### 5.2 Start S3

When `S3_URL` is set, it uploads a test object, trying again after 2, 4, 8 and 16 seconds.

- **Fails** when S3 is still not working after about 30 seconds: `S3_URL is set but S3 is not working`. A server that
  cannot reach its backups doesn't start, so it never starts empty and uploads over them.
- **Fix:** an S3 outage needs waiting out. Wrong keys or a removed bucket need `nighttroop.env` fixing. The rollback
  will fail the same way, so `update.sh` exits with 2; start it again with `docker compose up -d` once S3 works.
- Without `S3_URL`, S3 is off: nothing is restored or backed up, and the control plane says so.

#### 5.3 Remove leftovers

It removes the staging folders a killed restore or archive left beside the DBs. Nothing here stops the start.

#### 5.4 Restore the control DB, if it is missing

Only when `data/db/controldb.sqlite` is not there does it download the control DB's latest backup from S3.

- **Fails** when the download fails: the start stops.
- **The dangerous case:** when the file is missing and **there is no backup, or S3 is off**, the server starts as a
  **new server, in setup**. It knows none of its sites, though their files are still in `data/db`. Its health check
  passes, so `update.sh` keeps it.
  - **Fix:** stop it before anyone sets it up: `docker compose stop nighttroop`. Put `data/db/controldb.sqlite` back from
    the newest `backups/data-*.tar.gz` (`tar -xzf <copy> data/db/controldb.sqlite`) and start it again. Don't run setup on
    it.
  - **Gap:** nothing stops a server that has site DBs starting with an empty control DB.

#### 5.5 Restore missing batches

For each batch whose folder is missing, it downloads the batch's latest backup. A folder that is there is never
downloaded over.

- **Fails** when a download fails. The batch is left without a folder, so the next start tries it again.
- **Fix:** as for S3 above.

#### 5.6 Open the control, jobs, logs and live DBs

It opens each, running its new migrations.

- **Fails** when one cannot be opened or a migration fails: `cannot create control db`, naming the DB and migration.
  A migration that fails is rolled back with its transaction, so the DB is as it was, marked dirty. The next open
  forces it back a version and runs the migration again ([docs/migrations.md](migrations.md)).
- **Fix:** `update.sh` rolls back, and the DB is put back from the copy in step 4. Then fix the migration and
  release again.

#### 5.7 Open every site

It opens each live and disabled site's DB, on every core at once, running its new migrations. A site whose file is
missing is noted for the next step.

- **Fails** when any site cannot be opened or migrated (`cannot fill sites`). One bad site stops the whole start.
  This is on purpose, so no site is served half migrated and nothing is written that a rollback would lose.
- **Fix:** `update.sh` rolls back. Finding which site and why means reading the failed version's logs, which
  `update.sh` prints, and the site's file in `data-failed-<time>`.
- **Gap:** there is no way to see why one site fails, to try a fix on a copy, or to take one site out of the start
  (disable it, restore it from a backup, or force its migration back) without a shell and `sqlite3`.
- A large server takes a while here, about 2 ms a site.

#### 5.8 Restore missing sites

A live site whose file was missing is restored from its batch's latest S3 backup, and opened. Each is recorded as a
`site_db_missing` event.

- **Fails** when a site is still missing: S3 is off, the backup doesn't have it, or the download fails. It also fails
  when a restored site cannot be opened. `live sites' DBs are missing and were not restored` names the sites.
- **Fix:** put the site's file back from a copy in `./backups`, or decide the site is gone and delete it in the
  control DB. **Gap:** deleting it needs the server running, which it can't be.

#### 5.9 Make the admin

Development and staging make their admin. A server with no admin prints a setup code and serves only the setup page.

#### 5.10 Start the background jobs

Syncs, backups, alarms and the other jobs start, before the server takes requests. So a sync can write to a site DB
from here on.

#### 5.11 Take requests

It listens on port 3000. `/health` answers `ok` once the control DB does.

- **Fails** when the port can't be listened on: the process exits.
- **Gap:** the background jobs may already have written by then, so the rollback's copy would lose those writes. They
  are only syncs fetching what they can fetch again.

### 6. `update.sh` waits for it to be healthy

It waits up to `HEALTH_TIMEOUT` (default 180 seconds) for Docker's health check. That check asks `/health` every 10
seconds, after a start period of 120 seconds, and is unhealthy after 3 failures.

| What goes wrong | What happens | Fix |
| --- | --- | --- |
| The start fails ([step 5](#5-the-new-version-starts)) | It is never healthy, so the script rolls back ([step 7](#7-if-it-isnt-healthy-it-rolls-back)) | Read the logs it prints, and fix the cause. |
| A good version starts slowly: a big restore, many sites | It isn't healthy within 180 seconds (or within Docker's 150), so a version that would have worked is rolled back | `sudo HEALTH_TIMEOUT=600 ./update.sh ...`. **Gap:** Docker's own start period is fixed at 120 seconds in the image. |
| It is healthy but broken: a page errors, syncs fail, a query is wrong | The script keeps it, and it serves | Fix forward: release a fix and update again ([below](#once-a-new-version-has-served-fix-forward)). **Gap:** `/health` only checks the control DB; nothing checks a page, a site or a sync after a deploy. |
| It starts as a new server in setup ([5.4](#54-restore-the-control-db-if-it-is-missing)) | The script keeps it | See 5.4. |

When it is healthy, the script deletes all but the newest 5 copies of `./data` and previous images, and exits with 0.

### 7. If it isn't healthy, it rolls back

1. It prints the new version's last 50 log lines.
2. It stops it, and moves its `./data` to `./data-failed-<time>`.
3. It puts back `./data` from step 4's copy.
4. It starts the previous image, and waits for it to be healthy.
5. It exits with 1 when the previous version is back, or 2 when it isn't healthy either.

This loses nothing. A version that is never healthy took no requests, and its own DB changes are in
`data-failed-<time>`.

| What goes wrong | What happens | Fix |
| --- | --- | --- |
| The previous version isn't healthy either (exit 2) | Nighttroop is down | The cause isn't the release: S3, a setting, the disk, the VPS. Read `docker compose logs nighttroop`, fix it, and `docker compose up -d`. |
| The new version uploaded backups before it failed | S3 holds backups in the new version's format | The previous version replaces them on its next backups, within 5 minutes. Don't restore a new VPS from S3 until then. |

`data-failed-*` folders are kept until deleted by hand.

### Once a new version has served, fix forward

An earlier version can't open DBs a newer one migrated. So once a new version has served, going back means putting
back an old copy of `./data` and losing everything written since. Release a fix and update again instead.

Only as a last resort, roll back by hand from the copy made before that version, losing what was written since: see
the guide's **Rolling back by hand** ([Hostinger](deploy/hostinger/hostinger.md#rolling-back-by-hand)).

- **Gap:** nothing keeps what was written after that copy, to bring it into the older version's DBs.

## Things that can go wrong on any deploy

| What goes wrong | What happens | Fix |
| --- | --- | --- |
| `KEY_SETS` is changed, cut short or lost | Encrypted credentials (Dexcom, Nightscout) can't be read, and S3 backups can't be decrypted | Put the old value back from the password manager. **Gap:** there is no way to move to new keys yet (TODOs: *Testing deploy with rolling to a new keyset*). |
| A migration on `main` was edited rather than added, after the first production deploy | DBs already past it never run the change, so old DBs and new ones differ | CI refuses it once `data/sql/MIGRATIONS_FROZEN` is there. **Gap:** it isn't there yet, so nothing refuses it (TODOs P0). |
| A migration is slow on big tables | The start takes longer than `HEALTH_TIMEOUT`, and is rolled back | Give it longer with `HEALTH_TIMEOUT`, or make the migration cheaper. **Gap:** nothing measures a migration on a large DB before it ships. |
| The VPS dies mid-deploy | Whatever step it was in is left half done | Start the VPS, then `docker compose up -d`. If `./data` is damaged, put back the newest `backups/data-*.tar.gz`. |
| Two servers use the same bucket, e.g. a new VPS while the old one still runs | Each overwrites the other's backups | Stop the old server, and let it upload its last backups, before starting the new one. |
| A new VPS restores from S3 | It gets the control DB from up to 30 seconds before, and each batch from up to 5 minutes before, if the old server died without its last backup. The jobs, logs and live DBs start empty. | Expected. Sites made in those last seconds are in the control DB but missing from a batch, which stops the start ([5.8](#58-restore-missing-sites)). |

## A staging deploy

On each push to `main`, the **Deploy staging** workflow (`.github/workflows/deploy-staging.yml`) runs `flyctl deploy`
with the `Dockerfile`'s `staging` target, without waiting for CI. One deploy runs at a time, and the newest push
waiting is deployed next.

1. **Fly builds the image.** If the build fails, the deploy fails, and the old version keeps running.
2. **Fly stops the old machine**, as the volume can only be on one machine. The server stops as in
   [step 3](#3-the-old-version-stops), with no S3 to upload to.
3. **Fly starts the new machine** with `staging_start`, which:
   1. gives `/data` to the server's user;
   2. runs `try_migrate_or_delete`, which opens and migrates each DB, checks its schema is the one the migrations make
      on a new DB, and prepares its queries. **If any one fails, it deletes every DB**, and the server starts empty;
   3. starts the server, as in [step 5](#5-the-new-version-starts), with no S3.
4. **Fly waits for `/health`**, after 30 seconds' grace.

| What goes wrong | What happens | Fix |
| --- | --- | --- |
| A migration was changed | Every DB is deleted: testers' sites are gone | Expected on staging. The logs say `deleting every database, as one cannot be migrated`. |
| The server can't start: missing secrets, a wrong `HOST_NAME` | The deploy fails its health check, and staging is down | See the staging guide's [Troubleshooting](deploy/staging_flyio/README.md#troubleshooting). Fix it on `main`, or set the secret; the next deploy brings it back. **Gap:** whether Fly goes back to the old version when the health check fails isn't written down. |
| `FLY_API_TOKEN` expired | The workflow fails; staging keeps the old version | Make a new token (staging guide, step 5). |

## The gaps

These are the failures with no fix but a shell on the server, or none at all. Choose the tools from this list.

1. **A site that stops the start** ([5.7](#57-open-every-site), [5.8](#58-restore-missing-sites)):
   - seeing which site and why;
   - trying a fix on a copy;
   - taking the site out of the start: restoring it from a backup, forcing its migration back, or disabling or
     deleting it.
2. **A missing control DB starts a new server** ([5.4](#54-restore-the-control-db-if-it-is-missing)), and passes
   its health check, when site DBs are there.
3. **No check after a deploy beyond `/health`** ([step 6](#6-updatesh-waits-for-it-to-be-healthy)): nothing opens a
   page, a site or a sync.
4. **A failed copy leaves Nighttroop stopped** ([step 4](#4-data-is-copied)), and the script doesn't check for
   space first.
5. **What clients do while it is down** ([step 3](#3-the-old-version-stops)): which apps send again what they couldn't
   upload, and how far back each sync fetches.
6. **Fixing forward is the only way back** once a version has served
   ([above](#once-a-new-version-has-served-fix-forward)), and nothing measures a migration on large DBs first.
7. **The migrations aren't frozen**, as there is no production deploy yet.
8. **No way to move to a new `KEY_SETS`.**
9. **A slow start is rolled back**: Docker's 120-second start period is fixed in the image.
10. **Background jobs start before requests are taken** ([5.10](#510-start-the-background-jobs)), so a start that
    fails at the last step may have written.
11. **What Fly does when a staging deploy fails its health check.**
