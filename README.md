# TODO

## Before Alpha Testing

1. Define Nighttroop models, work on converters to DB params, and use those models to build json and htmx and html. Then can test all this across any form
1. Focus on sync settings slow down on bad and incorrect behaviour
1. rename all columns with type suffix e.g. password_hash, totp_secret_enc
1. Use custom methods to replace 	DurationSum:        stats.DurationSum, DurationSquaredSum: stats.DurationSquaredSum,

-- Errors
1. Come up with an actual logging and error framework that makes sense (SQLite DB for search?)
1. Force all returned errors to be of type
1. Optional error reporting with sentry (if a given key is present)

1. make sync settings report endpoints
1. make archiving and truncation a manual job, i.e. show unused users and select all and archive.  I think this will remove a lot of logic that can be shaky later on we can look at automating, if it gets really bad
1. Display users in control
1. Display job and cron stats in control
1. Sync
    1. Check Sync Creds before save/enable -- e.g. check nightscout is a real endpoint and auth works
    1. Sync treatments and entries using v3 API
    1. Use v3 last_modified nightscout sync, and existing nightscout_api to write the values to the DB

1. create_identity check limits, max identities
1. Control UI
    1. Table of users with DB sizes, status...
    1. enable disable, users (show screen on disable), archive, and truncate buttons
    1. Some kind of usage metrics to see

1. Deleting an identity should soft delete but hard remove passwords (names should be unique per user)

1. Site UI
    1. Show real data (dont bother with graph, just tables for now)
    1. Improve identity creation. Make it easier to encourage each device and app to get its own identity

1. Deploy to a fly.io dev/staging environment

## Before Beta Testing

1. Backfill button on sync page for Dexcom, NS (one off creds would be cool, so I can pull data from multiple sources )
1. archivied_at in the user table in control.db

1. Reports/Derived Values:
    1. AGP
    1. TTD
    1. % High/Low... 
    1. Small multiples Daily 

1. Alarms:
    1. Create Alarms
    1. Define main types
    1. Notifications for device/identity

1. Site UI:
    1. Config Timezone settings
    1. Timezones for profiles, user settings, just to make sure that when a treatment is logged we know when against the profile it actually happened

1. Bad Sites
    1. write punishment loop for bad syncing
    1. write entry compression
    1. write device status compression
    1. write treatments compressions
    1. limit number of identities
    1. Login Counter ratelimiter per IP and per username

1. Control UI
    1. Show bad users
    1. Block IP addresses
    1. Show all rate limits for users

1. API
    1. Write API endpoints for sync and identity

## Before First Release:

1. Think about Translations
1. Docs hosting
1. Fuzz testing
1. Performance testing (cost estimate for running full node)

## Bugs

1. Archive has a problem where they are deleting files from the db not removing from control


# Nighttroop

Nighttroop is primarly a simple, free live diabates dashboard that will quickly get you up and running.

The tradeoffs:
1. Historical data
1. Complexity with setups.



This section below is information for running your own nighttroop

------------------------------------------------------------------------------------------------------------------------

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