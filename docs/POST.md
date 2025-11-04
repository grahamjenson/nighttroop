# $0.01 Per-User Per-Year

Nightscout is a self-hosted, open-source observability server used to collect real-time diabetic data, such as Blood Glucose Levels (BGL), and share it with apps and devices for monitoring, alerting and analysis. We have a nightscout server for my son, his phone uploads his BGL to Nightscout which I download and display on my watch and phone, my wife has on her phone, his teachers have on their phones and we also have a sugarpixel display under the TV and an ipad next to my computer.

Self-hosting Nightscout is difficult and requires a ton of technical knowledge (even if you RTFM), so third parties have popped up and charge as much as $12 per month to host Nightscout for you. However, hosting many instances of Nightscout is a PITA because:

1. Nightscout can only host one persons data per server instance, so you have to run multiple instances of the same server. So management and control is an issue.
1. Each instance requires its own Node process and MongoDB database, so is very resource intensive.
1. Configuration of Nightscout is all over the place, environment variables, dot files, file flags, DB... This makes setup and configuration complex and a custom configuration layer necessary and difficult to build, and changing configs requires restarting the server.
1. Nightscout uses socket.io for real-time communication with its default `/socket.io` namespace, so each server needs its own domain. This makes certs and routing difficult.

I figure that a lot of these issues can only be solved with a ground up, API compatible (including socket.io) rewrite of Nightscout to be multi-user.
What makes this a unique project is that:

1. it probably has 10k users world wide, so no need to worry about hockey-stick hyper-scaling.
1. I don't want to make money from it, so I don't want to spend money on it.
1. I don't want to collect Personally Identifiable Information (PII) from the users, it is hosting health data so I don't want any of that hassle
 
I want the project to be sustainable, both in low complexity to setup and in hosting costs.
I figure that if each user cost $0.01 per-year, 10,000 users will only cost me $100 a year. 
Lets see how close I can get.

# The Plan: Golang + SQLite + Split user DBs + Onboarding + Rate limiting

I need a boring, maintainable, easy to setup, low complexity, cheap to run server -- so a single golang binary with SQLite built in.
SQLite means no external DB server (low complexity, low cost) and comes with the benefit of being lightning fast.

The biggest tradeoff using SQLite is the difficulty of deployment and disaster recovery because there is no built in replication or high availability.
This means that either a planned or unplanned move of physical machines will require an file snapshot of SQLite DBs stored externally somewhere like S3.
Having one giant database to snapshot, and upload periodically seems fragile and difficult compared to having each user have their own DB with sharded uploads.
There are streaming options like LiteStream, but I think a periodic upload should be fine for this project.

Not storing PII of a user makes onboarding super simple, it should be as simple as clicking a button and getting access.
The problem with the anonominity is that it becomes extra work to make sure users are behaving and not over consuming resources by:

1. Limiting user creation rate and total, e.g. one user per minute created and stop creating users at 10k.
1. Rate limiting all user API and background tasks
1. Limiting the size of user databases by truncating old information.
1. Archiving and deleting inactive users.

All of these are the extra work needed to keep the costs down.

# Back of the envelope pricing for 10K users

With the above implementation we have some variables to start estimating potential costs:

1. Frequency of Backups for User databases
2. Size of User databases
3. Requests per second

A SQLite DB with 50,000 data points (one for every minute in a month) creates a 1.5MB Sqlite DB, zipped it is 1/3rd the size at 500KB.
So if we kept 2 months of data for 10k users, that is 30GB on disk and 10GB zipped for backup. 
Given we will need seperate backup files for all DBS, that is a total of 70GB storage.

If we model the 10K users after how I use Nightscout with one device writing and multiple devices polling for new data, lets say 10 requests per minute.
For 10k users thats about 1500 requests per second that this server should handle. I have benchmarked a single core fly.io machine doing 250 requests per second.

To keep things simple, the server keeps open the 10K SQLite DB connections, the biggest resource used is the memory used for the cache set with `cache_size`.
So if each user has a 100KB cache limit, which is plenty since the majority of requests will only be for recent data, then that is 1GB of memory. So maybe need 2GB of RAM.

Lets see how much Fly.io would cost for this, lets get 6 cores just to be safe.
6 cores, 2Gb of RAM, 70GB storage, 

https://fly.io/calculator?m=0_0_0_0_0&f=s&b=syd.100&a=no_none&r=shared_0_1_syd&t=10_100_5&u=0_1_100&g=6_shared_730_1_2000_syd_0_70

That is about $32 per month. So already blowing our budget there with $384 per year.
For that price you could get a much larger Intel i7 2x512GB SSDs at Hetzner.
Self hosting on a $100 Raspberry pi + $100 NVME hat + drive, but you would also have to pay for internet, power...

## Backup Pricing for S3-ish services

The prices for S3, Cloudflare R2, and Tigris are:

NAME      COST PER GB/month   COST PER WRITE 1MM  COST PER READ 1MM    EGRESS     INGRES
Tigris    $0.02                 $5                       $0.5            FREE      FREE
R2        $0.015                $4.50                    $0.36           FREE      FREE
S3        $0.023                $5                       $0.4            0.09/GB   FREE      

The cost per GB is nothing. If we kept 10 versions of each user's DB it would still be at most $2.3 per month.
The cost of reads will only happen on deploys to new machines, which I hope is rare.

The writes will be where the money is spent. If we backup every users DB individually every 5 minutes, that would be $450 per month.

The solution is bundling the user DBs together before uploading, e.g. each user is in a pool for each letter (a-z), that is 26 letters that brings the write cost down to $1 per month.

# Implementation

## Deploy and Backup

1. Check all users DBs exist
2. If any don't, download that users pool
3. 

##
















--------------------------------------------------------------

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

This deploy flow should work for both releasing a new version on the same machine AND releasing a new version on a brand new machine. The only difference will be that DB download wont be needed if a new machine is not used.

The deploy time to a clean server (worst case) is then 

# Testing locally

```
docker run -p 9000:9000 -p 9001:9001 minio/minio server /data --console-address ":9001"
```

then login at http://localhost:9001 with minioadmin and then create bucket




# sqlite backup through golang interface

Here we will maintain a list of dirty dbs and then run a background loop every N seconds backing them up, also running another slower loop that takes the backups and sends them to S3. It is not quite streaming replication, in that there map be larger gaps that we might lose information, but we wont have to hold onto so many resources as we will be backing up in bulk.

10,000 tenants kept in a map, with each maintaining an open SQLiteDB connection we are using 2200Mb of memory (which is pretty linear, so 2MB per). Each user also may use cache set by `cache_size` which might add 2Mb more memory usage per user. So lets assume 4Mb per user memory usage.

Backing up the empty 10,000 dbs takes 10 seconds synchronously, and checking they need backing up takes 40ms.

Backing up 10k dbs to s3 synchronously takes about 20 seconds for a local minio server.

Test Restore, downloading 10k empty databases and then opening them all up

A tenant with 54,000 data points (a datum every 5 minutes for 6 months) creates a 1.5MB Sqlite DB. Lz4 seems to reduce this size by about 1/3 so 500KB. Lets double that to 3Mb as a cap for the users db size, getting compressed to 1MB, or with 10K users that is 30GB on disk and 10GB upload total.

On my local mackbook that took about 1-2mins to lz4 then upload. So, Doing all 10k tenants as a single file isnt flexible enough and doing as individual files isnt cost effective. I think grouping the tenants by starting char, then having a sync and restore 

Once you move to pools of tenants, lazy restore from S3 becomes much harder, so to simplify I will restore all tenant pools before starting http. This will mean that deploy and startup time will probably be delayed by the amount of time to download all tenant pools, but that should be easy to make concurrent. Each pool is 1/36 * 10,000 * 1Mb , so about 300MB.

The time it takes to backup and upload up 10,000 dbs is about 3 minutes, but doing so concurrently takes all CPUs at which time it cannot server HTTP requests. Staggering pool backups will help with this, but incure a time penalty. The time it takes to download and restore 10k tenants is significantly less.

Time without staggering is 3mins for all 36 pools, but when staggered it is 3s per 36 pools. we are saving nearly half the time by spreading out the uploads.

Restoring all Dbs from s3 takes 25s. Backing Up all Dbs when killing the service takes at worst 1.5mins.
That means that basically a deploy downtime is around 2 minutes if booting to a clean machine.

Improved performance by only uploading pools where something has changed. 



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


# LiteStream

10,000 empty tenant databases, Litestream is 7GB RAM, because each litestream connection had 2 long running go routines (AFAIK), one that loops to sync the data up defined with `MonitorInterval`, and one that holds open a select statement to stop the SQLite DB from checkpointing without the db knowing. `SyncInterval`. If these databases had any data in them the ram usage would be much larger.

CPU is really the thing that gets hit here. If the `MonitorInterval` is the default at 1 second, the CPU gets destroyed by 10,000 dbs all deciding to Sync. Increasing this to 5 minutes decreases that problem, but even when nothing needs to be synced during these 

Then there is the cost, Litestream is constantly writing to S3, which will be very expencive. For a single DB it might be a few dollars a month, but that multiplied by 10k is wallet emptyng.

We can get rid of 1/2 of the go routines by not monitoring for when a change occurs, and just remember to call Sync somehow if something writes to the db (e.g. maintain a dirty DB list).





# Noisy Neighbors/Rate Limiting

There are a bunch of ways in which a noisy neighbor can ruin this service. The benefit of running the entire thing in a single process means that at least we wont need any common horizontal rate limiting tools based on redis or shared locks.

Request rate:
"golang.org/x/time/rate" is a really easy way to stop a user from calling too many times, allowing for a common rate and a burst rate with `rate.NewLimiter(limit, burst_limit)`. This should stop resource consumption

Tenant limit
A single server has a hard limit of how many tenants it can handle, 10k. So a hard limit on creating more tenants should be done.

DB Size
A hard limit on row numbers for each table, triggered when the db goes over a certain size.

Tenant Age:
Unused tenants use up space in the system, take time in deploys and backups. Tenants not used for 30 minutes should be archived to S3, and soft deleted from the control DB, and zipped up and referenced. This means that there is a path to rehydration, if someone wants to get their data back.


