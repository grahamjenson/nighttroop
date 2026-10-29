# TODO

## Before Alpha Testing

1. Deploy a staging server ([docs/deploy/hostinger.md](docs/deploy/hostinger.md)) with S3 backups and `SENTRY_DSN` set,
   and check a fresh server restores from those backups
1. Test end to end with real devices: Loop, Trio, AAPS and xDrip+ uploading, follower apps (xDrip+ follower, LoopFollow,
   Nightguard) reading, and Dexcom and Libre sync
1. Medical disclaimer: Nighttroop is not a medical device and its alarms are not to be relied on, shown when a site is
   created and on the alarm settings
1. Privacy policy and terms, linked from site creation: what is kept (credentials encrypted, logs), for how long, and
   who can see it
1. A site's admin can delete their site and all its data (sites are only deleted now when unused)
1. A way for testers to report bugs and give feedback, linked from the site
1. A tester guide: creating a site, keeping its credentials safe (they cannot be recovered), and connecting uploaders,
   followers and sync
1. Check the archive bug below is fixed

## Before Beta Testing

1. Reports/Derived Values:
    1. AGP
    1. Time below and above range, as well as time in range
    1. Small multiples Daily

1. Alarms:
    1. Push notifications to phones, as alarms now only reach an open site page or a Nightscout socket client

1. Bad Sites
    1. write entry compression
    1. write device status compression
    1. write treatments compressions

1. Control UI
    1. Show bad users
    1. Block IP addresses
    1. Show all rate limits for users
    1. Restore one site from its backup

1. Accounts
    1. Recovering a site's admin without the initial backup key

1. Security
    1. Move encrypted values on an old key to the primary key as they are read

## Before First Release:

1. Think about Translations
1. Docs hosting
1. Fuzz testing beyond the API (`tests/api/fuzz_test.go`)
1. Performance testing (cost estimate for running full node)

## Bugs

1. Archive has a problem where they are deleting files from the db not removing from control


# Nighttroop

Nighttroop is a multi user rewrite of Nightscout.
Its core goals are to be:

1. Simple         - Single binary deployment, limited and optional external dependencies
1. Cheap Per User - It should consume very little resources per user
1. Multi-user     - A single server should support 10,000 users

It is written in Golang using SQLite to make it a single binary.




This section below is information for running your own nighttroop

------------------------------------------------------------------------------------------------------------------------

# Deployment: Hostinger VPS with Docker

[docs/deploy/hostinger.md](docs/deploy/hostinger.md) runs Nighttroop behind Caddy, with wildcard Let's Encrypt
certificates, from the image CI publishes to `ghcr.io`, and updates it with a script that rolls back if the new version
isn't healthy.

# Deployment: Fly.io

Fly.io Is a platform. This comes with the needed info to deploy the project

fly.toml.example is the example app description used to define the resources need
Something like this:

```
fly apps create my-unique-app-name --org <your-organization-name>
fly deploy -a my-unique-app-name
```
 
You can edit the fly.toml 

Create the s3 bucket for the app with 

```
fly storage create
```


------------------------------------------------------------------------------------------------------------------------

# Observability TODO

## Error Collection: Sentry
Sentry is backed into the code just need to add a sentry envar to see that calls

## Logs:  Fly.io Logs

## Metrics: Fly.io Metrics


------------------------------------------------------------------------------------------------------------------------

# Security 

We generate three keysets, a AES256GCM for secrets encryption, one for File encryption and one for JWT.

```
export KEY_SETS=$(go run utils/keysets/create_keysets.go)
```

Send these keysets to fly
```
fly secrets set KEY_SETS=$KEY_SETS
```

NOTE: Fly.io secrets are not easily read, so keep the KEY_SETS ENVAR saved OFFLINE!
// TODO generate QR code image to print


One the first run of the server, it will create an admin user and print the encrypted password and totp data.
```
export ENCRYPTED_ADMIN_BUNDLE=<base64encodeddata>
export KEY_SETS=<defined above>
```

```
go run utils/keysets/decrypt_admin_bundle.go
```

This will output the information to the admin account for the service.

### Security Operations
add a new key and roll the old key

```
go run utils/keysets/decrypt_admin_bundle.go "<base64 encoded admin info>"
```

This will add the new key and make it primary, so then all new values will be encrypted with that new primary value
TODO as encrypted values are read, if they are on an old key they will be moved over.


# GoTHSS Framework. (Go + Templ + HTMx + SQLite + SQLc + Tailwind)

The "framework" that Nighttrrop uses uses Go and relies heavily on generated code tools like SQLc, Templ and Tailwind.