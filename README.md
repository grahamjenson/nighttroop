# TODO

1. EPIC IS SYNCING SGVS FROM NS AND DEXCOM AND GETTING THAT DATA IN CARER
    1. Identities Page To retrieve NS token and information on communication
    1. Create non-admin device for care portal role
    1. generate and offer copy token button to call api
    1. Save data from sync as entry
    1. Save data tag with unique sync identity that created it (sgv is tagged with DexcomSync identity, so that we can trace where stuff comes from)
    1. fetch data in ns store to return to the user
    1. make models a bit nicer, maybe with generics e.g. Record[Entry] or something
    1. write punishment loop for bad syncing
    1. write wrapper for cron to report runtimes
    1. Enable turning off s3 (or making it local for testing)
    1. Deploy to a fly.io dev/staging environment
    1. Connect actual data and do some fun perf testing
    1. Cleanup overloaded terms User/Device/Identity

1. Timezones for profiles, user settings, just to make sure that when a treatment is logged we know whne against the profile it actually happened
2. Logging with contexts to DBs, 
1. error reporting with sentry, 
1. check that users dont get error responses traces
1. Nightscout reimplementation routes
5. Write an endpoint to add NightscoutURL and/OR dexcom login/share creds for a user
1. Write data sync logic and protections, punishing bad servers...
1. Archive has a problem where they are deleting files from the db not removing from control
1. archivied_at in the user table in control.db

# NightTroop

NightTroop is a server that hosts many NightScout servers at the same time.
It is a ground up rewrite of the NightScout server to make it easier and more scalable to host many users.

# On Boarding

Go to nighttroop.t1d.nz and click the button to create your user.
You can sync data from Dexcom and/or Nightscout instance (exactly one of each per user), 
so you can fill out your dexcom username/password + 


# Testing locally

Docker runs our local test s3 server `minio` (the only external dependency)

```
brew install sqlc
brew install templ
brew install go-air
brew install tailwindcss
brew install golang-migrate
brew install docker colima docker-compose
colima start
docker-compose up -d
```

```
go run main.go
```

Then go to `localhost:8080` and create a user, `control.localhost:8080` and login as control.
In Development all initial passwords are `password` and all TOTP is `123123`


# TODO:
1. Login Counter ratelimiter per IP and per username
1. NS endpoints
1. Dashboard/Profile/Security endpoints/pages for control and user
1. Sentry integration with errors/logs/metrics
1. Translations
1. app for multiple people at a time

The goal is to make an app for diabetes camps that can add all the kids and see their numbers and show alarms/highlight issues.




# Syncing
Each request set the `to-be-synced` to be `true`;
Write an infinite go routine per user every 30 seconds checks `to-be-synced` and will fetch data if `true`.


------------------------------------------------------------------------------------------------------------------------
------------------------------------------------------------------------------------------------------------------------
------------------------------------------------------------------------------------------------------------------------

This section below is for development of Nighttroop

# Local Development

# Running a Development server

# Testing

# Submitting changes

------------------------------------------------------------------------------------------------------------------------
------------------------------------------------------------------------------------------------------------------------
------------------------------------------------------------------------------------------------------------------------

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