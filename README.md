# SQLite Multi-tennant experiment

** WARNING: THIS IS PROBABLY A BAD IDEA **

Not every product has to be high availability, and no every piece of data needs to be treated like a precious gem and replicated around the planet for safety.

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

Unlike Users, Tenants don't interact with one another so sharding is a bit easier. There are no shared resources required, they are all very seperate.

This higher level requirements are the application should be able to have 10k tenants where:
1. each tenant uploads on average a piece of data every 5 minutes. So support 33 reqs/s min.
1. each tenant only costs $1 total per year to host
1. best effort to save tenants data, backed up at least every 5 minutes.

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

The deploy time to a clean server (worst case) is then 

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

10,000 empty tenant databases, Litestream is 7GB RAM, because each litestream connection had 2 long running go routines (AFAIK), one that loops to sync the data up defined with `MonitorInterval`, and one that holds open a select statement to stop the SQLite DB from checkpointing without the db knowing. `SyncInterval`. If these databases had any data in them the ram usage would be much larger.

CPU is really the thing that gets hit here. If the `MonitorInterval` is the default at 1 second, the CPU gets destroyed by 10,000 dbs all deciding to Sync. Increasing this to 5 minutes decreases that problem, but even when nothing needs to be synced during these 

Then there is the cost, Litestream is constantly writing to S3, which will be very expencive. For a single DB it might be a few dollars a month, but that multiplied by 10k is wallet emptyng.

We can get rid of 1/2 of the go routines by not monitoring for when a change occurs, and just remember to call Sync somehow if something writes to the db (e.g. maintain a dirty DB list).


# sqlite backup through golang interface

Here we will maintain a list of dirty dbs and then run a background loop every N seconds backing them up, also running another slower loop that takes the backups and sends them to S3. It is not quite streaming replication, in that there map be larger gaps that we might lose information, but we wont have to hold onto so many resources as we will be backing up in bulk.

10,000 tenants kept in a map, with each maintaining an open SQLiteDB connection we are using 2200Mb of memory (which is pretty linear, so 2MB per). Each user also may use cache set by `cache_size` which might add 2Mb more memory usage per user. So lets assume 4Mb per user memory usage. Maybe using the ATTACH feature in SQLite would use less RAM, but I am not sure of other implications related to backup and such. 

Backing up the empty 10,000 dbs takes 10 seconds synchronously, and checking they need backing up takes 40ms.

Backing up 10k dbs to s3 synchronously takes about 20 seconds for a local minio server.

Test Restore, downloading 10k empty databases and then opening them all up

A tenant with 54,000 data points (a datum every 5 minutes for 6 months) creates a 1.5MB Sqlite DB. Lz4 seems to reduce this size by about 1/3 so 500KB. Lets double that to 3Mb as a cap for the users db size, getting compressed to 1MB, or with 10K users that is 30GB on disk and 10GB upload total.

On my local mackbook that took about 1-2mins to lz4 then upload. So, Doing all 10k tenants as a single file isnt flexible enough and doing as individual files isnt cost effective. I think grouping the tenants by starting char, then having a sync and restore 

Rather than having a Backup Go routing and a Sync Goroutine, I think having a Go routine per tenant group would where it would backup then sync.

Once you move to pools of tenants, lazy restore from S3 becomes much harder, so to simplify I will restore all tenant pools before starting http. This will mean that deploy and startup time will probably be delayed by the amount of time to download all tenant pools, but that should be easy to make concurrent. Each pool is 1/36 * 10,000 * 1Mb , so about 300MB.

The time it takes to backup and upload up 10,000 dbs is about 3 minutes, but doing so concurrently takes all CPUs at which time it cannot server HTTP requests. Staggering pool backups will help with this, but incure a time penalty. The time it takes to download and restore 10k tenants took 4m

# DR

The server process is asked to die, via SIGTERM: 

The server process dies unexpectidly and needs to be restarted. This should be fine, because the WAL operations will keep the existing databases safe.

The server dies unexpectidly and we need to move to a new machine: 


If the server dies naturally with a SIGTERM then their should be enough time to backup and restart on a 


# Left to figure out

Migrations
1. I would put the version of the DB of the user in the control database, then when accessed check against a constant value to make sure it is up to date. 

Blast Radius
1. A single bad acting tenant could take a lot of resources, DB size, request thoughput. The fix would be to rate limit the DBs, and to truncate the DB tables to some value. To limit DB size.

Back Office stuff

Deleting old users in s3.

reducing cost of s3 usage

# Pricing estimates
Pricing for 10k users

Lets say we cap the user DB size at 5Mb, and truncate all past data. It will matter how fast data comes in as well, as each sync operation will cost. 

Control DB size will be negligble, each tenent DB would be 10MB

S3, Cloudflare R2, Tigris

NAME      COST PER GB/month   COST PER WRITE 1MM  COST PER READ 1MM    EGRESS     INGRES
Tigris    $0.02                 $5                       $0.5            FREE      FREE
R2        $0.015                $4.50                    $0.36           FREE      FREE
S3        $0.023                $5                       $0.4            0.09/GB   FREE      

Lets say 100% of users will add data every 5 mins and we sync every 5 mins then we will be 
uploading 50GB every 5 mins, with 10k writes, so 3MM writes per day, or 90MM writes per month, so like $450 per month.

Total data will be 50GB, so about $1 per month.

Reads will only happen per deploy, which, lets say averages once per day, so like 300k per month, so like $0.1

As you can see, the size of the database is not really the issue (other than egress costs from the server and resources used to upload and process.) The killer on the cost is the amount of writes, which the only lever is how often to backup. Every 5 mins may be acceptable in some cases, but in others maybe not. 

The solution may be compressing all or groups of tenants with lz4, before uploading, e.g. compressing based on first letter in tenant id would reduce the upload cost to $12 per month, but would cost more in complexity (especially with lazily fetching tenants to speed up deploys), and CPU and HDD resources. But it would speed up the sync process in total given it would be making less network calls.





# Alternatives
Turso https://turso.tech/blog/give-each-of-your-users-their-own-sqlite-database-b74445f4
LiteStream

https://sqlsync.dev/posts/stop-syncing-everything/



# Lets Talk the real use case

I want what is essentially a free IOT server where a single user wears multiple devices and sensors are reporting data frequently. Missing a few readings wont matter, what matters is getting the most immediate readings quickly. I want to make this server so cheap that I can fund it myself 