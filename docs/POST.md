# $0.01 Per-User Per-Year: NightTroop

Nightscout is a self-hosted, open-source observability server used to collect real-time diabetic data, such as Blood Glucose Levels (BGL), and share it with apps and devices for monitoring, alerting and analysis. We have a Nightscout server for my T1D son that his phone uploads his BGL and treatment data to. Then our family and his caregivers can download and see on his current state on our phones, watches and other devices. We also upload information to Nightscout that syncs down to his phone to control his insulin delivery, via setting target BGLs or adding carbs. **Nightscout is an integral part of my sons T1D management.**

Self-hosting Nightscout is difficult and requires a lot of technical knowledge that you don't magically get when you're diagnosed with diabetes. This has lead to third parties popping up to host Nightscout, but they can cost as much as $12 per month. 

To help people out and let them save a buck or two I tried hosting my own and other people Nightscout instances, but...

1. Nightscout can only host one persons data per server instance, so you have to run multiple instances of the server and database!
2. Configuration of Nightscout is all over the place, environment variables, dot files, plugins, file flags, DB... Config is annoying, water is also wet; but with nightscout each user has their own config (e.g. timezone) and changing anything requires restarting the server, so the difficulty has layers.
3. Nightscout uses socket.io for real-time communication with its default endpoint at `/socket.io`. So, to support websockets each server needs its own domain, e.g. `host1.example.com/socket.io`, `host2.example.com/socket.io`, making TLS cert, DNS, proxying and routing a more difficult.

I want to make an API compatible (including socket.io) rewrite of Nightscout to be multi-user, I am calling it "NightTroop" (because a group of scouts is a troop).
What makes NightTroop (or probably any nightscout rewrite) different from a typical company or product is that:

1. it would be a wild success if it had 10k users world wide, so no need to worry about hockey-stick hyper-scaling technologies.
1. charging money is a shitty thing to do, and can come off as exploitative to a community already that may have little or no choice.
1. collecting and hosting Personally Identifiable Information (PII) from users can run you amok of HIPPA laws, especially if you charge for the service.
1. users will want it to be open source and self host. The idea of not sharing code that your life may rely on, as well as let them do what they want with it is a good idea.
1. putting hurdles in front of it to sign up for and use, like requiring email, or login can stop some people using it.

For NightTroop to be sustainable it **must be** extremely low complexity and low cost. To make this interesting, lets put an arbitrary number of $0.01 per-user per-year; so to support 10,000 users costs $100 per year.

# The Plan

NightTroop is a multi-tennant Nighscout system. Each "tennent" is called a site, as each site can have multiple users.

## SQLite + Split user DBs

NightTroop needs to be boring, maintainable, easy to setup, low complexity, cheap to run -- a single binary executable with as few external dependencies cant get much more boring.
Using SQLite lets us get rid of using an external database dependency (e.g. MongoDB), but add complexity around deployment and disaster recovery.
A common way to use SQLite is externally backing up snapshots to somewhere like S3, e.g. Litestream.
So if we are moving servers during deployment, or have a disaster and need to reset data, we have it available.

There are two types of data NightTroop is concerned with:

1. Data mastered on the Person With Diabetes (PWD) phone, in an app like AAPS or Loop, e.g. BGL readings and treatments
1. Data that is unique to Nighttroop, e.g. control plane users, authentication tokens, sync information...

The externally mastered information can be backed up at a lower frequency to the NightTroop information, as it can be recovered (at least in the short term)
So separating this information into different SQLite DBs, makes sense, and further separating each sites data makes backing up easier to do incrementally, so not to block network IO or CPU periodically.

## Onboarding + Rate Limiting

Not storing PII of a user makes onboarding super simple, it should be "click a button -> get a site".
Anonominity adds extra work to make sure users are behaving and not over consuming resources.
One [noisy neighbor](https://en.wikipedia.org/wiki/Cloud_computing_issues#Performance_interference_and_noisy_neighbors) could take down performance for everyone.
The solution is to:

1. Limiting user creation and limit total users, e.g. one user per minute created and stop creating users at 10k.
1. Rate limiting API usage and any users background tasks.
1. Limiting returned data length, e.g. returning a months worth of data every query.
1. Limiting the size of user databases by truncating or compressing old information.
1. Archiving and deleting inactive users.

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

## Compute + RAM + Storage

I created a SQLite DB with 100,000 data points, one for every minute(ish) for 2 months and it was 3MB. So for 10k users would be a total of 30GB DBs on disk.

The main RAM usage of the server will be the SQLite in-memory cache, set with `cache_size` parameter. 
The most recent data is the most important, about 6 hours worth, which is about 10KB (on disk).
With `cache_size` set to 10KB and with 10K that is a 100Mb of RAM required for cache plus the base amount to run the app.

The needed compute can be calculated based on the requests per second (rps) per core.
Based in the usage assumptions above, the rps for 10k users is around 2000rps.
I wanted to see some numbers so I deployed to fly.io a simple golang app (described here https://maori.geek.nz/sqlite-http-api-zig-vs-golang-3fc2f89bdc79) that reads from a SQLite DB:

With a single core and 256MB of memory 30GbHDD that is $8.52 per month:
wrk  -d10s -c 1000 https://{site}/read                                                                            
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   227.38ms  216.59ms   1.55s    79.76%
    Req/Sec     2.29k   579.44     3.02k    91.92%
  45266 requests in 10.05s, 99.43MB read
Requests/sec:   4502.25
Transfer/sec:      9.89MB

with 2 cores and 512Mb of memory that is $10.54 per month:
wrk  -d10s -c 1000 https://{site}/read
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   101.94ms   81.99ms   1.98s    94.90%
    Req/Sec     4.67k     1.42k    6.76k    78.17%
  91622 requests in 10.03s, 178.52MB read
  Socket errors: connect 0, read 0, write 0, timeout 1
Requests/sec:   9134.38
Transfer/sec:     17.80MB

That is pretty linear.

I also wanted to quickly do a comparison against a lower level language, Zig, implementation just to make sure i am not leaving a lot on the table,

With a single core and 256MB of memory:
wrk -d10s -c 1000 https://{{site}}/read
  2 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   137.41ms   94.22ms   1.42s    89.06%
    Req/Sec     3.49k     1.30k    5.62k    67.01%
  68575 requests in 10.04s, 180.77MB read
  Non-2xx or 3xx responses: 361
Requests/sec:   6832.73
Transfer/sec:     18.01MB

Just to sanity check those numbers against some cheap hardware, I compared the results to Raspberry Pi 5 with an NVME HDD running on the local network.

 wrk  -d10s -c 1000 http://{site}/read
  2 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   625.07ms  181.96ms   1.41s    62.77%
    Req/Sec   820.35    540.31     2.55k    67.03%
  15457 requests in 10.11s, 78.40MB read
Requests/sec:   1528.45
Transfer/sec:      7.75MB

Given that Pi cost about $200, I think it would be much more efficient to run on fly.io and deal with the hassle of my own hardware.

## Backup Pricing for S3-ish services

The prices for S3, Cloudflare R2, and Tigris are:

NAME      COST PER GB/month   COST PER WRITE 1MM  COST PER READ 1MM    EGRESS     INGRES
Tigris    $0.02                 $5                       $0.5            FREE      FREE
R2        $0.015                $4.50                    $0.36           FREE      FREE
S3        $0.023                $5                       $0.4            0.09/GB   FREE      

As you can see, the cost per GB is nothing. If we kept 10 full snapshots of each user's DB it would still be at most $2.3 per month.
The cost of reads will only happen on deploys to new machines, which will be low.

The writes are where they get-ya. If we backup every users DB individually every 5 minutes, that would be $450 per month. 
This is why projects like Litestream wont work for us, it is cheap for a single DB but gets very expensive the more DBs you have.
The solution I came up with is to segment sites into batches (based on letter they start with), then backup and zip them all together.
10k users, batched into 26 groups, backed-up every 5 minutes is about $1 per month.

## Network Costs

On fly.io in New Zealand, bandwidth is $0.04 per GB. So with 10K users polling like me (very agressivly), and returning 1kb of data per call on average, with `users * amount of data * devices connected` that is: 

1. 1,667 requests per second 
2. 144GB of data per day 
3. that is $172 per month, $2,000 per year

This cost estimate surprised me. If everyone used *socket.io* instead of polling, this might drop the cost some but not that much. 
Looking at other hosting options, like Hetzner, has significantly lower price per/GB, basically free for this scale but has an increase in cost to compute at least $480 per year. As well as time and effort managing secrets, deployment...

Other network costs, like the Domain ($15 per year) and wildcard cert ($12 per year) are rounding errors compared to bandwidth egress costs.

# Conclusion

Fly.io costs per year is a minimum of $100 per year, with a massive $0.2 per user because of bandwidth. A hosting company like Hetzner base cost is $480 with a per year,
and basically free for additional users.

The cross over point is then about 2000 users, when fly.io gets more expensive, but each user at that point costs $0.25 per user.
On Hetzner the cost for 10k users is then **$0.05 cents per user per year** 5x what I originally was hoping for. 

In both instances this is a conservative estimate though:

1. The site databases will store more data per minute for longer, increasing HDD and RAM costs.
1. The users will spam much more aggressively, increasing CPU and bandwidth costs.
1. External costs like observability and error tracking each cost around $30 per month.
1. [Jevons Paradox](https://en.wikipedia.org/wiki/Jevons_paradox), where no matter what efficiencies I make, the demand and use will increase. Make it faster, users will just poll more frequently.

Anyway, these are all just guesses I started this project with. However, for the last year I have been actually building this system,
and everything is harder and more expensive than even I assumed. Post coming soon.









--------------------------------------------------------------
# Implementation

## Syncing
I want to be able to sync data from existing services, that way this service can be trialed and tested without worrying about it being stable. 
Then when comfortable it is working a user can migrate over to directly write to the service.
The wo services I need to pull from are:
1. Nightscout sites
1. Dexcom share 



## Deploy and Backup

1. Check all users DBs exist
2. If any don't, download that users pool
3. 




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


