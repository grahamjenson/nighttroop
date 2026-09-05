# Implementing Nighttroop

So to test this out I have to actually impllement it. This has lots of features that are going to impact the assumptions in the previous post.

1. CPU and RAM usage for things like encryption/decrpytion of backups, also bcrypt checking of credentials

1. CPU usage for checking tokens on nightscout was horrible
1. Getting 1-1 nightscout API is very difficult. The query structure passing to mongo, the worst issue was SQL Order as dynamic order is not really supported in sql queries, and it caches very badly.

List what is supported and what is not.

SQLC Goverter... 

# Problems

The WAL takes up a lot of space on disk

➜  nighttroop git:(main) ls -halt .dbs/db/site/a 
total 8632
-rw-r--r--   1 grahamjenson  staff   4.0M Sep  4 14:40 alicezzz.sqlite-wal
-rw-r--r--   1 grahamjenson  staff   216K Sep  4 14:40 alicezzz.sqlite
-rw-r--r--   1 grahamjenson  staff    32K Sep  4 14:40 alicezzz.sqlite-shm
drwxr-xr-x   5 grahamjenson  staff   160B Sep  4 14:40 .
drwxr-xr-x  28 grahamjenson  staff   896B Sep  4 14:40 ..
➜  nighttroop git:(main) ls -halt .dbs/backup/site 
total 0
drwxr-xr-x  3 grahamjenson  staff    96B Sep  4 14:42 a
drwxr-xr-x  4 grahamjenson  staff   128B Sep  4 14:42 ..
drwxr-xr-x  3 grahamjenson  staff    96B Sep  4 14:41 .
➜  nighttroop git:(main) ls -halt .dbs/backup/site/a 
total 456
drwxr-xr-x  3 grahamjenson  staff    96B Sep  4 14:42 .
-rw-r--r--  1 grahamjenson  staff   228K Sep  4 14:42 alicezzz.sqlite
drwxr-xr-x  3 grahamjenson  staff    96B Sep  4 14:41 ..
➜  nighttroop git:(main) 

