# Deploy Nighttroop: steps, failures and fixes

The deploy strategy: each step, its failures and fixes. Choose new deploy tools from it. **Gap** marks a failure with
no fix yet ([Gaps](#gaps)).

## Deploy types

| Type | Server | Deploy |
| --- | --- | --- |
| [Production update](#production-update) | A VPS that runs `docs/deploy/hostinger` or `docs/deploy/voyager` | [`update.sh`](deploy/update.sh), to an image built from a version tag |
| [Staging deploy](#staging-deploy) | The Fly.io server in [`docs/deploy/staging_flyio`](deploy/staging_flyio) | Each push to `main`. Its data can be wiped, so it is less important. |

Both run the same server, so the [start](#step-5-start-the-new-version) and [stop](#step-3-stop-the-old-version) are
the same.

## Data and backups

| Data | On the server | Backed up to S3 | If missing |
| --- | --- | --- | --- |
| Control DB: sites, identities, the admin, settings | `data/db/controldb.sqlite` | Every 30 seconds (`CONTROL_BACKUP_INTERVAL_IN_SECONDS`), and on stop | Every site: the server starts new, in setup |
| Site DBs: readings, treatments and settings of each site | `data/db/site/<batch>/<site>.sqlite`, a batch for each first letter | Each batch every 5 minutes (`BATCH_BACKUP_INTERVAL_IN_SECONDS`), and on stop | The start stops until restored ([5.8](#58-restore-missing-sites)) |
| Jobs DB: queued and past background jobs | `data/db/jobsdb.sqlite` | No | Queued jobs (they start again) and history |
| Logs DB: events, errors, logs | `data/db/logsdb.sqlite` | No | The control plane history |
| Live DB: data kept only while running (such as alarms) | `data/db/livedb.sqlite` | No | Active alarms, until new readings make them again |
| Local backups of the control and site DBs | `data/backup/` | They are uploaded | Nothing, while the DBs are there |
| `update.sh` copies of `./data` | `backups/data-<time>.tar.gz`, the newest 5 | No | The way back to before a release |
| `KEY_SETS` | `nighttroop.env` | No. Keep it in a password manager. | Every encrypted credential and S3 backup |

`KEY_SETS` encrypts S3 backups. With bucket versioning, S3 keeps replaced backups 90 days (control) or 30 days (sites).

## Production update

### Step 0: Release a version

1. Push a tag (such as `v1.2.3`) on a commit on `main`.
2. The **Docker image** workflow (`.github/workflows/docker.yml`) builds the default `Dockerfile` target.
3. It pushes `ghcr.io/grahamjenson/nighttroop:v1.2.3` and `:latest`.

The build downloads the DB-IP city database. Without it, the image knows only countries.

| Problem | Result | Fix |
| --- | --- | --- |
| The build fails | No image. Servers keep their version. | Fix it on `main`, and tag a new version. Never move a pushed tag. |
| The tag is on a commit that failed CI, or not on `main` | The image builds | Do not deploy it. Tag a fixed commit. **Gap:** Nothing stops the build. |
| `:latest` moves to an untried version | `update.sh` with no image argument deploys it | Deploy a version tag, never `:latest`. |
| The city database download fails | Sign-ins show only a country | None. The next build tries again. |

### Step 1: Check and pull the image

On the VPS, run `sudo ./update.sh ghcr.io/grahamjenson/nighttroop:v1.2.3`. The script:

1. Checks that it runs as root, to copy `./data` with its owner.
2. Checks that Nighttroop runs.
3. Pulls the new image, while the old version serves.
4. If that image runs now, exits with 0.

Each failure here is safe. The script stops and changes nothing:

| Problem | Fix |
| --- | --- |
| Not root | Use `sudo`. |
| Nighttroop does not run | Find why (`docker compose logs nighttroop`) first. Start it with `docker compose up -d`. |
| The pull fails: wrong tag, registry down, no network | Check the tag on the Packages page of the repository, and try again. |

### Step 2: Keep the running image

The script tags the running image `nighttroop-rollback:<time>`, to start it again whatever its tag then points to.

### Step 3: Stop the old version

`docker compose stop` sends `SIGTERM`, and kills the server after `stop_grace_period` (2 minutes). The server:

1. Stops HTTP requests and closes every Socket.IO connection. Requests in progress get 5 seconds.
2. Stops background jobs (syncs, alarms, backups). Running jobs get up to 1 minute.
3. Backs up the control DB locally and uploads it to S3.
4. Backs up and uploads each batch of site DBs with changes since its last upload.
5. Closes every DB.

**Note:** Nighttroop is down from here until it is healthy in [step 6](#step-6-wait-for-a-healthy-server), usually
less than 1 minute. Caddy answers with a `502`.

| Problem | Result | Fix |
| --- | --- | --- |
| S3 is unreachable | The last uploads fail (logged). The DBs on disk are whole. | None for this server. Do not move to a new VPS until the new version backs up to S3. |
| A job or the uploads take more than 2 minutes | Docker kills the server. The SQLite write-ahead log keeps the DBs whole. | As above. If frequent, increase `stop_grace_period`. **Gap:** No record of an unfinished last backup. |
| A phone or app uploads during the stop | The upload fails | Loop, AAPS, Trio and xDrip+ send it again later. **Gap:** Not documented or tested: which apps retry, how far back Dexcom, Libre and Nightscout syncs fetch. |

### Step 4: Copy `./data`

With nothing writing, the script copies `./data` to `./backups/data-<time>.tar.gz`. A rollback puts this copy back.

| Problem | Result | Fix |
| --- | --- | --- |
| The disk is full, or the copy fails | **The script stops, and Nighttroop stays stopped**, on the old image (`.env` is not changed yet) | Free space (`du -sh /opt/nighttroop/*`, old `data-failed-*` folders, `docker system prune`). Then `docker compose up -d nighttroop`. **Gap:** The script does not restart it, or check space first. |

### Step 5: Start the new version

The script writes the new image into `.env` and runs `docker compose up -d`. The server does 5.1 to 5.11 in order.

- A failed step ends the process with its error. Docker starts it again (`restart: unless-stopped`).
- A start that always fails restarts until `update.sh` stops waiting ([step 6](#step-6-wait-for-a-healthy-server))
  and rolls back.

#### 5.1 Read the settings

From `nighttroop.env` and the compose file ([docs/config.md](config.md)).

- **Failure:** `cannot create server config`. A bad value, or a missing necessary setting (such as `KEY_SETS`,
  `OPERATOR_NAME`, `SUPPORT_EMAIL`). The public `KEY_SETS` of `.env.dev`, and `minioadmin` S3 keys, are bad values.
- **Fix:** Correct `nighttroop.env`, and update again. A release that needs a new setting must say so.

#### 5.2 Start S3

If `S3_URL` is set, the server uploads a test object. It tries again after 2, 4, 8 and 16 seconds.

- **Failure:** `S3_URL is set but S3 is not working`, after about 30 seconds. Thus a server never starts empty and
  uploads over its backups.
- **Fix:** For an outage, wait. For wrong keys or a removed bucket, fix `nighttroop.env`. The rollback fails the same
  way, so `update.sh` exits with 2. When S3 works, run `docker compose up -d`.
- Without `S3_URL`, S3 is off: no restore, no backups. The control plane shows this.

#### 5.3 Remove leftovers

The server removes staging folders that a killed restore or archive left next to the DBs. This never stops the start.

#### 5.4 Restore the control DB if missing

Only if `data/db/controldb.sqlite` is not there, the server downloads the latest control DB backup from S3.

- **Failure:** The download fails. The start stops.
- **Warning:** Stop a server that starts in setup with site files on disk. It knows none of its sites.
  - Cause: the file is missing, and **there is no backup, or S3 is off**.
  - It starts as a **new server, in setup**, with the site files still in `data/db`.
  - Its health check passes, so `update.sh` keeps it.
- **Fix:**
  1. Before anyone sets it up, run `docker compose stop nighttroop`.
  2. Get `data/db/controldb.sqlite` from the newest `backups/data-*.tar.gz`:
     `tar -xzf <copy> data/db/controldb.sqlite`.
  3. Start it again. Do not run setup.
- **Gap:** Nothing stops a start with site DBs and an empty control DB.

#### 5.5 Restore missing batches

For each missing batch folder, the server downloads the latest batch backup. It never downloads over a folder.

- **Failure:** A download fails. The batch has no folder, so the next start tries again.
- **Fix:** As for [5.2](#52-start-s3).

#### 5.6 Open the control, jobs, logs and live DBs

The server opens each and runs its new migrations.

- **Failure:** `cannot create control db`, with the DB and migration. Its transaction rolls back a failed migration,
  so the DB is as before, marked dirty. The next open forces it back a version and runs the migration again
  ([docs/migrations.md](migrations.md)).
- **Fix:** `update.sh` rolls back, with the DB from the [step 4](#step-4-copy-data) copy. Fix the migration and
  release again.

#### 5.7 Open every site

The server opens each live and disabled site DB, on all cores together, and runs new migrations. It records sites with
a missing file for 5.8. Time: about 2 ms a site.

- **Failure:** `cannot fill sites`. One bad site stops the full start, intentionally: no site is served half migrated,
  and nothing is written that a rollback would lose.
- **Fix:** `update.sh` rolls back. To find the site and cause, read the failed version logs (`update.sh` prints them),
  and the site file in `data-failed-<time>`.
- **Gap:** Without a shell and `sqlite3`, you cannot see why one site fails, try a fix on a copy, or remove one site
  from the start (disable it, restore it from a backup, force its migration back).

#### 5.8 Restore missing sites

The server restores each live site with a missing file from its latest S3 batch backup, and opens it. Each is a
`site_db_missing` event.

- **Failure:** `live sites' DBs are missing and were not restored`, with the sites. S3 is off, the backup does not
  have the site, the download fails, or a restored site does not open.
- **Fix:** Get the site file from a copy in `./backups`. Or decide the site is gone, and delete it in the control DB.
- **Gap:** The delete needs the server to run, and it cannot.

#### 5.9 Make the admin

Development and staging make their admin. A server with no admin prints a setup code and serves only the setup page.

#### 5.10 Start the background jobs

Syncs, backups, alarms and other jobs start before requests. From here, a sync can write to a site DB.

#### 5.11 Accept requests

The server listens on port 3000. `/health` answers `ok` when the control DB answers.

- **Failure:** The port is not available. The process exits.
- **Gap:** Background jobs can write before this, and the rollback copy loses those writes. They are only syncs, which
  fetch the data again.

### Step 6: Wait for a healthy server

`update.sh` waits up to `HEALTH_TIMEOUT` (default 180 seconds) for the Docker health check. The check asks `/health`
every 10 seconds, after a 120-second start period. It is unhealthy after 3 failures.

| Problem | Result | Fix |
| --- | --- | --- |
| The start fails ([step 5](#step-5-start-the-new-version)) | Never healthy. Rollback ([step 7](#step-7-roll-back-if-not-healthy)). | Read the printed logs. Fix the cause. |
| A good version starts slowly: big restore, many sites | Not healthy in 180 seconds (or 150 for Docker). A working version rolls back. | `sudo HEALTH_TIMEOUT=600 ./update.sh ...`. **Gap:** The image fixes the Docker start period at 120 seconds. |
| Healthy but broken: a page errors, syncs fail, a query is wrong | The script keeps it | [Fix forward](#after-a-version-serves-fix-forward). **Gap:** `/health` checks only the control DB, not a page, site or sync. |

If healthy, the script deletes all but the newest 5 `./data` copies and previous images, and exits with 0.

### Step 7: Roll back if not healthy

The script:

1. Prints the last 50 log lines of the new version.
2. Stops it, and moves its `./data` to `./data-failed-<time>`.
3. Puts back `./data` from the step 4 copy.
4. Starts the previous image, and waits until it is healthy.
5. Exits with 1 if the previous version is back, or 2 if not.

A rollback loses nothing: a version that is never healthy took no requests. Delete `data-failed-*` folders manually.

| Problem | Result | Fix |
| --- | --- | --- |
| The previous version is not healthy (exit 2) | Nighttroop is down | Not the release: S3, a setting, the disk, the VPS. Read `docker compose logs nighttroop`, fix it, and `docker compose up -d`. |
| The new version uploaded backups before it failed | S3 has backups in the new format | The previous version replaces them in 5 minutes. Do not restore a new VPS from S3 before then. |

### After a version serves: fix forward

Earlier versions cannot open DBs that a newer version migrated. After a new version serves, release a fix and update
again.

- Last resort only: roll back manually from the copy made before that version, and lose all data written since. See
  **Rolling back by hand** ([Hostinger](deploy/hostinger/hostinger.md#rolling-back-by-hand)).
- **Gap:** Nothing keeps the data written after that copy, to move it into the older DBs.

## Failures on any deploy

| Problem | Result | Fix |
| --- | --- | --- |
| `KEY_SETS` is changed, cut short or lost | Encrypted credentials (Dexcom, Nightscout) and S3 backups cannot be decrypted | Put back the old value from the password manager. **Gap:** No way to move to new keys yet (TODOs: *Testing deploy with rolling to a new keyset*). |
| A migration on `main` was edited, not added, after the first production deploy | DBs past it never get the change. Old and new DBs differ. | CI refuses it when `data/sql/MIGRATIONS_FROZEN` exists. **Gap:** It does not exist yet (TODOs P0). |
| A migration is slow on big tables | The start takes more than `HEALTH_TIMEOUT` and rolls back | Increase `HEALTH_TIMEOUT`, or make the migration cheaper. **Gap:** Nothing measures a migration on a large DB before it ships. |
| The VPS stops during a deploy | The current step is half done | Start the VPS, then `docker compose up -d`. If `./data` is damaged, put back the newest `backups/data-*.tar.gz`. |
| Two servers use one bucket (such as a new VPS while the old one runs) | Each overwrites the other's backups | Stop the old server, with its last uploads, before you start the new one. |
| A new VPS restores from S3 | If the old server stopped before its last backup: control DB up to 30 seconds old, batches up to 5 minutes. Jobs, logs and live DBs start empty. | Expected. Sites made in those seconds are in the control DB but not a batch: the start stops ([5.8](#58-restore-missing-sites)). |

## Staging deploy

On each push to `main`, the **Deploy staging** workflow (`.github/workflows/deploy-staging.yml`) runs `flyctl deploy`
with the `staging` `Dockerfile` target. It does not wait for CI. One deploy runs at a time, then the newest waiting
push.

1. **Fly builds the image.** If this fails, the old version continues.
2. **Fly stops the old machine**, because the volume can be on only one machine. The server stops as in
   [step 3](#step-3-stop-the-old-version), with no S3.
3. **Fly starts the new machine** with `staging_start`, which:
   1. Gives `/data` to the server user.
   2. Runs `try_migrate_or_delete`: it opens and migrates each DB, checks its schema matches a new DB, and prepares
      its queries. **If one fails, it deletes every DB**, and the server starts empty.
   3. Starts the server as in [step 5](#step-5-start-the-new-version), with no S3.
4. **Fly waits for `/health`**, after 30 seconds of grace.

| Problem | Result | Fix |
| --- | --- | --- |
| A migration was changed | Every DB is deleted, with the testers' sites | Expected on staging. The logs say `deleting every database, as one cannot be migrated`. |
| The server cannot start: missing secrets, a wrong `HOST_NAME` | The health check fails. Staging is down. | See [Troubleshooting](deploy/staging_flyio/README.md#troubleshooting). Fix it on `main`, or set the secret. The next deploy fixes staging. **Gap:** Not documented if Fly goes back to the old version. |
| `FLY_API_TOKEN` expired | The workflow fails. Staging keeps the old version. | Make a new token ([staging guide, step 5](deploy/staging_flyio/README.md#5-deploy-every-commit-on-main)). |

## Gaps

Failures with no fix except a shell on the server, or none. Choose the tools from this list. Details are in each step.

1. **A site stops the start**: no tools to inspect, fix or remove it
   ([5.7](#57-open-every-site), [5.8](#58-restore-missing-sites)).
2. **A missing control DB starts a new server** that passes its health check
   ([5.4](#54-restore-the-control-db-if-missing)).
3. **No check after a deploy except `/health`** ([step 6](#step-6-wait-for-a-healthy-server)).
4. **A failed copy leaves Nighttroop stopped**, with no space check first ([step 4](#step-4-copy-data)).
5. **Client behaviour during the stop**: which apps send again, how far back syncs fetch
   ([step 3](#step-3-stop-the-old-version)).
6. **Fix forward is the only way back** ([fix forward](#after-a-version-serves-fix-forward)), and nothing measures
   migrations on large DBs first.
7. **The migrations are not frozen**, because there is no production deploy yet.
8. **No way to move to a new `KEY_SETS`.**
9. **A slow start rolls back**: the Docker start period is fixed ([step 6](#step-6-wait-for-a-healthy-server)).
10. **Background jobs start before requests**, so a failed start can have written data
    ([5.10](#510-start-the-background-jobs)).
11. **Fly behaviour when a staging deploy fails its health check** ([staging](#staging-deploy)).
