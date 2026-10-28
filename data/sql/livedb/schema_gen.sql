CREATE INDEX connections_site ON connections (site_id, kind);

CREATE TABLE alarms (
    site_id TEXT NOT NULL CHECK(length(site_id) <= 32),
    identity_id INTEGER NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('urgent_low', 'low', 'urgent_high', 'high', 'no_data')),
    -- active_since is when it became active, or 0 while it is not
    active_since INTEGER NOT NULL,
    notified_at INTEGER NOT NULL,
    snoozed_until INTEGER NOT NULL,
    -- alerting is whether it is active and neither snoozed nor quieted by a louder alarm, so should be heard
    alerting INTEGER NOT NULL CHECK(alerting IN (0, 1)),
    PRIMARY KEY (site_id, identity_id, type)
) STRICT;

CREATE TABLE connections (
    id INTEGER PRIMARY KEY,
    site_id TEXT NOT NULL CHECK(length(site_id) <= 32),
    identity_id INTEGER NOT NULL,
    kind TEXT NOT NULL CHECK(kind IN ('event_stream')),
    collection TEXT NOT NULL CHECK(length(collection) <= 64),
    opened_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
) STRICT;

CREATE TABLE nightscout_cursors (
    site_id TEXT NOT NULL CHECK(length(site_id) <= 32),
    collection TEXT NOT NULL CHECK(collection IN ('entries', 'treatments', 'devicestatus')),
    source TEXT NOT NULL CHECK(length(source) = 64),
    last_modified INTEGER NOT NULL,
    PRIMARY KEY (site_id, collection)
) STRICT;

