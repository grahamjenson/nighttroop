# Staging Nighttroop on Fly.io

A staging Nighttroop on [Fly.io](https://fly.io) for alpha testing. It is open to everyone, so testers can try the
newest code.

## Overview

- **Deploys:** each push to `main`, without a wait for CI ([step 5](#5-deploy-every-commit-on-main)).
- **Data:** on a volume, through deploys and restarts. No backups. A migration change wipes it
  ([database check](#volume-and-database-check)).
- **Banner:** `staging` mode puts a large banner on every page, so nobody thinks it is a real site or keeps necessary
  data on it.
- **Control admin:** made from two secrets that you set once. The sign-in is the same after every deploy and wipe.

| File | What it is |
| --- | --- |
| [`fly.toml`](fly.toml) | The Fly.io config for the staging server |
| [`.github/workflows/deploy-staging.yml`](../../../.github/workflows/deploy-staging.yml) | The workflow that deploys each push to `main` |

## How it works

### Host names

Staging uses `staging.nighttroop.nz`, a subdomain of your own domain:

| Host | What it serves |
| --- | --- |
| `staging.nighttroop.nz` | The public page, where testers create sites |
| `control.staging.nighttroop.nz` | The control plane, for you |
| `<site>.staging.nighttroop.nz` | Each site |
| `<site>.ns.staging.nighttroop.nz` | The Nightscout-compatible API of each site, for apps like AAPS and Loop |

### Machine

- **One machine**, because each machine has its own databases.
- **Image:** from the repository `Dockerfile`.
- **Size:** one shared CPU, 256 MB of memory, 512 MB of swap.
- **Memory (measured with 500 sites):** about 50 MB idle, plus 1.5 MB for each site used in the last 30 seconds. Thus
  about 130 sites in use at once before it swaps.
- **Auto stop:**
  - Fly stops it after a few idle minutes, and starts it on the next request. The data stays on the volume.
  - That request waits a couple of seconds for 500 sites, plus the Fly machine boot.
  - While stopped, nothing runs: Dexcom, Libre and Nightscout syncs, the virtual connector, inactive site deletion,
    resource checks. They run again when it starts.
  - A phone that uploads every few minutes, or an open page, keeps it running.

### Network and visitor IP

- The Fly proxy ends HTTPS and forwards to port 3000.
- Fly sets `FLY_APP_NAME`, so Nighttroop reads the visitor IP from the `Fly-Client-IP` header. It reads it only from
  `TRUSTED_PROXIES`, which [`fly.toml`](fly.toml) sets to `172.16.0.0/16`.
- Other sources (such as another app or machine on your organisation's private network) could set any IP and avoid
  the per-IP rate limits. Nighttroop ignores their header.
- Fly does not document where its proxy connects from. It was seen at `172.16.x.x`. Other private network machines have
  `fdaa:` addresses. To check, see [The visitors all share one IP](#the-visitors-all-share-one-ip).

### Volume and database check

- **Volume:** `/data` holds the databases, on the `nighttroop_staging_data` volume. The first deploy makes it.
- **Start:** the `staging` `Dockerfile` target starts as root, because Fly mounts the volume owned by root. It gives
  `/data` to the server user, and runs the server as that user.
- **Check:** before the server starts,
  `try_migrate_or_delete` ([`cmd/try_migrate_or_delete`](../../../cmd/try_migrate_or_delete/try_migrate_or_delete.go))
  opens each database as the server does. It migrates it, checks its schema against a new database, and prepares
  queries for only the first site database (thousands of sites share one schema).
- **Wipe:** if one does not open or has a different schema, it deletes all (`/data/db`, `/data/backup`). The server
  starts empty.
  - Thus testers never see data the new version cannot read, or rules it no longer has.
  - Cause: a migration was edited, not added, even only to change a `CHECK` that every query still works with.
  - The logs say `deleting every database, as one cannot be migrated`, with the database and why (such as a schema
    statement that a new one does not have).
- **Downtime:** each deploy stops the server for about a minute. The volume attaches to one machine only, so Fly stops
  the old one before the new one starts, opens every database and passes its health check. Pages and apps get errors
  until then.

### No S3 and limits

- `S3_URL` is not set, so S3 is off:
  - Nothing is restored or backed up.
  - The control plane warns that S3 is not connected. This is expected.
  - Sites cannot be deleted, because a delete first archives to S3. The next deploy removes them anyway.
- [`fly.toml`](fly.toml) caps the server at 500 sites.
- A site that is never used is deleted after a day.
- The usual rate limits apply.

## Requirements

- **A Fly.io account** with a payment method. One small auto-stop machine and its 1 GB volume cost a few dollars a
  month at most.
- **[`flyctl`](https://fly.io/docs/flyctl/install/)**, logged in (`fly auth login`).
- **A domain whose DNS you can change**, for `staging.nighttroop.nz` and its wildcards. A `fly.dev` name does not work:
  Nighttroop needs wildcard subdomains.
- **Go**, to make the keys (or any machine with this repository).
- **Admin access to this GitHub repository**, to add a variable and a secret.

## Install

### 1. Create the app

Choose an app name (Fly app names are global). From the repository root, create it without a deploy:

```sh
fly apps create nighttroop-staging
```

### 2. Configure the app

1. Edit [`fly.toml`](fly.toml), then commit it to `main`:

   | Setting | Value | Reason |
   | --- | --- | --- |
   | `HOST_NAME` and the health check `Host` | Your staging domain | They must match. The server answers only its own host name. |
   | `OPERATOR_NAME`, `SUPPORT_EMAIL` | Who runs the server, and how testers contact you | The terms and privacy policy name them. Required. |
   | `primary_region` | The [region](https://fly.io/docs/reference/regions/) nearest your testers | |

2. Store the keys as a secret. `KEY_SETS` is required. Staging keeps nothing, so make new keys at any time.

   ```sh
   fly secrets set --app nighttroop-staging --stage KEY_SETS="$(go run cmd/keysets/create_keysets.go)"
   ```

3. Store the control admin password (16 characters minimum) and TOTP secret (16 base32 characters minimum, `A`–`Z`,
   `2`–`7`) as secrets. Both are required. This makes, prints and sets them:

   ```sh
   PASSWORD="$(openssl rand -base64 24)"
   TOTP_SECRET="$(python3 -c 'import base64,os; print(base64.b32encode(os.urandom(20)).decode())')"
   echo "Control admin password: $PASSWORD"; echo "TOTP secret: $TOTP_SECRET"
   fly secrets set --app nighttroop-staging --stage \
     STAGING_ADMIN_PASSWORD="$PASSWORD" STAGING_ADMIN_TOTP_SECRET="$TOTP_SECRET"
   ```

4. Save both in your password manager. They stay the control admin values permanently.
5. Add the TOTP secret to your authenticator app.
6. Optional: add [config](../../config.md) settings. Non-secret ones go in the `fly.toml` `[env]`. Secret ones use
   `fly secrets set --stage` (such as `SENTRY_DSN`, to see the testers' errors).

**Warning:** Never put the password or TOTP secret in `fly.toml`. The server is open to everyone, and this is its
control plane.

### 3. Deploy the first time

From the repository root:

```sh
fly deploy . --app nighttroop-staging --config docs/deploy/staging_flyio/fly.toml --dockerfile Dockerfile \
  --build-target staging --ha=false --env DEPLOYMENT=nighttroop-staging
```

| Option | What it does |
| --- | --- |
| `.` | Makes the repository root the build context. The `--config` and `--dockerfile` paths are relative to it. |
| `--build-target staging` | The staging image, which checks the volume databases before the server starts |
| `--ha=false` | One machine. Fly makes two by default, each with its own databases. |
| `--env DEPLOYMENT=nighttroop-staging` | Names the Sentry environment after the app, apart from other staging servers. The workflow does the same. |

Check:

- [ ] `fly status --app nighttroop-staging` shows one machine, `started`, with its check passing.
- [ ] `fly logs --app nighttroop-staging` shows the server listening, in `Staging` mode.

### 4. Connect your domain

1. Give the app IP addresses:

   ```sh
   fly ips allocate-v4 --shared --app nighttroop-staging
   fly ips allocate-v6 --app nighttroop-staging
   fly ips list --app nighttroop-staging
   ```

2. At your DNS provider, add `A` (IPv4) and `AAAA` (IPv6) records to those addresses for `staging.nighttroop.nz`,
   `*.staging.nighttroop.nz` and `*.ns.staging.nighttroop.nz`.
3. Ask Fly for certificates. A wildcard covers one level only, so `*.ns.staging.nighttroop.nz` needs its own. Quote
   the wildcards, so the shell keeps the `*`:

   ```sh
   fly certs add staging.nighttroop.nz --app nighttroop-staging
   fly certs add "*.staging.nighttroop.nz" --app nighttroop-staging
   fly certs add "*.ns.staging.nighttroop.nz" --app nighttroop-staging
   ```

4. Prove that you own the wildcards (Let's Encrypt needs DNS control):
   1. Run `fly certs show "*.staging.nighttroop.nz"`, and the same for `*.ns.staging.nighttroop.nz`. Each shows a
      `CNAME` for `_acme-challenge.staging.nighttroop.nz` or `_acme-challenge.ns.staging.nighttroop.nz`.
   2. Add those records exactly.
   3. Remove old `TXT` records at those `_acme-challenge` names.
   4. If your DNS provider proxies records (such as Cloudflare), do not proxy these.
5. Run `fly certs list --app nighttroop-staging` until each says `Ready`, usually a few minutes.

Check:

- [ ] `curl https://staging.nighttroop.nz/health` prints `{"status":"ok"}`.
- [ ] <https://staging.nighttroop.nz> shows the public page with the **THIS IS A TEST SITE** banner.

### 5. Deploy every commit on main

The **Deploy staging** workflow ([`.github/workflows/deploy-staging.yml`](../../../.github/workflows/deploy-staging.yml))
runs on each push to `main`.

- **No wait for CI:** commits passed CI on their pull request, so staging gets them about 5 minutes sooner. CI still
  runs on `main`, and its README badge goes red on failure.
- **Build failure:** the deploy fails. Testers keep the last version that built.
- **Pushes during a deploy:** the deploy finishes. New pushes wait as one run, the newest replacing older ones. A burst
  of merges is thus one more deploy.

To turn it on:

1. Make a deploy token:

   ```sh
   fly tokens create deploy --app nighttroop-staging
   ```

2. In the repository **Settings → Secrets and variables → Actions**, add:
   - the secret **`FLY_API_TOKEN`**: the full token, with its `FlyV1 ` prefix;
   - the variable **`STAGING_FLYIO_APP`**: the app name (such as `nighttroop-staging`).
3. Optional: to deploy the head of `main` now, run **Deploy staging** on the **Actions** tab (**Run workflow**).

To stop:

- Deploys: delete the `STAGING_FLYIO_APP` variable.
- The server: `fly scale count 0 --app nighttroop-staging`.
- The app: `fly apps destroy nighttroop-staging`.

### 6. Sign in to the control plane

Only you need it, to see the sites, jobs and logs.

1. Open <https://control.staging.nighttroop.nz>.
2. Sign in as `admin`, with `STAGING_ADMIN_PASSWORD` and a code from the authenticator that has
   `STAGING_ADMIN_TOTP_SECRET`.

A server with no admin makes it at start, so it survives deploys and wipes. New secrets do not change an existing
admin: [wipe manually](#wipe-manually) to apply them.

## Operate

### Tell testers what to expect

- A new version that changes the database migrations wipes everything.
- After a wipe, phone Nightscout URLs and tokens stop working. Testers create a new site and set up their apps again.

### Test with realistic data

Testers can turn on the virtual connector (site Connectors page) instead of a real Dexcom, Libre or Nightscout sync.

### See what runs

`fly releases --app nighttroop-staging` lists deploys, newest first. Each workflow run on the **Actions** tab names its
commit.

### Wipe manually

Remove the databases, then restart. The server starts empty.

```sh
fly ssh console --app nighttroop-staging -C "rm -rf /data/db /data/backup"
fly machine restart --app nighttroop-staging
```

### Fix low performance

- Memory and CPU: the control plane usage page.
- More memory: `fly scale memory 512 --app nighttroop-staging`.
- Always on: set `auto_stop_machines = "off"` and `min_machines_running = 1` in `fly.toml`.

## Troubleshooting

### The deploy hangs, then fails its health check

- `HOST_NAME` and the check `Host` in `fly.toml` must match, or every check is refused.
- `fly logs` shows why the server stopped. Usually `KEY_SETS` is missing or cut short: set it
  ([step 2](#2-configure-the-app)).

### A certificate stays `Awaiting configuration`

`fly certs show <host>` shows the missing record. For a wildcard, it is nearly always the `_acme-challenge` `CNAME`.
Check that it matches exactly, is not proxied, and has no `TXT` record next to it.

### Testers' data is gone

A migration change [wiped the databases](#volume-and-database-check).

- `fly logs`: `deleting every database, as one cannot be migrated`, with the database and why.
- `fly releases`: which deploy.
- `fly volumes list`: if the volume is still there. A machine without it starts empty each time.

### There are two machines, and pages show different data

Each has its own databases. Run `fly scale count 1 --app nighttroop-staging`. Deploy with `--ha=false`, as the workflow
does.

### The visitors all share one IP

Symptom: rate limits hit too often, or one IP on every request. Cause: the Fly proxy connects from outside
`TRUSTED_PROXIES`, so Nighttroop ignores its `Fly-Client-IP`.

1. On the control plane logs page, find a `Captured` log of your request (`CAPTURE_REQUESTS` is on).
2. If its `remote_ip` is a Fly address, not yours, add it to `TRUSTED_PROXIES` in `fly.toml` and deploy.

**Warning:** Never add `fdaa::/16` or `0.0.0.0/0`. Other machines on your private network connect from there, and can
then claim any IP.

### The Deploy staging workflow is skipped

- It runs on each push to `main`, only when the `STAGING_FLYIO_APP` variable is set.
- A run that fails with `unauthorized` needs a new `FLY_API_TOKEN` ([step 5](#5-deploy-every-commit-on-main)).
