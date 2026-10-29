-- Live state the server keeps only while it runs. This DB is always in memory and never backed up, so it starts
-- empty each time the server does.

-- An identity's alarm, kept from one check to the next while it is active, snoozed, or its last notification
-- still matters, all times in milliseconds
CREATE TABLE alarms (
    site_id TEXT NOT NULL CHECK(length(site_id) <= 32),
    identity_id INTEGER NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('urgent_low', 'low', 'urgent_high', 'high', 'no_data', 'device_battery')),
    -- active_since is when it became active, or 0 while it is not
    active_since INTEGER NOT NULL,
    notified_at INTEGER NOT NULL,
    snoozed_until INTEGER NOT NULL,
    -- alerting is whether it is active and neither snoozed nor quieted by a louder alarm, so should be heard
    alerting INTEGER NOT NULL CHECK(alerting IN (0, 1)),
    PRIMARY KEY (site_id, identity_id, type)
) STRICT;

-- A connection a site's identity holds open, kept while it is open, so a site can hold only so many. An event stream
-- is one of a collection's.
CREATE TABLE connections (
    id INTEGER PRIMARY KEY,
    site_id TEXT NOT NULL CHECK(length(site_id) <= 32),
    identity_id INTEGER NOT NULL,
    kind TEXT NOT NULL CHECK(kind IN ('event_stream')),
    collection TEXT NOT NULL CHECK(length(collection) <= 64),
    opened_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
) STRICT;

CREATE INDEX connections_site ON connections (site_id, kind);
