# The Tragedy Of The Commons

NightTroop is free and shared, and anonymous and can do some pretty interesting things. Here is a list of the things we must protect against.



### Sync
Users can sync data from a few sources, none more dangerous than the nightscout URL.



### DB Size

Any place a user can input some data, SQLite doesnt have a strict limit on a column size so we must add it, other wise they could dump wikipedia as a username.

1. All TEXT Columns need a Check limit on them 

SQLite is also not typed, but sqlc generated typed code so we shouldn't need to check user types are correct, e.g. user dumps wikipedia into an INT column.

After some time legitim

### HTTP

Rate Limit everything.

1. IP Address
1. User
1.

Data can be expensive so make sure large files are cached.


### Free Users

Rate limit global user creation, as well as by IP.

Delete users if they are not logged into (quickly)

Archive users if they are not logged into for a long period

Limit Total users
