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


