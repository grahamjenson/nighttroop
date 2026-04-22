# TODO

2. Nightscout reimplementation routes
3. make sure tink is thread safe
5. Write an endpoint to add NightscoutURL and/OR dexcom login/share creds for a user
1. Write data sync logic and protections, punishing bad servers...

# NightTroop

NightTroop is a server that hosts many NightScout servers at the same time.
It is a ground up rewrite of the NightScout server to make it easier and more scalable to host many users.

# On Boarding

Go to nighttroop.t1d.nz and click the button to create your user.
You can sync data from Dexcom and/or Nightscout instance (exactly one of each per user), 
so you can fill out your dexcom username/password + 




# Testing locally

Docker runs our local test s3 server `minio`

```
brew install docker colima docker-compose
colima start
docker-compose up
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
