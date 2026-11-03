# Deploying Nighttroop to a Hostinger VPS

This guide runs Nighttroop on a [Hostinger](https://www.hostinger.com/vps-hosting) VPS with Docker. It is written for
someone new to Hostinger who is comfortable with SSH and a Linux shell.

When you are done you will have:

- **Caddy** answering on ports 80 and 443. It gets Let's Encrypt certificates and renews them by itself.
  - The certificates are wildcards, proven through Hostinger's DNS API, so every site's subdomain is covered.
- **Nighttroop** behind Caddy, on a private Docker network with no ports of its own.
- **A data folder** on the VPS with the SQLite databases, optionally backed up to S3.
- **An update script** that backs up, updates, checks health and rolls back if the new version is not healthy.

The files this guide uses are in this folder:

| File | What it is |
| --- | --- |
| [`hostinger.compose.yml`](hostinger.compose.yml) | Docker Compose file for Nighttroop and Caddy |
| [`update.sh`](update.sh) | The safe update script |

## Contents

1. [How the pieces fit](#how-the-pieces-fit)
2. [What you need](#what-you-need)
3. [Set up the VPS](#1-set-up-the-vps)
4. [Open the firewall](#2-open-the-firewall)
5. [Point your domain at the VPS](#3-point-your-domain-at-the-vps)
6. [Make a Hostinger API token](#4-make-a-hostinger-api-token)
7. [Choose where backups go](#5-choose-where-backups-go)
8. [Put the files on the VPS](#6-put-the-files-on-the-vps)
9. [Configure it](#7-configure-it)
10. [Start it](#8-start-it)
11. [Set up the control plane](#9-set-up-the-control-plane)
12. [Updating Nighttroop safely](#updating-nighttroop-safely)
13. [Other upkeep](#other-upkeep)
14. [Troubleshooting](#troubleshooting)

## How the pieces fit

Nighttroop serves several kinds of host names, all from one domain (`example.com` below):

| Host | What it serves |
| --- | --- |
| `example.com` | The public page, where people create sites |
| `control.example.com` | The control plane, for the server's admin |
| `<site>.example.com` | Each site |
| `<site>.ns.example.com` | Each site's Nightscout-compatible API, for apps like AAPS and Loop |

So the server needs:

- **DNS:** records for `example.com`, `*.example.com` and `*.ns.example.com`, all pointing at the VPS.
- **Certificates:** one for `example.com` and wildcards for `*.example.com` and `*.ns.example.com`.
  - Let's Encrypt only gives wildcards to someone who proves they control the domain's DNS.
  - Caddy does that with a DNS-01 challenge: it adds a TXT record through Hostinger's API, which is why it needs an
    API token.
  - Caddy is built with the community [`caddy-dns-hostinger`](https://github.com/sbrunk/caddy-dns-hostinger) module for
    this.

Requests then flow like this:

1. A browser or app connects to Caddy on port 443.
2. Caddy proxies the request to Nighttroop on port 3000, over the private Docker network.
3. Nighttroop believes the client IP Caddy sends in `X-Forwarded-For`, because Caddy's address (`172.30.0.2`) is its
   only trusted proxy (`TRUSTED_PROXIES`). This keeps rate limits per visitor rather than per proxy.

## What you need

- **A Hostinger VPS.** Web and shared hosting plans can't run Nighttroop: it is a long-running server keeping its
  databases on local disk.
  - The smallest KVM plan is enough to start. Nighttroop is built to be cheap per user.
- **A domain using Hostinger's nameservers**, so Caddy can make DNS records through the Hostinger API.
  - A domain bought from Hostinger uses them already.
  - A domain from another registrar needs its nameservers changed to Hostinger's (see
    [step 3](#3-point-your-domain-at-the-vps)).
- **A GitHub account with access to this repository.** The image is published to its container registry, which is
  private while the repository is.
- **Somewhere for backups (optional but strongly recommended):** S3-compatible storage. Hostinger doesn't sell any;
  see [step 5](#5-choose-where-backups-go).
- **On your own computer:** a clone of this repository, and `ssh` and `scp`.

## 1. Set up the VPS

1. **Buy the VPS.** In [hPanel](https://hpanel.hostinger.com), buy a KVM VPS plan.
2. **Pick the operating system.** During setup choose the **Docker** template, which is Ubuntu with Docker and Docker
   Compose installed.
   - To change it later: open the VPS, then **OS & Panel**, **Operating System**, and search for Docker under
     **Change OS**. Changing the OS wipes the VPS.
   - Plain **Ubuntu 24.04** also works; install Docker yourself with
     [Docker's instructions for Ubuntu](https://docs.docker.com/engine/install/ubuntu/).
3. **Set how you log in.** Set a strong root password, and add your SSH public key so you can log in without it.
4. **Note the VPS's IP address.** It is shown on the VPS's **Overview** page in hPanel: open **VPS**, then **Manage** on
   your server. Below it is written `203.0.113.10`.
5. **Log in and update the system:**

   ```sh
   ssh root@203.0.113.10
   apt update && apt upgrade -y
   ```

6. **Check Docker Compose is new enough.** The compose file needs Docker Compose 2.23.1 or later, for inline configs:

   ```sh
   docker compose version
   ```

This guide manages Nighttroop over SSH with `docker compose`, not through hPanel's Docker Manager, because the
update script runs on the server.

## 2. Open the firewall

Hostinger's VPS firewall is managed in hPanel, outside the VPS. Once a firewall is turned on it drops everything that no
rule accepts, so add every rule before turning it on.

1. In hPanel, open **VPS**, then **Manage** on your server.
2. Open **Security**, then **Firewall**, and create a firewall configuration.
3. Add these **accept** rules, each from any source:

   | Protocol | Port | Why |
   | --- | --- | --- |
   | TCP | 22 | SSH. Add it first, so you can't lock yourself out |
   | TCP | 80 | HTTP, which Caddy redirects to HTTPS |
   | TCP | 443 | HTTPS |
   | UDP | 443 | HTTP/3 |

4. Activate the configuration for your VPS.

Don't open port 3000: Nighttroop is only reachable through Caddy.

## 3. Point your domain at the VPS

### If the domain is registered somewhere else

1. Add the domain to your Hostinger account, so its DNS zone shows in hPanel.
2. At that registrar, change its nameservers to the ones hPanel shows for the domain.
   - They are usually `ns1.dns-parking.com` and `ns2.dns-parking.com`, but some accounts get other `dns-parking.com`
     names, so use what hPanel shows.
3. Wait for the change to spread. It can take up to a day or two, though it is often much faster.

### Add the records

1. In hPanel, open **Domains**, then your domain, then **DNS / Nameservers**.
2. Remove any existing `A`, `AAAA` or `CNAME` records named `@`, `*` or `ns`. New domains get parking records that
   would clash.
   - Leave `MX` and `TXT` records alone if you use email on the domain.
3. Add three `A` records, each pointing to the VPS's IP:

   | Type | Name | Points to | TTL |
   | --- | --- | --- | --- |
   | A | `@` | `203.0.113.10` | 3600 |
   | A | `*` | `203.0.113.10` | 3600 |
   | A | `*.ns` | `203.0.113.10` | 3600 |

4. If your VPS has an IPv6 address, add the same three as `AAAA` records with it.
   - Its IPv6 address is on the same **Overview** page as the IPv4 one.
5. Check the records from your own computer. All three should print the VPS's IP:

   ```sh
   dig +short example.com
   dig +short anything.example.com
   dig +short anything.ns.example.com
   ```

You don't need a `TXT` record for the certificates: Caddy adds and removes those itself.

## 4. Make a Hostinger API token

Caddy uses this token to add the `_acme-challenge` TXT records Let's Encrypt checks.

1. In hPanel, open your profile, then **API**, or go straight to <https://hpanel.hostinger.com/profile/api>.
2. Make a new token, named e.g. `nighttroop-caddy`.
3. **Pick a long expiry.**
   - When the token expires, Caddy can no longer renew the certificates, and the site goes down when they run out.
   - Put a reminder in your calendar to replace it before then.
4. Copy the token now; it is only shown once.

Keep the token secret. It can manage your Hostinger account through the API, not just this domain's DNS.

## 5. Choose where backups go

Nighttroop can back up its databases to S3-compatible storage. It also restores from there when it starts on an empty
data folder, for example on a new VPS.

- **Without S3:** Nighttroop still runs, but your only copies are on the VPS.
  - The update script's local backups, and Hostinger's VPS backups and snapshots, are still there.
- **With S3:** Hostinger doesn't sell object storage. Good options with free or cheap tiers are:
  - [Cloudflare R2](https://developers.cloudflare.com/r2/)
  - [Backblaze B2](https://www.backblaze.com/cloud-storage)

To set up S3:

1. Make a private bucket, e.g. `nighttroop-backups`.
2. Make an access key that can only read and write that bucket.
3. Note four values for [step 7](#7-configure-it): the S3 endpoint URL, the bucket name, the access key ID and the
   secret key.

The backups are encrypted with your `KEY_SETS` before they are uploaded.

## 6. Put the files on the VPS

1. **Make Nighttroop's folder.** On the VPS, make the folder it will live in, with a `data` folder inside. The image runs
   Nighttroop as user ID `10001`, so `data` must belong to that user:

   ```sh
   mkdir -p /opt/nighttroop/data
   chown 10001:10001 /opt/nighttroop/data
   ```

2. **Copy the files up.** From your clone of this repository, on your own computer:

   ```sh
   scp docs/deploy/hostinger.compose.yml docs/deploy/update.sh root@203.0.113.10:/opt/nighttroop/
   ```

3. **Let the VPS pull the image.** The image is at `ghcr.io/grahamjenson/nighttroop`. While the repository is private,
   the VPS has to log in to GitHub's container registry:
   1. On GitHub, make a **classic** personal access token with only the `read:packages` scope:
      **Settings**, then **Developer settings**, **Personal access tokens**, **Tokens (classic)**.
   2. On the VPS, log in with your GitHub username, pasting the token as the password:

      ```sh
      docker login ghcr.io -u <your-github-username>
      ```

## 7. Configure it

There are two files of settings in `/opt/nighttroop`. Make them readable by root only, as they hold secrets:

```sh
cd /opt/nighttroop
touch .env nighttroop.env
chmod 600 .env nighttroop.env
```

### `.env`: settings for Docker Compose and Caddy

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

### `nighttroop.env`: settings for Nighttroop

The compose file already sets the host name, HTTPS and the trusted proxy, so this file only holds secrets and options.

1. Make the encryption keys. The image includes the tool:

   ```sh
   docker run --rm --entrypoint create_keysets ghcr.io/grahamjenson/nighttroop:latest
   ```

2. Put the keys and your other settings in `nighttroop.env`:

   ```sh
   # The keys from create_keysets: one long line
   KEY_SETS=...

   # S3 backups (step 5); leave these out to run without them
   S3_URL=https://<account-id>.r2.cloudflarestorage.com
   S3_BUCKET=nighttroop-backups
   S3_ACCESS_KEY=...
   S3_SECRET_KEY=...

   # Optional: report errors to Sentry
   # SENTRY_DSN=...
   ```

3. **Keep a copy of `KEY_SETS` somewhere safe, off the VPS**, such as a password manager.
   - It encrypts the secrets in the databases and the backups.
   - Without it, the data can't be read, and a backup can't be restored onto a new VPS.

## 8. Start it

```sh
cd /opt/nighttroop
docker compose up -d --build
```

The first start takes a few minutes:

- **Building Caddy:** the `--build` compiles Caddy with the Hostinger DNS module.
- **Pulling Nighttroop:** the image is downloaded.
- **Getting certificates:** Caddy asks for them, which can take a minute or two while the DNS challenge record spreads.

Watch Caddy get the certificates:

```sh
docker compose logs -f caddy
```

You should see `certificate obtained successfully` for `example.com`, `*.example.com` and `*.ns.example.com`. Press
Ctrl+C to stop watching.

Then check that everything works:

- [ ] `docker compose ps` shows both containers running, and `nighttroop` as `healthy`.
- [ ] `curl https://example.com/health` prints `{"status":"ok"}`.
- [ ] <https://example.com> goes to the setup page at `https://control.example.com/setup`, with a valid padlock.
- [ ] `http://example.com` redirects to HTTPS.

## 9. Set up the control plane

A new server has no control plane admin, so every page goes to the setup page until you make one. Only someone with
the setup code the server prints to its logs can, so nobody else can take a new server before you do.

1. Find the setup code in the logs:

   ```sh
   docker compose logs nighttroop | grep -A1 "SET UP NIGHTTROOP"
   ```

   - The code is printed on every start until setup is done, so restart Nighttroop to see it again.
2. Open <https://control.example.com/setup> and enter the code.
3. It shows the admin's username, password and TOTP secret, once:
   1. Store the password and TOTP secret somewhere safe.
   2. Add the TOTP secret to your authenticator app, e.g. by scanning its QR code.
   3. Enter the code from your authenticator app and press **Finish setup**.
4. Log in at <https://control.example.com>. The rest of the server, such as creating sites, now works.

## Updating Nighttroop safely

### Where new versions come from

Merges to `main` build no image. A version is released by pushing a tag for it, on a commit CI passed on:

```sh
git tag v1.2.3 && git push origin v1.2.3
```

The **Docker image** workflow (`.github/workflows/docker.yml`) then builds the server's image from that tag and pushes
two tags to `ghcr.io/grahamjenson/nighttroop`:

- **`:v1.2.3`** is that version, and never changes.
- **`:latest`** is the newest version.

Deploy the version tags, so you always know which version is running and can go back to it. You can find them:

- on the repository's **Packages** page on GitHub;
- in the repository's tags.

### Running an update

On the VPS:

```sh
cd /opt/nighttroop
sudo ./update.sh ghcr.io/grahamjenson/nighttroop:v1.2.3
```

The script:

1. **Pulls the new image** while the old version keeps running, so a bad tag or a registry problem changes nothing.
   - If the new image is the one already running, it stops here.
2. **Keeps the running image** under its own name, `nighttroop-rollback:<time>`, so it can always start it again.
3. **Stops Nighttroop**, which uploads a last backup to S3 on the way down (it allows two minutes).
4. **Copies `./data`** to `./backups/data-<time>.tar.gz` while nothing is writing to it.
5. **Starts the new image** and waits up to 3 minutes for its health check to pass.
   - The health check asks the server's `/health`, which answers once the control database does.
6. **If it is healthy, it is done.** It keeps the newest 5 backups and previous images and deletes older ones.
7. **If it isn't healthy, it rolls back:**
   1. prints the new version's last logs;
   2. stops it, and moves its `./data` aside to `./data-failed-<time>`;
   3. puts back `./data` from the copy made in step 4;
   4. starts the previous image again, and checks that it is healthy.

Nighttroop is down from step 3 until it is healthy in step 5, usually well under a minute. Caddy keeps running
throughout.

How the script finishes:

| Exit code | Meaning |
| --- | --- |
| `0` | Updated, or already running that image |
| `1` | The new version wasn't healthy and the previous one is back, with its data |
| `2` | Neither version is healthy; look at `docker compose logs nighttroop` |

You can change its defaults with environment variables:

- `sudo HEALTH_TIMEOUT=300 ./update.sh ...` waits 5 minutes for the health check instead of 3, e.g. for a large
  restore.
- `sudo KEEP_BACKUPS=10 ./update.sh ...` keeps 10 backups and previous images instead of 5.

### What a rollback can't undo

- **Data the new version wrote** before it was rolled back is not in the restored `./data`. It is kept in
  `./data-failed-<time>` if you need it.
- **S3 backups the new version made.** It may have uploaded backups in its new format while it ran.
  - The restored version runs from its local data, and replaces those uploads with its own on its next backups.
  - Only a fresh VPS would restore from S3, so don't move servers until the old version has backed up again.

Delete `./data-failed-*` folders once you are sure you don't need them; the script doesn't.

### Rolling back by hand

To go back to an earlier version, use the copy of `./data` made just before that version was replaced. Its databases
are in the format that version expects.

```sh
cd /opt/nighttroop
docker compose stop nighttroop
mv data data-before-rollback
tar -xzf backups/data-<time>.tar.gz          # makes ./data again
docker images nighttroop-rollback            # the earlier images, by time
sed -i 's|^NIGHTTROOP_IMAGE=.*|NIGHTTROOP_IMAGE=nighttroop-rollback:<time>|' .env
docker compose up -d nighttroop
```

## Other upkeep

- **Logs:**
  - `docker compose logs -f nighttroop` or `docker compose logs -f caddy`.
  - Each container keeps up to 50 MB of logs.
- **Certificates:** Caddy renews them by itself, well before they expire, and keeps them in the `caddy_data` volume.
  You don't need cron or certbot.
- **Updating Caddy:** rebuild it on the newest Caddy 2.11 release, then restart it:

  ```sh
  docker compose build --pull caddy && docker compose up -d caddy
  ```

- **Updating the VPS:** run `apt update && apt upgrade -y` now and then, and reboot when it asks. Both containers start
  again by themselves.
- **Snapshots:**
  - Before a big change, take a snapshot in hPanel: **VPS**, then **Manage**, **Backups & Monitoring**, **Snapshots &
    Backups**, **Create Snapshot**.
  - It is a whole-machine fallback alongside the script's backups. Hostinger keeps one snapshot at a time, and its own
    automatic backups weekly.
- **Disk space:**
  - `du -sh /opt/nighttroop/*` shows what the data, backups and failed-update folders use.
  - `docker system df` shows what images use.

## Troubleshooting

### Caddy doesn't get certificates

Look at `docker compose logs caddy` for the error. Common causes:

- **`401` or `unauthorized` from Hostinger:** the API token is wrong or has expired. Make a new one, put it in `.env`,
  and run `docker compose up -d caddy`.
- **Timeouts waiting for the TXT record:**
  - Check the domain really uses Hostinger's nameservers: `dig +short NS example.com` should show `dns-parking.com`
    names.
  - Caddy checks the record with public resolvers (`1.1.1.1` and `8.8.8.8`), which see it once Hostinger's nameservers
    have it.
- **Too many failed attempts:** Let's Encrypt limits how often a domain can fail. While you fix things, try Let's
  Encrypt's staging service, which has looser limits but gives untrusted certificates:
  1. Add `acme_ca https://acme-staging-v02.api.letsencrypt.org/directory` inside the `tls` block of the Caddyfile in
     `hostinger.compose.yml`.
  2. Run `docker compose up -d caddy`.
  3. Remove the line again once certificates are obtained.

### The browser shows a 404 page

Nighttroop only answers host names under `DOMAIN`.

- Check `DOMAIN` in `.env` is exactly your domain, then run `docker compose up -d`.
- A subdomain that isn't a site, such as `www.example.com`, gets a 404 by design.

### Nighttroop isn't healthy

Look at `docker compose logs nighttroop`:

- **`cannot create server config`:** a setting has a bad value, such as `TRUSTED_PROXIES`.
- **A setting seems to be ignored:** check its spelling in `nighttroop.env`. Settings there reach Nighttroop as
  environment variables, so a misspelled one is ignored and the default is used.
- **Permission denied under `/data`:** the data folder isn't owned by user `10001`:

  ```sh
  chown -R 10001:10001 /opt/nighttroop/data
  ```

- **Still starting after a while:** a first start on a new VPS restores from S3, which takes longer the more data there
  is. Give it a few minutes.

### Everyone seems to share one IP address

If rate limits are hit far too often, check `TRUSTED_PROXIES` in `hostinger.compose.yml` is Caddy's address,
`172.30.0.2`, and that the `internal` network's subnet is unchanged.

- Don't widen `TRUSTED_PROXIES` to a whole range anyone can send from. Anything it trusts can claim to be any IP, and
  so get past the rate limits.

### The `172.30.0.0/24` network is already used on the VPS

Change both the `subnet` and `ip_range` under `networks` in `hostinger.compose.yml`, and set Caddy's `ipv4_address`
and `TRUSTED_PROXIES` to the same new address. It must be inside the subnet but outside `ip_range`.
