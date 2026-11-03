# A staging Nighttroop on Fly.io

This guide runs a staging Nighttroop on [Fly.io](https://fly.io) for alpha testing. It is open to everyone, so testers can
try the newest code.

- **It runs every commit on `main`.** Each time CI passes on `main`, a GitHub Actions workflow builds that commit and
  deploys it.
- **It keeps its data until a migration change breaks it.** The databases are on a volume, so they last through
  deploys and restarts. When a new version changes a migration so an old database can no longer be migrated, that deploy wipes
  them and starts empty, so a tester never hits data the new version cannot read. Nothing is backed up.
- **It says so on every page.** It runs in `staging` mode, which puts a large banner at the top of every page, so
  nobody mistakes it for a real site or keeps anything they need on it.
- **Its control plane never needs setting up.** Staging makes the control admin from two secrets you set once, so you
  sign in with the same password and authenticator code after every deploy, even one that wipes the data.

The files this guide uses:

| File | What it is |
| --- | --- |
| [`fly.toml`](fly.toml) | The Fly.io config for the staging server |
| [`.github/workflows/deploy-staging.yml`](../../../.github/workflows/deploy-staging.yml) | The workflow that deploys each commit on `main` CI passed on |

## Contents

1. [How it works](#how-it-works)
2. [What you need](#what-you-need)
3. [Create the app](#1-create-the-app)
4. [Configure it](#2-configure-it)
5. [Deploy it the first time](#3-deploy-it-the-first-time)
6. [Point your domain at it](#4-point-your-domain-at-it)
7. [Deploy every commit on main](#5-deploy-every-commit-on-main)
8. [Sign in to the control plane](#6-sign-in-to-the-control-plane)
9. [Day to day](#day-to-day)
10. [Troubleshooting](#troubleshooting)

## How it works

Nighttroop serves several kinds of host names from one domain. Here the staging server uses `staging.example.com`, a subdomain
of your own domain:

| Host | What it serves |
| --- | --- |
| `staging.example.com` | The public page, where testers create sites |
| `control.staging.example.com` | The control plane, for you |
| `<site>.staging.example.com` | Each site |
| `<site>.ns.staging.example.com` | Each site's Nightscout-compatible API, for apps like AAPS and Loop |

- **One machine.** It runs the image built from the repository's `Dockerfile`, always on.
  - Fly's proxy ends HTTPS and forwards to the server on port 3000.
  - Nighttroop reads the visitor's IP from Fly's `Fly-Client-IP` header, because Fly sets `FLY_APP_NAME`.
  - There is only one machine because each machine would have its own databases.
- **A volume.** The databases are in `/data`, on the `nighttroop_staging_data` volume, which the first deploy makes.
  - The image is the `Dockerfile`'s `staging` target. It starts as root, as Fly mounts the volume owned by root, gives
    `/data` to the server's user, and runs the server as that user.
  - Before the server starts, `try_migrate_or_delete`
    ([`scripts/try_migrate_or_delete`](../../../scripts/try_migrate_or_delete/try_migrate_or_delete.go)) opens each
    database as the server does: it migrates it and prepares its queries. As there can be thousands of sites, a site
    database's queries are only prepared when no other with the same schema has had them prepared.
  - If any cannot be opened, as when a migration was edited rather than added, it deletes them all, `/data/db` and
    `/data/backup`, and the server starts empty. The logs say `deleting every database, as one cannot be migrated`,
    with the database and why.
  - **Each deploy stops the server for a minute or so.** The volume can only be attached to one machine, so Fly stops
    the old one before the new one starts, opens every database, and passes its health check. Pages and apps get
    errors until it is back.
- **No S3.** `S3_URL` is left unset, so the S3 check fails at startup. Then:
  - nothing is restored on start and nothing is backed up;
  - the control plane warns that S3 is not connected, which is expected here;
  - sites cannot be deleted, as deleting archives a site to S3 first. The next deploy removes them anyway.
- **Small limits.**
  - [`fly.toml`](fly.toml) caps the server at 500 sites.
  - A site that is never used is deleted after a day.
  - The usual rate limits apply.

## What you need

- **A Fly.io account** with a payment method. One small, always-on machine costs a few dollars a month.
- **[`flyctl`](https://fly.io/docs/flyctl/install/)** installed and logged in (`fly auth login`).
- **A domain whose DNS you can change**, for `staging.example.com` and its wildcards.
  - A `fly.dev` name won't work, as Nighttroop needs wildcard subdomains.
- **Go**, to make the keys (or any machine with this repository checked out).
- **Admin access to this GitHub repository**, to add a variable and a secret.

## 1. Create the app

Pick a name for the app. Fly app names are global, e.g. `nighttroop-staging-yourname`. Then create it without deploying,
from the repository's root:

```sh
fly apps create nighttroop-staging-yourname
```

## 2. Configure it

1. **Edit [`fly.toml`](fly.toml):**
   - Change `HOST_NAME` and the health check's `Host` to your staging domain. They must match, as the server only answers
     its own host name.
   - Set `primary_region` to the [region](https://fly.io/docs/reference/regions/) nearest your testers.
   - Commit the change to `main`.
2. **Make the keys and store them as a secret.**
   - The server won't start without `KEY_SETS`.
   - As nothing is kept, losing them loses nothing: make new ones whenever you like.

   ```sh
   fly secrets set --app nighttroop-staging-yourname --stage KEY_SETS="$(go run scripts/keysets/create_keysets.go)"
   ```

3. **Choose the control admin's password and TOTP secret, and store them as secrets.**
   - Staging won't start without them. They are the control admin's for good, so save both in your password manager
     and add the TOTP secret to your authenticator app.
   - The password needs at least 16 characters, and the TOTP secret at least 16 base32 characters (`A`–`Z`, `2`–`7`).
     This makes both and prints them to save, then sets them:

   ```sh
   PASSWORD="$(openssl rand -base64 24)"
   TOTP_SECRET="$(python3 -c 'import base64,os; print(base64.b32encode(os.urandom(20)).decode())')"
   echo "Control admin password: $PASSWORD"; echo "TOTP secret: $TOTP_SECRET"
   fly secrets set --app nighttroop-staging-yourname --stage \
     STAGING_ADMIN_PASSWORD="$PASSWORD" STAGING_ADMIN_TOTP_SECRET="$TOTP_SECRET"
   ```

   - Never put them in `fly.toml`: the server is open to everyone, and this is its control plane.

4. **Optionally**, add more settings from [the config reference](../../config.md):
   - non-secret ones in `fly.toml`'s `[env]`;
   - secret ones with `fly secrets set --stage`, e.g. `SENTRY_DSN`, to see the testers' errors.

## 3. Deploy it the first time

From the repository's root:

```sh
fly deploy . --app nighttroop-staging-yourname --config docs/deploy/staging_flyio/fly.toml --dockerfile Dockerfile \
  --build-target staging --ha=false
```

- **The leading `.`** makes the repository's root the build context. The `--config` and `--dockerfile` paths are
  relative to it.
- **`--build-target staging`** builds the staging image, which checks the databases on the volume before the server starts.
- **`--ha=false`** keeps it to one machine. Fly makes two by default, and each would have its own databases.

Check it is up:

- [ ] `fly status --app nighttroop-staging-yourname` shows one machine, `started`, with its check passing.
- [ ] `fly logs --app nighttroop-staging-yourname` shows the server listening, in `Staging` mode.

## 4. Point your domain at it

1. **Give the app IP addresses:**

   ```sh
   fly ips allocate-v4 --shared --app nighttroop-staging-yourname
   fly ips allocate-v6 --app nighttroop-staging-yourname
   fly ips list --app nighttroop-staging-yourname
   ```

2. **Add DNS records at your DNS provider.** Point `staging.example.com`, `*.staging.example.com` and `*.ns.staging.example.com`
   at those addresses: an `A` record to the IPv4 address and an `AAAA` record to the IPv6 address for each.
3. **Ask Fly for certificates.** A wildcard covers only one level, so `*.ns.staging.example.com` needs its own. Quote the
   wildcards so the shell leaves the `*` alone:

   ```sh
   fly certs add staging.example.com --app nighttroop-staging-yourname
   fly certs add "*.staging.example.com" --app nighttroop-staging-yourname
   fly certs add "*.ns.staging.example.com" --app nighttroop-staging-yourname
   ```

4. **Prove you own the wildcards.** Let's Encrypt only issues a wildcard to whoever controls the domain's DNS.
   1. Run `fly certs show "*.staging.example.com"` (and the same for `*.ns.staging.example.com`). It shows a `CNAME` for
      `_acme-challenge.staging.example.com` (and `_acme-challenge.ns.staging.example.com`).
   2. Add those records exactly as shown.
   3. Remove any old `TXT` record at the same `_acme-challenge` names.
   4. If your DNS provider proxies records, e.g. Cloudflare, don't proxy these.
5. **Wait until every certificate is ready.** Run `fly certs list --app nighttroop-staging-yourname` until each says
   `Ready`, usually a few minutes.

Check it works:

- [ ] `curl https://staging.example.com/health` prints `{"status":"ok"}`.
- [ ] <https://staging.example.com> shows the public page with the **THIS IS A TEST SITE** banner.

## 5. Deploy every commit on main

The **Deploy staging** workflow ([`.github/workflows/deploy-staging.yml`](../../../.github/workflows/deploy-staging.yml)) runs each
time CI finishes on a push to `main`.

- **When CI passed,** it checks out that commit and deploys it with the same command as above, building the image on
  Fly's builders.
- **When CI failed,** that commit is not deployed, so testers keep the last version that passed.
- **When a newer commit arrives mid-deploy,** it cancels the older deploy.
- **It does nothing until you turn it on:**
  1. Make a deploy token for the app:

     ```sh
     fly tokens create deploy --app nighttroop-staging-yourname
     ```

  2. In the repository's **Settings → Secrets and variables → Actions**:
     - add the secret **`FLY_API_TOKEN`**, holding the whole token, including its `FlyV1 ` prefix;
     - add the variable **`STAGING_FLYIO_APP`**, holding the app's name, e.g. `nighttroop-staging-yourname`.
  3. To deploy now, without waiting for a commit, run **Deploy staging** from the repository's **Actions** tab
     (**Run workflow**). A manual run deploys the head of `main`.

To stop deploying, delete the `STAGING_FLYIO_APP` variable. To stop the server too, run
`fly scale count 0 --app nighttroop-staging-yourname`, or `fly apps destroy nighttroop-staging-yourname` to remove it.

## 6. Sign in to the control plane

Testers don't need it; you only need it to see the server's sites, jobs and logs.

1. Open <https://control.staging.example.com>.
2. Sign in as `admin`, with `STAGING_ADMIN_PASSWORD` and a code from the authenticator app you gave
   `STAGING_ADMIN_TOTP_SECRET`.

The admin is made whenever the server starts without one, so it is the same after every deploy, even one that wipes
the databases. Changing the secrets doesn't change an admin the server already has: wipe it by hand (see below) to make
the admin again from the new ones.

## Day to day

- **Tell testers what to expect:**
  - a new version that changes the databases' migrations wipes everything;
  - when it does, a phone's Nightscout URL and token stop working, so testers create a new site and set their apps up
    again.
- **Testing with realistic data:** testers can turn on virtual sync in a site's sync settings, instead of syncing a
  real Dexcom, Libre or Nightscout account.
- **What is running:** `fly releases --app nighttroop-staging-yourname` lists each deploy, newest first. The workflow run on
  the **Actions** tab names its commit.
- **Wiping it by hand:** remove the databases, then restart the server, which starts empty:

  ```sh
  fly ssh console --app nighttroop-staging-yourname -C "rm -rf /data/db /data/backup"
  fly machine restart --app nighttroop-staging-yourname
  ```
- **If it is struggling:**
  - the control plane's usage page shows memory and CPU;
  - `fly scale memory 1024 --app nighttroop-staging-yourname` gives it more memory.

## Troubleshooting

### The deploy hangs, then fails its health check

- `HOST_NAME` and the check's `Host` in `fly.toml` must be the same, or every check is refused.
- `fly logs` says why the server stopped. Most often `KEY_SETS` is missing or was cut short: set it again as in
  [step 2](#2-configure-it).

### A certificate stays `Awaiting configuration`

- `fly certs show <host>` says which record is missing.
- For a wildcard it is nearly always the `_acme-challenge` `CNAME`: check it matches exactly, is not proxied, and has no
  `TXT` record beside it.

### Testers' data is gone

A deploy wipes the databases when a migration change means one can no longer be migrated.

- `fly logs` says `deleting every database, as one cannot be migrated`, with the database and why.
- `fly releases` shows which deploy it was.
- `fly volumes list` shows whether the volume is still there; a machine made without it starts empty every time.

### There are two machines, and pages show different data

Each machine has its own databases. Remove the extra one with `fly scale count 1 --app nighttroop-staging-yourname`, and
deploy with `--ha=false`, as the workflow does.

### The Deploy staging workflow is skipped

- It runs only after CI passes on a push to `main`, and only when the `STAGING_FLYIO_APP` variable is set.
- A run that fails with `unauthorized` needs a new `FLY_API_TOKEN`, from [step 5](#5-deploy-every-commit-on-main).
