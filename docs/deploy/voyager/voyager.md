# Deploying Nighttroop to a Voyager VPS

This guide runs Nighttroop on a [Voyager](https://voyager.nz/business/hosting/virtual-servers) VPS with Docker, in
Voyager's Auckland data centre. It is written for someone comfortable with SSH and a Linux shell.

When you are done you will have:

- **Caddy** answering on ports 80 and 443. It gets Let's Encrypt certificates and renews them by itself.
  - The certificates are wildcards, proven through Cloudflare's DNS API, so every site's subdomain is covered.
- **Nighttroop** behind Caddy, on a private Docker network with no ports of its own.
- **A data folder** on the VPS with the SQLite databases, optionally backed up to S3.
- **An update script** that backs up, updates, checks health and rolls back if the new version is not healthy.

The files this guide uses are in this folder:

| File | What it is |
| --- | --- |
| [`voyager.compose.yml`](voyager.compose.yml) | Docker Compose file for Nighttroop and Caddy |
| [`update.sh`](update.sh) | The safe update script, shared with [the Hostinger guide](hostinger.md) |

## Contents

1. [How the pieces fit](#how-the-pieces-fit)
2. [What you need](#what-you-need)
3. [Set up the VPS](#1-set-up-the-vps)
4. [Turn on the firewall](#2-turn-on-the-firewall)
5. [Move the domain's DNS to Cloudflare](#3-move-the-domains-dns-to-cloudflare)
6. [Make a Cloudflare API token](#4-make-a-cloudflare-api-token)
7. [Choose where backups go](#5-choose-where-backups-go)
8. [Put the files on the VPS](#6-put-the-files-on-the-vps)
9. [Configure it](#7-configure-it)
10. [Start it](#8-start-it)
11. [Set up the control plane](#9-set-up-the-control-plane)
12. [Updating Nighttroop safely](#updating-nighttroop-safely)
13. [Other upkeep](#other-upkeep)
14. [Troubleshooting](#troubleshooting)

## How the pieces fit

Nighttroop serves several kinds of host names, all from one domain (`example.nz` below):

| Host | What it serves |
| --- | --- |
| `example.nz` | The public page, where people create sites |
| `control.example.nz` | The control plane, for the server's admin |
| `<site>.example.nz` | Each site |
| `<site>.ns.example.nz` | Each site's Nightscout-compatible API, for apps like AAPS and Loop |

So the server needs:

- **DNS:** records for `example.nz`, `*.example.nz` and `*.ns.example.nz`, all pointing at the VPS.
- **Certificates:** one for `example.nz` and wildcards for `*.example.nz` and `*.ns.example.nz`.
  - Let's Encrypt only gives wildcards to someone who proves they control the domain's DNS.
  - Caddy does that with a DNS-01 challenge: it adds a TXT record through a DNS provider's API.
  - Voyager's DNS (through 1st Domains) has no API Caddy can use, so the domain's DNS moves to
    [Cloudflare](https://www.cloudflare.com/plans/), whose free plan has one. The domain can stay registered where it
    is. Caddy is built with the [`caddy-dns/cloudflare`](https://github.com/caddy-dns/cloudflare) module for this.

Requests then flow like this:

1. A browser or app connects to Caddy on port 443, straight to the VPS. Cloudflare only answers DNS; it does not
   proxy the traffic.
2. Caddy proxies the request to Nighttroop on port 3000, over the private Docker network.
3. Nighttroop believes the client IP Caddy sends in `X-Forwarded-For`, because Caddy's address (`172.30.0.2`) is its
   only trusted proxy (`TRUSTED_PROXIES`). This keeps rate limits per visitor rather than per proxy.

## What you need

- **A Voyager VPS.** Voyager's shared web hosting can't run Nighttroop: it is a long-running server keeping its
  databases on local disk.
  - The smallest Linux VPS is enough to start. Nighttroop is built to be cheap per user.
  - Voyager's Cloud Servers also work: follow the same steps on an Ubuntu server there.
- **A domain**, registered with Voyager's 1st Domains or anyone else, whose nameservers you can change.
- **A free Cloudflare account**, for the domain's DNS.
- **A GitHub account with access to this repository.** The image is published to its container registry, which is
  private while the repository is.
- **Somewhere for backups (optional but strongly recommended):** S3-compatible storage; see
  [step 5](#5-choose-where-backups-go).
- **On your own computer:** a clone of this repository, and `ssh` and `scp`.

## 1. Set up the VPS

1. **Deploy the VPS.** In Voyager's VPS Manager, deploy a Linux VPS from the **Ubuntu 24.04 LTS** template, in
   Auckland.
   - Add your SSH public key if it asks for one; otherwise note the root password it gives you.
2. **Note the VPS's IP address**, and its IPv6 address if it has one. Below the IPv4 address is written `203.0.113.10`.
3. **Log in and update the system:**

   ```sh
   ssh root@203.0.113.10
   apt update && apt upgrade -y
   ```

4. **Log in with your key only.** Once your key works, stop SSH accepting passwords, so nobody can guess them:

   ```sh
   echo "PasswordAuthentication no" > /etc/ssh/sshd_config.d/10-no-passwords.conf
   systemctl reload ssh
   ```

   - Check you can still log in from a second terminal before closing the first.
5. **Install Docker** with [Docker's instructions for Ubuntu](https://docs.docker.com/engine/install/ubuntu/), using
   its `apt` repository. That installs Docker Compose too.
6. **Check Docker Compose is new enough.** The compose file needs Docker Compose 2.23.1 or later, for inline configs:

   ```sh
   docker compose version
   ```

## 2. Turn on the firewall

This guide uses `ufw`, Ubuntu's firewall, on the VPS itself. If Voyager's VPS Manager also has a firewall for your
server, open the same ports there.

```sh
ufw allow 22/tcp       # SSH: add it first, so you can't lock yourself out
ufw allow 80/tcp       # HTTP, which Caddy redirects to HTTPS
ufw allow 443/tcp      # HTTPS
ufw allow 443/udp      # HTTP/3
ufw enable
ufw status
```

- Don't open port 3000: Nighttroop is only reachable through Caddy.
- Docker adds its own rules for the ports a container publishes, which `ufw` doesn't see. The compose file only
  publishes Caddy's 80 and 443, which are open anyway.

## 3. Move the domain's DNS to Cloudflare

### Add the domain to Cloudflare

1. In the [Cloudflare dashboard](https://dash.cloudflare.com), **Add a domain**, enter `example.nz`, and pick the
   **Free** plan.
2. Cloudflare copies the records it can find. Remove any `A`, `AAAA` or `CNAME` records named `example.nz`, `*` or
   `*.ns`.
   - Keep `MX` and `TXT` records if you use email on the domain.
3. Note the two nameservers Cloudflare gives you, such as `ada.ns.cloudflare.com` and `bob.ns.cloudflare.com`.

### Change the nameservers

1. Where the domain is registered, change its nameservers to Cloudflare's two.
   - With 1st Domains, open the domain in the **Domain Manager** and change its name servers. Delegating the domain
     removes 1st Domains' own DNS zone, so recreate any records you need, such as email, in Cloudflare.
2. Wait for Cloudflare to say the domain is **Active**. It can take up to a day or two, though it is often much
   faster.

### Add the records

1. In Cloudflare, open the domain, then **DNS**, then **Records**.
2. Add three `A` records, each pointing to the VPS's IP, with **Proxy status** set to **DNS only** (a grey cloud):

   | Type | Name | IPv4 address | Proxy status |
   | --- | --- | --- | --- |
   | A | `@` | `203.0.113.10` | DNS only |
   | A | `*` | `203.0.113.10` | DNS only |
   | A | `*.ns` | `203.0.113.10` | DNS only |

   - **They must be DNS only.** Cloudflare's proxy would end HTTPS itself, can't cover `*.ns.example.nz` on the free
     plan, and would hide each visitor's IP from Nighttroop's rate limits.
3. If your VPS has an IPv6 address, add the same three as `AAAA` records with it, also DNS only.
4. Check the records from your own computer. All three should print the VPS's IP:

   ```sh
   dig +short example.nz
   dig +short anything.example.nz
   dig +short anything.ns.example.nz
   ```

You don't need a `TXT` record for the certificates: Caddy adds and removes those itself.

## 4. Make a Cloudflare API token

Caddy uses this token to add the `_acme-challenge` TXT records Let's Encrypt checks.

1. In Cloudflare, open **My Profile**, then **API Tokens**, then **Create Token**.
2. Use the **Edit zone DNS** template.
3. Under **Zone Resources**, choose **Include**, **Specific zone**, and your domain, so the token can only change this
   domain's DNS.
4. Create it, and copy the token now; it is only shown once.

The token doesn't expire unless you give it an end date. If you do, put a reminder in your calendar to replace it,
as Caddy can't renew the certificates without it.

## 5. Choose where backups go

Nighttroop can back up its databases to S3-compatible storage. It also restores from there when it starts on an empty
data folder, for example on a new VPS.

- **Without S3:** Nighttroop still runs, but your only copies are on the VPS, with the update script's local backups
  and any backups Voyager keeps of the VPS.
- **With S3:** good options are:
  - [Cloudflare R2](https://developers.cloudflare.com/r2/), in the same account as the DNS. Give the bucket the
    **Oceania** location hint to keep the backups near New Zealand.
  - [Backblaze B2](https://www.backblaze.com/cloud-storage).
  - Object storage from Voyager, if your plan includes S3-compatible storage; ask Voyager.

The backups are of people's health data. They are encrypted with your `KEY_SETS` before they are uploaded, but if they
have to stay in New Zealand, choose storage there.

To set up S3:

1. Make a private bucket, e.g. `nighttroop-backups`.
2. Make an access key that can only read and write that bucket.
3. Note four values for [step 7](#7-configure-it): the S3 endpoint URL, the bucket name, the access key ID and the
   secret key.

## 6. Put the files on the VPS

1. **Make Nighttroop's folder.** On the VPS, make the folder it will live in, with a `data` folder inside. The image runs
   Nighttroop as user ID `10001`, so `data` must belong to that user:

   ```sh
   mkdir -p /opt/nighttroop/data
   chown 10001:10001 /opt/nighttroop/data
   ```

2. **Copy the files up.** From your clone of this repository, on your own computer:

   ```sh
   scp docs/deploy/voyager.compose.yml docs/deploy/update.sh root@203.0.113.10:/opt/nighttroop/
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

### `nighttroop.env`: settings for Nighttroop

The compose file already sets the host name, HTTPS and the trusted proxy, so this file only holds secrets and options.
Every setting is listed in [`docs/config.md`](../config.md).

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

- **Building Caddy:** the `--build` compiles Caddy with the Cloudflare DNS module.
- **Pulling Nighttroop:** the image is downloaded.
- **Getting certificates:** Caddy asks for them, which can take a minute while the DNS challenge record spreads.

Watch Caddy get the certificates:

```sh
docker compose logs -f caddy
```

You should see `certificate obtained successfully` for `example.nz`, `*.example.nz` and `*.ns.example.nz`. Press
Ctrl+C to stop watching.

Then check that everything works:

- [ ] `docker compose ps` shows both containers running, and `nighttroop` as `healthy`.
- [ ] `curl https://example.nz/health` prints `{"status":"ok"}`.
- [ ] <https://example.nz> goes to the setup page at `https://control.example.nz/setup`, with a valid padlock.
- [ ] `http://example.nz` redirects to HTTPS.

## 9. Set up the control plane

A new server has no control plane admin, so every page goes to the setup page until you make one. Only someone with
the setup code the server prints to its logs can, so nobody else can take a new server before you do.

1. Find the setup code in the logs:

   ```sh
   docker compose logs nighttroop | grep -A1 "SET UP NIGHTTROOP"
   ```

   - The code is printed on every start until setup is done, so restart Nighttroop to see it again.
2. Open <https://control.example.nz/setup> and enter the code.
3. It shows the admin's username, password and TOTP secret, once:
   1. Store the password and TOTP secret somewhere safe.
   2. Add the TOTP secret to your authenticator app, e.g. by scanning its QR code.
   3. Enter the code from your authenticator app and press **Finish setup**.
4. Log in at <https://control.example.nz>. The rest of the server, such as creating sites, now works.

## Updating Nighttroop safely

Updating works as on Hostinger, with the same script. Deploy a version tag, as
[the Hostinger guide explains](hostinger.md#where-new-versions-come-from), so you always know which version is running:

```sh
cd /opt/nighttroop
sudo ./update.sh ghcr.io/grahamjenson/nighttroop:v1.2.3
```

`update.sh` uses the compose file `.env` names, so it runs `voyager.compose.yml` here. It pulls the new image, keeps
the running one, copies `./data` while Nighttroop is stopped, starts the new version, and rolls back to the previous
image and data if the new one isn't healthy within 3 minutes.

What each step does, its exit codes, what a rollback can't undo, and how to roll back by hand are in
[the Hostinger guide's update section](hostinger.md#updating-nighttroop-safely); they are the same here.

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
- **Backups of the whole VPS:** check in Voyager's VPS Manager, or with Voyager, whether your plan keeps backups or
  snapshots of the VPS, and take one before a big change if it can. They are a whole-machine fallback alongside the
  script's backups and S3.
- **Disk space:**
  - `du -sh /opt/nighttroop/*` shows what the data, backups and failed-update folders use.
  - `docker system df` shows what images use.

## Troubleshooting

### Caddy doesn't get certificates

Look at `docker compose logs caddy` for the error. Common causes:

- **`Invalid request headers`, `Authentication error` or `403` from Cloudflare:** the API token is wrong, has
  expired, or can't edit this domain's DNS. Make a new one as in [step 4](#4-make-a-cloudflare-api-token), put it in
  `.env`, and run `docker compose up -d caddy`.
- **Timeouts waiting for the TXT record:** check the domain really uses Cloudflare's nameservers:
  `dig +short NS example.nz` should show the two `ns.cloudflare.com` names, and Cloudflare should show it as
  **Active**.
- **Too many failed attempts:** Let's Encrypt limits how often a domain can fail. While you fix things, try Let's
  Encrypt's staging service, which has looser limits but gives untrusted certificates:
  1. Add `acme_ca https://acme-staging-v02.api.letsencrypt.org/directory` inside the `tls` block of the Caddyfile in
     `voyager.compose.yml`.
  2. Run `docker compose up -d caddy`.
  3. Remove the line again once certificates are obtained.

### Pages show Cloudflare errors, or everyone shares a few IP addresses

The records are proxied by Cloudflare. Set each of the three to **DNS only** (a grey cloud) in Cloudflare's **DNS**
page; Caddy must answer visitors itself.

### The browser shows a 404 page

Nighttroop only answers host names under `DOMAIN`.

- Check `DOMAIN` in `.env` is exactly your domain, then run `docker compose up -d`.
- A subdomain that isn't a site, such as `www.example.nz`, gets a 404 by design.

### The site can't be reached at all

- `ufw status` should show 80 and 443 allowed.
- If Voyager's VPS Manager has a firewall for the server, it needs the same ports open.
- `docker compose ps` should show Caddy running with `0.0.0.0:80` and `0.0.0.0:443`.

### Nighttroop isn't healthy

Look at `docker compose logs nighttroop`:

- **`cannot create server config`:** a setting has a bad value, such as `TRUSTED_PROXIES`.
- **A setting seems to be ignored:** check its spelling in `nighttroop.env` against
  [`docs/config.md`](../config.md). Settings there reach Nighttroop as environment variables, so a misspelled one is
  ignored and the default is used.
- **Permission denied under `/data`:** the data folder isn't owned by user `10001`:

  ```sh
  chown -R 10001:10001 /opt/nighttroop/data
  ```

- **Still starting after a while:** a first start on a new VPS restores from S3, which takes longer the more data there
  is. Give it a few minutes.
- **`Site DB is missing, so the site is not opened`:** a live site's database file is missing and was not restored
  from its batch's backup in S3, so the server does not start (`live sites' DBs are missing and were not restored:`
  names them). Put its file back, such as from a copy in `./backups`, and restart.

### The `172.30.0.0/24` network is already used on the VPS

Change both the `subnet` and `ip_range` under `networks` in `voyager.compose.yml`, and set Caddy's `ipv4_address` and
`TRUSTED_PROXIES` to the same new address. It must be inside the subnet but outside `ip_range`.
