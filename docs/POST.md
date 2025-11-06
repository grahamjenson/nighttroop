# $0.01 Per-User Per-Year: NightTroop

Nightscout is a self-hosted, open-source observability server used to collect real-time diabetic data, such as Blood Glucose Levels (BGL), and share it with apps and devices for monitoring, alerting and analysis. We have a Nightscout server for my T1D son, his phone uploads his BGL to Nightscout that I display on my watch and phone, my wife has on her phone, and his school has on a "follow" phone.

Self-hosting Nightscout is difficult and requires a ton of technical knowledge that you don't magically get when diagnosed with diabetes. Third parties have popped up to host Nightscout for you and charge as much as $12 per month. As a recovering SysAdmin I wanted to host my own Nightscout and maybe host other peoples as well, but Nightscout is a PITA because:

1. Nightscout can only host one persons data per server instance, so you have to run multiple instances of the same server. So it is like managing and maintaining a host of servers so scale is an issue.
1. Each instance requires its own Node process and MongoDB database, so is very resource intensive to scale.
1. Configuration of Nightscout is all over the place, environment variables, dot files, file flags, DB... This makes setup and configuration complex and a custom configuration layer necessary and difficult to build, and changing configs often requires restarting the server.
1. Nightscout uses socket.io for real-time communication with its default endpoint at `/socket.io`. So, to support websockets each server needs its own domain, and so TLS cert, DNS, special proxying and routing...

A lot of these issues can only be solved with a ground up, API compatible (including socket.io) rewrite of Nightscout to be multi-user.
So I am going to try make "NightTroop" (a group of scouts is a troop), and see if I can get something working.

What makes NightTroop a unique project is that:

1. it probably has 10k users world wide, so no need to worry about hockey-stick hyper-scaling technologies.
1. I don't want to make money from it, so I don't want to spend money on it.
1. I don't want to collect Personally Identifiable Information (PII) from the users. Since it is hosting health data, I don't want any concern of misuse.
 
For NightTroop to be sustainable it has to be low complexity and low cost, so the goal is for each user cost $0.01 per-year; 10,000 users is a $100 per year. 
Lets see how close it can get.

# The Plan

## Golang + SQLite + Split user DBs

NightTroop needs to be boring, maintainable, easy to setup, low complexity, cheap to run -- so a single golang binary with SQLite built in fits that description.
Using SQLite means no external DB server (low complexity, low cost) and comes with the benefit of being lightning fast.

The biggest tradeoff using SQLite is the difficulty of deployment and disaster recovery because there is no built in replication or high availability.
To make a trustworthy service there will need to be external storage of SQLite DBs.
Given recent data is so much more important than past data, periodic (say every 5 minutes) snapshots should be fine. 
Snapshotting then uploading one giant SQLite DB would block IO then network.
Sharding data over many USER DB's solves blocking and a few other problems, e.g. managing 10k 2MB DBs should be easier than managing 1 giant 20GB database.

## Onboarding + Rate Limiting

Not storing PII of a user makes onboarding super simple, it should be as simple as clicking a button and getting access.
The problem with the anonominity is that it becomes extra work to make sure users are behaving and not over consuming resources.
The solution is to:

1. Limiting user creation and limit total users, e.g. one user per minute created and stop creating users at 10k.
1. Rate limiting API usage and any users background tasks.
1. Limiting returned data length, e.g. returning a months worth of data every query.
1. Limiting the size of user databases by truncating old information.
1. Archiving and deleting inactive users (easier with sharded user DBs).

All of these are the extra work needed to keep the costs down, but should make resource usage more predictable.

# Back of the envelope pricing for 10K users

With the above plan we can start estimating some costs, which break down into:

1. Compute cost from estimated requests per second, RAM usage, and volume storage required.
1. Backup cost from the frequency of snapshots, and size of user DB.
1. Other costs: domain + certs + egress + bandwidth

Lets also put some real numbers to assumptions so that we can estimate these values:

1. A user writes a value once per minute
2. A user (like me) has multiple devices polling for information, so reads are 10 requests per minute.
3. Most recent information (2 days) is the most important, and we can truncate old (2 months) data

I am going to use fly.io as an estimate for cloud costs, because I am familiar with that and understand the costs.

## Compute 

I created a SQLite DB with 100,000 data points, one for every minute(ish) for 2 months and it was 3MB, so a DB for 10k users would be a total of 30GB DBs on disk.

Requests per second for 10k would be around 2000 requests per second. I have benchmarked a single core fly.io machine doing 250 requests per second (https://maori.geek.nz/golang-sqlite-on-fly-io-with-litefs-a-quick-benchmark-9aa7aeee9561), so lets say we need 8 cores.

The main RAM usage of the server is the SQLite in-memory cache for each users DB connection, set with `cache_size` parameter. 
A users most of the time will only access the last 2 days data, which is about 100KB (on disk).
With `cache_size` set to 100KB and with 10K that is a minimum of 1GB of RAM required, lets make that 2GB for all other uses.

Lets see how much Fly.io would cost for 8 cores, 2Gb RAM, 30GB volume: $22.42 per month, $270 per year

https://fly.io/calculator?f=s&b=ams.100&a=no_none&m=0_0_0_0_0&r=shared_0_1_syd&t=10_100_5&u=0_1_100&g=8_shared_730_1_2000_ams_0_30

This is about the cost of a pretty big machine at Hetzner, or the price of a RaspberryPi with NVME. Either way, we have already blown out our budget.

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

## Other

A Domain will cost $15 per year and a wildcard cert is $1 per month on fly.io. 
Nighttroop show be relativly low bandwidth since it wont be serving any media.

# Implementation

## Deploy and Backup

1. Check all users DBs exist
2. If any don't, download that users pool
3. 

# Conclusion
$0.01 per User per year as a goal was never going to be realistic. There are so many costs associated with a service like this like error tracking, logging, not to mention my time developing.
















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

10,000 users kept in a map, with each maintaining an open SQLiteDB connection we are using 2200Mb of memory (which is pretty linear, so 2MB per). Each user also may use cache set by `cache_size` which might add 2Mb more memory usage per user. So lets assume 4Mb per user memory usage.

Backing up the empty 10,000 dbs takes 10 seconds synchronously, and checking they need backing up takes 40ms.

Backing up 10k dbs to s3 synchronously takes about 20 seconds for a local minio server.

Test Restore, downloading 10k empty databases and then opening them all up

A user with 54,000 data points (a datum every 5 minutes for 6 months) creates a 1.5MB Sqlite DB. Lz4 seems to reduce this size by about 1/3 so 500KB. Lets double that to 3Mb as a cap for the users db size, getting compressed to 1MB, or with 10K users that is 30GB on disk and 10GB upload total.

On my local mackbook that took about 1-2mins to lz4 then upload. So, Doing all 10k users as a single file isnt flexible enough and doing as individual files isnt cost effective. I think grouping the users by starting char, then having a sync and restore 

Once you move to pools of users, lazy restore from S3 becomes much harder, so to simplify I will restore all user pools before starting http. This will mean that deploy and startup time will probably be delayed by the amount of time to download all user pools, but that should be easy to make concurrent. Each pool is 1/36 * 10,000 * 1Mb , so about 300MB.

The time it takes to backup and upload up 10,000 dbs is about 3 minutes, but doing so concurrently takes all CPUs at which time it cannot server HTTP requests. Staggering pool backups will help with this, but incure a time penalty. The time it takes to download and restore 10k users is significantly less.

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
1. A single bad acting user could take a lot of resources, DB size, request thoughput. The fix would be to rate limit the DBs, and to truncate the DB tables to some value. To limit DB size.

Back Office stuff

Deleting old users in s3.

reducing cost of s3 usage

# Pricing estimates
Pricing for 10k users

Lets say we cap the user DB size at 5Mb, and truncate all past data. It will matter how fast data comes in as well, as each sync operation will cost. 

Control DB size will be negligble, each user DB would be 10MB

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

The solution may be compressing all or groups of users with lz4, before uploading, e.g. compressing based on first letter in user id would reduce the upload cost to $12 per month, but would cost more in complexity (especially with lazily fetching users to speed up deploys), and CPU and HDD resources. But it would speed up the sync process in total given it would be making less network calls.





# Alternatives
Turso https://turso.tech/blog/give-each-of-your-users-their-own-sqlite-database-b74445f4
LiteStream

https://sqlsync.dev/posts/stop-syncing-everything/


# LiteStream

10,000 empty user databases, Litestream is 7GB RAM, because each litestream connection had 2 long running go routines (AFAIK), one that loops to sync the data up defined with `MonitorInterval`, and one that holds open a select statement to stop the SQLite DB from checkpointing without the db knowing. `SyncInterval`. If these databases had any data in them the ram usage would be much larger.

CPU is really the thing that gets hit here. If the `MonitorInterval` is the default at 1 second, the CPU gets destroyed by 10,000 dbs all deciding to Sync. Increasing this to 5 minutes decreases that problem, but even when nothing needs to be synced during these 

Then there is the cost, Litestream is constantly writing to S3, which will be very expencive. For a single DB it might be a few dollars a month, but that multiplied by 10k is wallet emptyng.

We can get rid of 1/2 of the go routines by not monitoring for when a change occurs, and just remember to call Sync somehow if something writes to the db (e.g. maintain a dirty DB list).





# Noisy Neighbors/Rate Limiting

There are a bunch of ways in which a noisy neighbor can ruin this service. The benefit of running the entire thing in a single process means that at least we wont need any common horizontal rate limiting tools based on redis or shared locks.

Request rate:
"golang.org/x/time/rate" is a really easy way to stop a user from calling too many times, allowing for a common rate and a burst rate with `rate.NewLimiter(limit, burst_limit)`. This should stop resource consumption

User limit
A single server has a hard limit of how many users it can handle, 10k. So a hard limit on creating more users should be done.

DB Size
A hard limit on row numbers for each table, triggered when the db goes over a certain size.

User Age:
Unused users use up space in the system, take time in deploys and backups. Users not used for 30 minutes should be archived to S3, and soft deleted from the control DB, and zipped up and referenced. This means that there is a path to rehydration, if someone wants to get their data back.


