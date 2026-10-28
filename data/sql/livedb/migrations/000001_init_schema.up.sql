-- Live state the server keeps only while it runs. This DB is always in memory and never backed up, so it starts
-- empty each time the server does.

-- An identity's alarm, kept from one check to the next while it is active, snoozed, or its last notification
-- still matters, all times in milliseconds
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
