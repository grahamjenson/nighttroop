# TODO

1. JWT implementation middleware for both control and user
2. Nightscout reimplementation routes
3. make sure tink is thread safe


# NightTroop

NightTroop is a server that hosts many NightScout servers at the same time.
It is a ground up rewrite of the NightScout server to make it easier and more scalable to host many users.

# On Boarding

Go to nighttroop.t1d.nz and click the button to create your user.
You can sync data from Dexcom and/or Nightscout instance (exactly one of each per user), 
so you can fill out your dexcom username/password + 


# Deploying to a service
It comes ready to deploy to fly.io as a service. 
You should be able to run this for as cheap as a few hundred dollars a year and support 10k users.

If you dont want to host you can get a user from another host like nighttroop.t1d.nz.

# Testing locally

```
docker run -p 9000:9000 -p 9001:9001 minio/minio server /data --console-address ":9001"
```

Then

```
go run main.go --dev
```




# Goals:
1. 


The goal is to make an app for diabetes camps that can add all the kids and see their numbers and show alarms/highlight issues.


Write the nightscout basic endpoints.
Write an endpoint to add NightscoutURL and/OR dexcom login/share creds for a user
Have a property per user of `to-be-synced`. 
Each request set the `to-be-synced` to be `true`;

Write an infinite go routine per user every 30 seconds checks `to-be-synced` and will fetch data if `true`.


Write an app, let it take a name and a picture, add a user. Then call the service create a user



# Security 

We generate two keysets, one for AES256GCM and one streaming version of that for files

```
export KEY_SETS=$(go run utils/keysets/create_keysets.go)
```

Send these keysets to fly
```
fly secrets set KEY_SETS=$KEY_SETS
```

NOTE: Fly.io secrets are not easily read, so keep the KEY_SETS ENVAR saved offline.


One the first run of the server, it will create an admin user and print the encrypted password and totp data.
```
export ENCRYPTED_ADMIN_BUNDLE=<base64encodeddata>
export KEY_SETS=<defined above>
```

```
go run utils/keysets/decrypt_admin_bundle.go
```

### Security Operations
add a new key and roll the old key

```
go run utils/keysets/decrypt_admin_bundle.go "<base64 encoded admin info>"
```

This will add the new key and make it primary, so then all new values will be encrypted with that new primary value
TODO as encrypted values are read, if they are on an old key they will be moved over.
