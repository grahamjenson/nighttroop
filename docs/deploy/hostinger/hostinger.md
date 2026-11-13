# Deploy Nighttroop to a Hostinger VPS

This guide installs Nighttroop with Docker on a [Hostinger](https://www.hostinger.com/vps-hosting) VPS. It is for
operators new to Hostinger who know SSH and a Linux shell.

## Overview

### Result

- [`hostinger.compose.yml`](hostinger.compose.yml) runs:
  - **Caddy** on ports 80 and 443. It gets and renews Let's Encrypt certificates automatically.
  - **Nighttroop** behind Caddy, on a private Docker network, with no ports of its own.
- **A data folder** holds the SQLite databases, optionally backed up to S3.
- [`update.sh`](../update.sh), the safe update script, backs up, updates, checks health, and rolls back if the new
  version is not healthy.

### Host names

All host names come from one domain (`example.com` in this guide):

- `example.com`: the public page, where people create sites.
- `control.example.com`: the control plane, for the control admin.
- `<site>.example.com`: each site.
- `<site>.ns.example.com`: the Nightscout-compatible API of each site, for apps like AAPS and Loop.

### DNS and certificates

- **DNS:** `example.com`, `*.example.com` and `*.ns.example.com` point to the VPS.
- **Certificates:** `example.com`, and wildcards `*.example.com` and `*.ns.example.com` for the subdomain of each site.
- **Wildcard proof:** Let's Encrypt gives wildcards only with proof of DNS control.
  - Caddy adds a TXT record through Hostinger's API (a DNS-01 challenge). It needs an API token for this.
  - Caddy is built with the community [`caddy-dns-hostinger`](https://github.com/sbrunk/caddy-dns-hostinger) module.

### Request flow

1. A browser or app connects to Caddy on port 443.
2. Caddy proxies the request to Nighttroop on port 3000, over the private Docker network.
3. Nighttroop uses the client IP in `X-Forwarded-For`, from its only trusted proxy (`TRUSTED_PROXIES`), Caddy
   (`172.30.0.2`). Thus rate limits apply to each visitor.

## Requirements

- **A Hostinger VPS.** The smallest KVM plan is sufficient to start. Nighttroop is built to be cheap per user.
  - Web and shared hosting cannot run Nighttroop, a long-running server with its databases on local disk.
- **A domain on Hostinger's nameservers,** for Caddy's DNS records. Domains bought from Hostinger have them. For
  others, see [step 3](#3-point-the-domain-at-the-vps).
- **GitHub access to this repository.** Its container registry has the image, and is private while the repository is.
- **S3-compatible storage** (optional, strongly recommended). Hostinger does not sell it. See
  [step 5](#5-choose-where-backups-go).
- **On your computer:** a clone of this repository, `ssh` and `scp`.

## Prepare the host

### 1. Set up the VPS

1. In [hPanel](https://hpanel.hostinger.com), buy a KVM VPS plan.
2. Select the **Docker** template (Ubuntu with Docker and Docker Compose).
   - Or select plain **Ubuntu 24.04**, and install Docker with
     [Docker's instructions for Ubuntu](https://docs.docker.com/engine/install/ubuntu/).
   - To change it later: **OS & Panel**, **Operating System**, search for Docker under **Change OS**.
     **Caution:** This erases the VPS.
3. Set a strong root password, and add your SSH public key.
4. Note the VPS IP on its **Overview** page (**VPS**, **Manage**). This guide uses `203.0.113.10`.
5. Log in and update the system:

   ```sh
   ssh root@203.0.113.10
   apt update && apt upgrade -y
   ```

6. Check that Docker Compose is 2.23.1 or later, for inline configs:

   ```sh
   docker compose version
   ```

This guide uses `docker compose` over SSH, not hPanel's Docker Manager, because the update script runs on the server.

### 2. Open the firewall

hPanel manages the firewall, outside the VPS.

**Warning:** Add all rules before you activate the firewall. It drops all traffic that no rule accepts.

1. In hPanel, open **VPS**, **Manage**, **Security**, **Firewall**.
2. Create a configuration with these **accept** rules, each from any source:
   - TCP 22: SSH. Add it first, so that you cannot lock yourself out.
   - TCP 80: HTTP, which Caddy redirects to HTTPS.
   - TCP 443: HTTPS.
   - UDP 443: HTTP/3.
3. Activate the configuration for your VPS.

**Warning:** Do not open port 3000. Nighttroop is reachable only through Caddy.

## Prepare the domain and storage

### 3. Point the domain at the VPS

#### If a different registrar has the domain

1. Add the domain to your Hostinger account, so that hPanel shows its DNS zone.
2. At the registrar, set the nameservers that hPanel shows. These are usually `ns1.dns-parking.com` and
   `ns2.dns-parking.com`, but some accounts get other `dns-parking.com` names.
3. Wait for the change to spread: up to one or two days, often much less.

#### Add the records

1. In hPanel, open **Domains**, your domain, **DNS / Nameservers**.
2. Remove `A`, `AAAA` and `CNAME` records named `@`, `*` or `ns` (conflicting parking records). Keep `MX` and `TXT`
   records if you use email on the domain.
3. Add `A` records `@`, `*` and `*.ns` to `203.0.113.10`, with TTL 3600.
4. If the VPS has an IPv6 address (on its **Overview** page), add the same records as `AAAA` records.
5. On your computer, check that each command prints the VPS IP:

   ```sh
   dig +short example.com
   dig +short anything.example.com
   dig +short anything.ns.example.com
   ```

You do not need `TXT` records for the certificates. Caddy adds and removes them.

### 4. Make a Hostinger API token

Caddy uses it to add the `_acme-challenge` TXT records that Let's Encrypt checks.

1. In hPanel, open your profile, **API** (<https://hpanel.hostinger.com/profile/api>).
2. Make a token with a long expiry, for example `nighttroop-caddy`.
3. Copy the token. Hostinger shows it only once.
4. Put a reminder in your calendar to replace the token before it expires. Without a valid token, Caddy cannot renew
   the certificates, and the site goes down when they expire.

**Warning:** Keep the token secret. It can manage your whole Hostinger account through the API, not only this DNS.

### 5. Choose where backups go

Nighttroop can back up its databases to S3-compatible storage, encrypted with your `KEY_SETS`. It restores from there
when it starts on an empty data folder, for example on a new VPS.

- **Without S3:** your only copies are the update script's local backups on the VPS, and Hostinger's VPS backups and
  snapshots.
- **With S3:** Hostinger does not sell object storage. [Cloudflare R2](https://developers.cloudflare.com/r2/) and
  [Backblaze B2](https://www.backblaze.com/cloud-storage) have free or cheap tiers.

To set up S3:

1. Make a private bucket, for example `nighttroop-backups`.
2. Make an access key that can only read and write that bucket.
3. Note the endpoint URL, bucket, access key ID and secret key for [step 7](#7-make-the-settings-files).

## Install

### 6. Put the files on the VPS

1. On the VPS, make the folder. `data` must belong to user ID `10001`, which the image runs Nighttroop as:

   ```sh
   mkdir -p /opt/nighttroop/data
   chown 10001:10001 /opt/nighttroop/data
   ```

2. On your computer, from your clone, copy the files:

   ```sh
   scp docs/deploy/hostinger/hostinger.compose.yml docs/deploy/update.sh root@203.0.113.10:/opt/nighttroop/
   ```

3. On GitHub, make a **classic** personal access token with only the `read:packages` scope (**Settings**,
   **Developer settings**, **Personal access tokens**, **Tokens (classic)**).
4. On the VPS, log in to GitHub's container registry, with the token as the password. This lets the VPS pull
   `ghcr.io/grahamjenson/nighttroop` while the repository is private.

   ```sh
   docker login ghcr.io -u <your-github-username>
   ```

## Configure

### 7. Make the settings files

Both settings files hold secrets. Make them readable by root only:

```sh
cd /opt/nighttroop
touch .env nighttroop.env
chmod 600 .env nighttroop.env
```

#### `.env`: settings for Docker Compose and Caddy

```sh
# So plain `docker compose` commands in this folder use the right file
COMPOSE_FILE=hostinger.compose.yml

# Your domain, without https:// or a trailing dot
DOMAIN=example.com

# Let's Encrypt emails this address about problems with your certificates
ACME_EMAIL=you@example.com

# The token from step 4
HOSTINGER_API_TOKEN=paste-the-token-here

# The Nighttroop image to run; update.sh changes this line
NIGHTTROOP_IMAGE=ghcr.io/grahamjenson/nighttroop:latest
```

#### `nighttroop.env`: settings for Nighttroop

The compose file sets the host name, HTTPS and the trusted proxy. This file holds only secrets and options.

1. Make the encryption keys with the tool in the image:

   ```sh
   docker run --rm --entrypoint create_keysets ghcr.io/grahamjenson/nighttroop:latest
   ```

2. Put the keys and your other settings in `nighttroop.env`:

   ```sh
   # The keys from create_keysets: one long line
   KEY_SETS=...

   # Required: who runs this server and how its users contact you, named in its terms and privacy policy
   OPERATOR_NAME=Your Name or Organisation
   SUPPORT_EMAIL=you@example.org

   # S3 backups (step 5); leave these out to run without them
   S3_URL=https://<account-id>.r2.cloudflarestorage.com
   S3_BUCKET=nighttroop-backups
   S3_ACCESS_KEY=...
   S3_SECRET_KEY=...

   # Optional: report errors to Sentry
   # SENTRY_DSN=...
   ```

3. Keep a copy of `KEY_SETS` off the VPS, for example in a password manager.

**Warning:** Do not lose `KEY_SETS`. It encrypts the secrets in the databases and the backups. Without it, you cannot
read the data or restore a backup onto a new VPS.

## Start

### 8. Start Nighttroop

1. Start the containers. The first start takes a few minutes: it compiles Caddy with the Hostinger DNS module
   (`--build`), pulls Nighttroop and gets the certificates.

   ```sh
   cd /opt/nighttroop
   docker compose up -d --build
   ```

2. Watch the Caddy logs for `certificate obtained successfully` for `example.com`, `*.example.com` and
   `*.ns.example.com`. This can take a minute or two while the DNS challenge record spreads. Then push Ctrl+C.

   ```sh
   docker compose logs -f caddy
   ```

3. Check that:
   - [ ] `docker compose ps` shows both containers running, and `nighttroop` as `healthy`.
   - [ ] `curl https://example.com/health` prints `{"status":"ok"}`.
   - [ ] <https://example.com> goes to the setup page at `https://control.example.com/setup`, with a valid padlock.
   - [ ] `http://example.com` redirects to HTTPS.

### 9. Set up the control plane

A new server has no control admin, so all pages go to the setup page. Setup needs the code from the server logs, so
nobody else can take the server first.

1. Find the setup code. The server logs it at each start until setup is complete, so restart Nighttroop to see it
   again.

   ```sh
   docker compose logs nighttroop | grep -A1 "SET UP NIGHTTROOP"
   ```

2. Open <https://control.example.com/setup> and enter the code.
3. Store the control admin's password and TOTP secret safely. The page shows them, with the username, only once.
4. Add the TOTP secret to your authenticator app (for example, scan its QR code).
5. Enter the code from the app, then push **Finish setup**.
6. Sign in at <https://control.example.com>. The rest of the server, such as site creation, now works.

## Updating Nighttroop safely

### Where new versions come from

Merges to `main` build no image. To release a version, push a tag on a commit that passed CI:

```sh
git tag v1.2.3 && git push origin v1.2.3
```

The **Docker image** workflow (`.github/workflows/docker.yml`) builds that tag and pushes it to
`ghcr.io/grahamjenson/nighttroop` as:

- **`:v1.2.3`**: that version. It never changes.
- **`:latest`**: the newest version.

Deploy version tags, so that you know which version runs and can go back to it. Find them on the repository's GitHub
**Packages** page, or in its tags.

### Running an update

On the VPS:

```sh
cd /opt/nighttroop
sudo ./update.sh ghcr.io/grahamjenson/nighttroop:v1.2.3
```

The script:

1. **Pulls the new image** while the old version runs, so a bad tag or registry problem changes nothing. Stops if that
   image runs already.
2. **Keeps the running image** as `nighttroop-rollback:<time>`, to start it again.
3. **Stops Nighttroop,** which uploads a last backup to S3 (two minutes allowed).
4. **Copies `./data`** to `./backups/data-<time>.tar.gz`, while nothing writes to it.
5. **Starts the new image** and waits up to 3 minutes for its health check, `/health`. This answers when the control
   database does.
   - The server serves only after it opens and migrates each live site's database. A release that cannot open one is
     never healthy, and takes no data.
6. **If healthy:** keeps the newest 5 backups and previous images, and deletes older ones.
7. **If not healthy, rolls back:**
   1. Prints the new version's last logs.
   2. Stops it, and moves its `./data` to `./data-failed-<time>`.
   3. Restores `./data` from step 4.
   4. Starts the previous image, and checks that it is healthy.

Nighttroop is down from step 3 until healthy in step 5, usually much less than a minute. Caddy runs all the time.

| Exit code | Meaning |
| --- | --- |
| `0` | Updated, or already running that image |
| `1` | The new version was not healthy. The previous one is back, with its data |
| `2` | Neither version is healthy. See `docker compose logs nighttroop` |

Options:

- `sudo HEALTH_TIMEOUT=300 ./update.sh ...` waits 5 minutes for the health check, not 3 (for example, for a large
  restore).
- `sudo KEEP_BACKUPS=10 ./update.sh ...` keeps 10 backups and previous images, not 5.

### What a rollback cannot undo

- **Data the new version wrote** is not in the restored `./data`. It stays in `./data-failed-<time>`. The script does
  not delete these folders: delete them when you do not need them.
- **S3 backups the new version made,** possibly in its new format. The restored version runs from local data, and its
  next backups replace them.

**Warning:** Do not move servers until the old version backs up again. Only a new VPS restores from S3.

### Fix forward once a new version has served

The script's rollback loses nothing: a version that is not healthy served nothing. After a new version served, do not
roll back. Update to a version with a fix. An earlier version cannot open databases that a newer one migrated. To go
back, you must restore an old `./data` copy, and you lose everything written since.

### Rolling back by hand

**Caution:** Do this only as a last resort. You lose everything written since the copy.

Restore the copy of `./data` made just before the earlier version was replaced. Its databases have the format that
version expects. The commands keep the current data in `data-before-rollback`.

```sh
cd /opt/nighttroop
docker compose stop nighttroop
mv data data-before-rollback
tar -xzf backups/data-<time>.tar.gz          # makes ./data again
docker images nighttroop-rollback            # the earlier images, by time
sed -i 's|^NIGHTTROOP_IMAGE=.*|NIGHTTROOP_IMAGE=nighttroop-rollback:<time>|' .env
docker compose up -d nighttroop
```

## Maintenance

- **Logs:** `docker compose logs -f nighttroop` or `docker compose logs -f caddy`. Up to 50 MB for each container.
- **Certificates:** Caddy renews them well before expiry, in the `caddy_data` volume. No cron or certbot.
- **Caddy:** rebuild on the newest Caddy 2.11 release and restart:

  ```sh
  docker compose build --pull caddy && docker compose up -d caddy
  ```

- **VPS:** run `apt update && apt upgrade -y` from time to time, and reboot when it asks. Both containers restart
  automatically.
- **Snapshots:** before a big change, in hPanel: **VPS**, **Manage**, **Backups & Monitoring**, **Snapshots &
  Backups**, **Create Snapshot**. A whole-machine fallback, in addition to the script's backups. Hostinger keeps one
  snapshot at a time, and its own automatic backups weekly.
- **Disk space:** `du -sh /opt/nighttroop/*` (data, backups and failed-update folders) and `docker system df`
  (images).

## Troubleshooting

### Caddy does not get certificates

Find the error in `docker compose logs caddy`:

- **`401` or `unauthorized` from Hostinger:** the token is wrong or expired. Put a new token in `.env`, then run
  `docker compose up -d caddy`.
- **Timeouts for the TXT record:** `dig +short NS example.com` must show `dns-parking.com` names (Hostinger's
  nameservers). Caddy checks with public resolvers (`1.1.1.1` and `8.8.8.8`), which see the record once Hostinger has
  it.
- **Too many failed attempts:** Let's Encrypt limits failures for each domain. While you repair the problem, use its
  staging service (looser limits, untrusted certificates):
  1. In `hostinger.compose.yml`, add `acme_ca https://acme-staging-v02.api.letsencrypt.org/directory` in the `tls`
     block of the Caddyfile.
  2. Run `docker compose up -d caddy`.
  3. When Caddy has the certificates, remove the line.

### The browser shows a 404 page

Nighttroop answers only host names under `DOMAIN`.

- Make sure that `DOMAIN` in `.env` is exactly your domain, then run `docker compose up -d`.
- A subdomain that is not a site (for example `www.example.com`) gets a 404 by design.

### Nighttroop is not healthy

Find the cause in `docker compose logs nighttroop`:

- **`cannot create server config`:** a setting has a bad value, for example `TRUSTED_PROXIES`.
- **A setting has no effect:** check its spelling in `nighttroop.env`. Nighttroop ignores misspelled environment
  variables and uses the default.
- **Permission denied under `/data`:** the data folder must belong to user `10001`:

  ```sh
  chown -R 10001:10001 /opt/nighttroop/data
  ```

- **Still starting after a while:** a first start on a new VPS restores from S3. More data takes longer. Wait a few
  minutes.
- **`Site DB is missing, so the site is not opened`:** a live site's database is missing, and its batch's S3 backup
  did not restore it. The server does not start (`live sites' DBs are missing and were not restored:` names them). Put
  the file back (for example, from `./backups`), then restart.

### Everyone seems to share one IP address

If rate limits occur much too often, check in `hostinger.compose.yml` that:

- `TRUSTED_PROXIES` is Caddy's address, `172.30.0.2`.
- The subnet of the `internal` network is unchanged.

**Warning:** Do not widen `TRUSTED_PROXIES` to a range that anyone can send from. A trusted address can claim any IP,
and get past the rate limits.

### The `172.30.0.0/24` network is already used on the VPS

In `hostinger.compose.yml`:

1. Change `subnet` and `ip_range` under `networks`.
2. Set Caddy's `ipv4_address` and `TRUSTED_PROXIES` to the same new address, inside the subnet but outside `ip_range`.
