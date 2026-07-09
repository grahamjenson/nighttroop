# Implementing Nighttroop

So to test this out I have to actually impllement it. This has lots of features that are going to impact the assumptions in the previous post.

1. CPU and RAM usage for things like encryption/decrpytion of backups, also bcrypt checking of credentials

1. CPU usage for checking tokens on nightscout was horrible
1. Getting 1-1 nightscout API is very difficult. The query structure passing to mongo, the worst issue was SQL Order as dynamic order is not really supported in sql queries, and it caches very badly.

List what is supported and what is not.

SQLC Goverter... 