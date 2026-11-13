# Deploy Nighttroop to a Voyager VPS

This guide installs Nighttroop with Docker on a [Voyager](https://voyager.nz/business/hosting/virtual-servers) VPS,
in Voyager's Auckland data centre. It is for operators who know SSH and a Linux shell.

## Overview

### Result

- [`voyager.compose.yml`](voyager.compose.yml) runs:
  - **Caddy** on ports 80 and 443. It gets and renews Let's Encrypt certificates automatically.
  - **Nighttroop** behind Caddy, on a private Docker network, with no ports of its own.
- **A data folder** holds the SQLite databases, optionally backed up to S3.
- [`update.sh`](../update.sh), the safe update script shared with [the Hostinger guide](../hostinger/hostinger.md),
  backs up, updates, checks health, and rolls back if the new version is not healthy.

### Host names

All host names come from one domain (`example.nz` in this guide):

- `example.nz`: the public page, where people create sites.
- `control.example.nz`: the control plane, for the control admin.
- `<site>.example.nz`: each site.
- `<site>.ns.example.nz`: the Nightscout-compatible API of each site, for apps like AAPS and Loop.

### DNS and certificates

- **DNS:** `example.nz`, `*.example.nz` and `*.ns.example.nz` point to the VPS.
- **Certificates:** `example.nz`, and wildcards `*.example.nz` and `*.ns.example.nz` for the subdomain of each site.
- **Wildcard proof:** Let's Encrypt gives wildcards only with proof of DNS control.
  - Caddy adds a TXT record through the API of a DNS provider (a DNS-01 challenge).
  - Voyager's DNS (through 1st Domains) has no API that Caddy can use. Thus the domain's DNS moves to
    [Cloudflare](https://www.cloudflare.com/plans/), which has one on its free plan.
  - The domain can stay registered where it is.
  - Caddy is built with the [`caddy-dns/cloudflare`](https://github.com/caddy-dns/cloudflare) module.

### Request flow

1. A browser or app connects to Caddy on port 443, directly on the VPS. Cloudflare only answers DNS, and does not
   proxy the traffic.
2. Caddy proxies the request to Nighttroop on port 3000, over the private Docker network.
3. Nighttroop uses the client IP in `X-Forwarded-For`, from its only trusted proxy (`TRUSTED_PROXIES`), Caddy
   (`172.30.0.2`). Thus rate limits apply to each visitor.

## Requirements

- **A Voyager VPS.** The smallest Linux VPS is sufficient to start. Nighttroop is built to be cheap per user.
  - Voyager's shared web hosting cannot run Nighttroop, a long-running server with its databases on local disk.
  - Voyager's Cloud Servers also work: do the same steps on an Ubuntu server there.
- **A domain whose nameservers you can change,** registered with Voyager's 1st Domains or anyone else.
- **A free Cloudflare account** for the domain's DNS.
- **GitHub access to this repository.** Its container registry has the image, and is private while the repository is.
- **S3-compatible storage** (optional, strongly recommended). See [step 5](#5-choose-where-backups-go).
- **On your computer:** a clone of this repository, `ssh` and `scp`.

## Prepare the host

### 1. Set up the VPS

1. In Voyager's VPS Manager, deploy a Linux VPS in Auckland from the **Ubuntu 24.04 LTS** template.
   - Add your SSH public key if it asks for one. If not, note the root password that it gives.
2. Note the VPS IP, and its IPv6 address if it has one. This guide uses `203.0.113.10`.
3. Log in and update the system:

   ```sh
   ssh root@203.0.113.10
   apt update && apt upgrade -y
   ```

4. When your key works, stop SSH from accepting passwords, so that nobody can guess them:

   ```sh
   echo "PasswordAuthentication no" > /etc/ssh/sshd_config.d/10-no-passwords.conf
   systemctl reload ssh
   ```

   - **Caution:** Before you close the first terminal, check that you can log in from a second terminal.
5. Install Docker and Docker Compose from Docker's `apt` repository, with
   [Docker's instructions for Ubuntu](https://docs.docker.com/engine/install/ubuntu/).
6. Check that Docker Compose is 2.23.1 or later, for inline configs:

   ```sh
   docker compose version
   ```

### 2. Turn on the firewall

1. Turn on `ufw`, the Ubuntu firewall, on the VPS:

   ```sh
   ufw allow 22/tcp       # SSH: add it first, so you can't lock yourself out
   ufw allow 80/tcp       # HTTP, which Caddy redirects to HTTPS
   ufw allow 443/tcp      # HTTPS
   ufw allow 443/udp      # HTTP/3
   ufw enable
   ufw status
   ```

2. If Voyager's VPS Manager also has a firewall for your server, open the same ports there.

**Warning:** Do not open port 3000. Nighttroop is reachable only through Caddy.

`ufw` does not see the rules that Docker adds for published ports. Only Caddy's 80 and 443 are published, and they
are open already.

## Prepare the domain and storage

### 3. Move the domain's DNS to Cloudflare

#### Add the domain to Cloudflare

1. In the [Cloudflare dashboard](https://dash.cloudflare.com), **Add a domain**: `example.nz`, on the **Free** plan.
2. Cloudflare copies the records that it finds. Remove `A`, `AAAA` and `CNAME` records named `example.nz`, `*` or
   `*.ns`. Keep `MX` and `TXT` records if you use email on the domain.
3. Note the two Cloudflare nameservers, for example `ada.ns.cloudflare.com` and `bob.ns.cloudflare.com`.

#### Change the nameservers

1. At the registrar, change the nameservers of the domain to the two Cloudflare nameservers.
   - With 1st Domains: open the domain in the **Domain Manager** and change its name servers.
   - **Caution:** This delegation removes the 1st Domains DNS zone. Make the records that you need again in
     Cloudflare, for example email.
2. Wait until Cloudflare shows the domain as **Active**: up to one or two days, often much less.

#### Add the records

1. In Cloudflare, open the domain, **DNS**, **Records**.
2. Add `A` records `@`, `*` and `*.ns` to `203.0.113.10`, with **Proxy status** set to **DNS only** (a grey cloud).
   - **Warning:** Use DNS only. The Cloudflare proxy ends HTTPS itself, cannot cover `*.ns.example.nz` on the free
     plan, and hides the IP of each visitor from the Nighttroop rate limits.
3. If the VPS has an IPv6 address, add the same records as `AAAA` records, also DNS only.
4. On your computer, check that each command prints the VPS IP:

   ```sh
   dig +short example.nz
   dig +short anything.example.nz
   dig +short anything.ns.example.nz
   ```

You do not need `TXT` records for the certificates. Caddy adds and removes them.

### 4. Make a Cloudflare API token

Caddy uses it to add the `_acme-challenge` TXT records that Let's Encrypt checks.

1. In Cloudflare, open **My Profile**, **API Tokens**, **Create Token**.
2. Use the **Edit zone DNS** template.
3. Under **Zone Resources**, select **Include**, **Specific zone** and your domain. The token can then change only
   this domain's DNS.
4. Create the token, and copy it. Cloudflare shows it only once.

The token does not expire unless you give it an end date. If you do, put a reminder in your calendar to replace it.
Without a valid token, Caddy cannot renew the certificates.

### 5. Choose where backups go

Nighttroop can back up its databases to S3-compatible storage, encrypted with your `KEY_SETS`. It restores from there
when it starts on an empty data folder, for example on a new VPS.

- **Without S3:** your only copies are the update script's local backups on the VPS, and any VPS backups that Voyager
  keeps.
- **With S3:**
  - [Cloudflare R2](https://developers.cloudflare.com/r2/), in the same account as the DNS. Give the bucket the
    **Oceania** location hint to keep the backups near New Zealand.
  - [Backblaze B2](https://www.backblaze.com/cloud-storage).
  - Voyager object storage, if your plan includes S3-compatible storage. Ask Voyager.

**Note:** The backups contain people's health data. If they must stay in New Zealand, choose storage there.

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
   scp docs/deploy/voyager/voyager.compose.yml docs/deploy/update.sh root@203.0.113.10:/opt/nighttroop/
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
# So plain `docker compose` commands in this folder, and update.sh, use the right file
COMPOSE_FILE=voyager.compose.yml

# Your domain, without https:// or a trailing dot
DOMAIN=example.nz

# Let's Encrypt emails this address about problems with your certificates
ACME_EMAIL=you@example.nz

# The token from step 4
CLOUDFLARE_API_TOKEN=paste-the-token-here

# The Nighttroop image to run; update.sh changes this line
NIGHTTROOP_IMAGE=ghcr.io/grahamjenson/nighttroop:latest
```

#### `nighttroop.env`: settings for Nighttroop

The compose file sets the host name, HTTPS and the trusted proxy. This file holds only secrets and options.
[`docs/config.md`](../../config.md) lists all settings.

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

1. Start the containers. The first start takes a few minutes: it compiles Caddy with the Cloudflare DNS module
   (`--build`), pulls Nighttroop and gets the certificates.

   ```sh
   cd /opt/nighttroop
   docker compose up -d --build
   ```

2. Watch the Caddy logs for `certificate obtained successfully` for `example.nz`, `*.example.nz` and
   `*.ns.example.nz`. This can take a minute while the DNS challenge record spreads. Then push Ctrl+C.

   ```sh
   docker compose logs -f caddy
   ```

3. Check that:
   - [ ] `docker compose ps` shows both containers running, and `nighttroop` as `healthy`.
   - [ ] `curl https://example.nz/health` prints `{"status":"ok"}`.
   - [ ] <https://example.nz> goes to the setup page at `https://control.example.nz/setup`, with a valid padlock.
   - [ ] `http://example.nz` redirects to HTTPS.

### 9. Set up the control plane

A new server has no control admin, so all pages go to the setup page. Setup needs the code from the server logs, so
nobody else can take the server first.

1. Find the setup code. The server logs it at each start until setup is complete, so restart Nighttroop to see it
   again.

   ```sh
   docker compose logs nighttroop | grep -A1 "SET UP NIGHTTROOP"
   ```

2. Open <https://control.example.nz/setup> and enter the code.
3. Store the control admin's password and TOTP secret safely. The page shows them, with the username, only once.
4. Add the TOTP secret to your authenticator app (for example, scan its QR code).
5. Enter the code from the app, then push **Finish setup**.
6. Log in at <https://control.example.nz>. The rest of the server, such as site creation, now works.

## Updating Nighttroop safely

The update is the same as on Hostinger, with the same script. Deploy a version tag, so that you know which version
runs ([where new versions come from](../hostinger/hostinger.md#where-new-versions-come-from)):

```sh
cd /opt/nighttroop
sudo ./update.sh ghcr.io/grahamjenson/nighttroop:v1.2.3
```

The script uses the compose file that `.env` names, here `voyager.compose.yml`. It pulls the new image, keeps the
running one, and copies `./data` while Nighttroop is stopped. If the new version is not healthy within 3 minutes, it
rolls back to the previous image and data. This loses nothing: a version that is not healthy served nothing.

**Caution:** After a new version served, fix forward: update to a version with a fix. To go back, you must restore an
old copy of `./data`, and you lose everything written since.

[The Hostinger guide's update section](../hostinger/hostinger.md#updating-nighttroop-safely) gives each step, the
exit codes, what a rollback cannot undo, and how to roll back by hand. They are the same here.

## Maintenance

- **Logs:** `docker compose logs -f nighttroop` or `docker compose logs -f caddy`. Up to 50 MB for each container.
- **Certificates:** Caddy renews them well before expiry, in the `caddy_data` volume. No cron or certbot.
- **Caddy:** rebuild on the newest Caddy 2.11 release and restart:

  ```sh
  docker compose build --pull caddy && docker compose up -d caddy
  ```

- **VPS:** run `apt update && apt upgrade -y` from time to time, and reboot when it asks. Both containers restart
  automatically.
- **Whole-VPS backups:** if your plan keeps VPS backups or snapshots (see Voyager's VPS Manager, or ask Voyager), take
  one before a big change. A whole-machine fallback, in addition to the script's backups and S3.
- **Disk space:** `du -sh /opt/nighttroop/*` (data, backups and failed-update folders) and `docker system df`
  (images).

## Troubleshooting

### Caddy does not get certificates

Find the error in `docker compose logs caddy`:

- **`Invalid request headers`, `Authentication error` or `403` from Cloudflare:** the token is wrong, expired, or
  cannot edit this domain's DNS. Put a new token ([step 4](#4-make-a-cloudflare-api-token)) in `.env`, then run
  `docker compose up -d caddy`.
- **Timeouts for the TXT record:** the domain must use Cloudflare's nameservers. `dig +short NS example.nz` must show
  the two `ns.cloudflare.com` names, and Cloudflare must show the domain as **Active**.
- **Too many failed attempts:** Let's Encrypt limits failures for each domain. While you repair the problem, use its
  staging service (looser limits, untrusted certificates):
  1. In `voyager.compose.yml`, add `acme_ca https://acme-staging-v02.api.letsencrypt.org/directory` in the `tls`
     block of the Caddyfile.
  2. Run `docker compose up -d caddy`.
  3. When Caddy has the certificates, remove the line.

### Pages show Cloudflare errors, or everyone shares a few IP addresses

Cloudflare proxies the records, but Caddy must answer visitors itself. On the Cloudflare **DNS** page, set the three
records to **DNS only** (a grey cloud).

### The browser shows a 404 page

Nighttroop answers only host names under `DOMAIN`.

- Make sure that `DOMAIN` in `.env` is exactly your domain, then run `docker compose up -d`.
- A subdomain that is not a site (for example `www.example.nz`) gets a 404 by design.

### The site cannot be reached

- `ufw status` must show 80 and 443 allowed.
- A firewall for the server in Voyager's VPS Manager must have the same ports open.
- `docker compose ps` must show Caddy with `0.0.0.0:80` and `0.0.0.0:443`.

### Nighttroop is not healthy

Find the cause in `docker compose logs nighttroop`:

- **`cannot create server config`:** a setting has a bad value, for example `TRUSTED_PROXIES`.
- **A setting has no effect:** compare its spelling in `nighttroop.env` with [`docs/config.md`](../../config.md).
  Nighttroop ignores misspelled environment variables and uses the default.
- **Permission denied under `/data`:** the data folder must belong to user `10001`:

  ```sh
  chown -R 10001:10001 /opt/nighttroop/data
  ```

- **Still starting after a while:** a first start on a new VPS restores from S3. More data takes longer. Wait a few
  minutes.
- **`Site DB is missing, so the site is not opened`:** a live site's database is missing, and its batch's S3 backup
  did not restore it. The server does not start (`live sites' DBs are missing and were not restored:` names them). Put
  the file back (for example, from `./backups`), then restart.

### The `172.30.0.0/24` network is already used on the VPS

In `voyager.compose.yml`:

1. Change `subnet` and `ip_range` under `networks`.
2. Set Caddy's `ipv4_address` and `TRUSTED_PROXIES` to the same new address, inside the subnet but outside `ip_range`.
