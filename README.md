# SQLite Multi-tennant experiment

** WARNING: THIS IS PROBABLY A BAD IDEA **

I wanted to try make a minimal multi-tenant application using SQLite, the problem is deploying and backing. SQLite is super fast and simple but since the DB is on disk when deploying the whole... I want to experiment how we can deploy a SQLite multitenant app without sacrificing simplicity or speed. The simplicity of having an entire application in a single executable, no additional DBs or caches, or backup processes to worry about, and speed in that all reads will be local so no network latency required.

The biggest tradeoff will be zero downtime deploys, but hopefuly we can reduce the down time to a minamil amount.

The requirements for the app are:

1. Anyone can create a tenant and immediatly get credentials and a unique subdomain
1. The credentials can be used:
    - to write an int to a scores table
    - fetch latest scores
    - fetch sum of all scores
1. A control plane shows tenants, their database sizes, ordered by last active time.
1. Tenants are deleted if they do not interact in 30 minutes.

Unlike Users Tenants don't interact with one another so sharding is a bit easier. There are no shared resources required, they are all very seperate.

# Deploy

There will be downtime because if you replicate writes from the same db from multiple applications it will corrupt the db.

Deploy flow for same machine:
1. NEW: Start
1. NEW: Try grab Deploy Lock and is blocked.

1. OLD: Stop Serving HTTP
1. OLD: Wait for litesteam to finish all DB replications
1. OLD: Release Deploy Lock
1. OLD: KILL

1. NEW: Deploy Lock grabbed
1. NEW: Download main database
1. NEW: Start HTTP

DBs for each user will be lazily downloaded (if needed) when they next call. This will hopefully stop thundering herd.

This deploy flow should work for both releasing a new version on the same machine AND releasing a new version on a brand new machine. The only difference will be that DB download wont be needed if a new machine is not used.


# Testing locally

```
docker run -p 9000:9000 -p 9001:9001 minio/minio server /data --console-address ":9001"
```

then login at http://localhost:9001 with minioadmin and then create bucket

# Notes

The ControlDB and each UserDB can be created, and then just held without checking. Manually calling the Litestream sync function when a write occurs. Does this function block, if it doesnt then we can just spam it

How long manual sync cost, can force sync on important changes like adding a new user. Which we should check actually uploads the new DB.

Option 1: Open a litestream db each time a user is accessed and leave it open and access it over a period of time and sync after if something changed (I think litestream handles diffs, so just always sync in band)

Option 2: Open a litestream instance for every user, which will have to be closed at the end. Then let the syncing happen automatically (default 1 second)

Option 1: will be slow and will cause sync issues if multiple litestream instances are looking at the same DB

Option 2: Leaving it open we will need to check memory consumption of each db. CPU could be mitigated with increasing the MonitorInterval, although suffering higher probabiliity of lost data in a disaster. This will be a cost incurred on S3 writes as well. Bakc of the envelope pricing for AWS s3 might be $4 per day if every seccond contains a write, $0.4 in cloudflare R2. For hundreds of thousands of users that is still too much.



# LiteStream

10,000 Active Tenants Litestream dbs is 7000MB RAM, because each litestream connection had 2 long running go routines (AFAIK), one that loops to sync the data up defined with `MonitorInterval`, and one that holds open a select statement to stop the SQLite DB from checkpointing without the db knowing. `SyncInterval`
CPU is really the thing that gets hit here. If the `MonitorInterval` is the default at 1 second, the CPU gets destroyed by 10,000 dbs all deciding to Sync. Increasing this to 5 minutes decreases that problem, but even when nothing needs to be synced during these 

We can get rid of 1/2 of the go routines by not monitoring for when a change occurs, and just remember to call Sync somehow if something writes to the db (e.g. maintain a dirty DB list).


# sqlite backup through golang interface

Here we will maintain a list of dirty dbs and then run a background loop every N seconds backing them up. It is not quite streaming replication, in that there map be larger gaps that we might lose information, but we wont have to hold onto so many resources as we will be backing up in bulk.

10,000 tenants kept in a map, with each maintaining an open SQLiteDB connection we are using 2200Mb of memory (which is pretty linear, so 2MB per ). Backing up the empty 10,000 dbs takes 10 seconds synchronously, and checking they need backing up takes 40ms.

Backing up 10k dbs to s3 synchronously takes about 20 seconds for a local minio server.