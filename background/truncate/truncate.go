package truncate

// TODO this is the background runner that will go over all DBs including control, and truncate to size.
// This should stop abusive users, but also provide a limit.
// A possibility is to limit per user, put the limits in the control/db so that they can be changed if needed
// Truncate shouldnt be based on dates, but on row counts,

// Loop is
// Go over each user DB, check size of DB, if over limit, execute truncation queries on every table to bring back in line
// Then check that the DB has reduced in size, though this might be difficult with the WAL
// TODO write an event for that

// Control DB, check size is not over some limit, execute truncations queries if it is.
