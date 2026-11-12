-- The control DB should have enough information such that if we lost 100% of the
-- site data we could recover enough information to restart the server
-- The only thing the site should see is the loss of data, and not loss of access

------------------
--- Control Tables
------------------


--------------------------
--- SITE TABLES
--------------------------

-- MAIN SITE TABLE

-- unique ID
-- subdomain is the unique identifier for the site and is used in the URL
-- pool is the first charater of the subdomain
-- backup key is provided to the site on creation, used as proof of ownership in case of lost information
-- only soft delete Sites


CREATE TABLE sites (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    batch TEXT GENERATED ALWAYS AS (SUBSTR(id, 1, 1)) VIRTUAL NOT NULL,
    
    backup_key_hash TEXT NOT NULL CHECK(length(backup_key_hash) <= 64), 
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    

    max_db_size_kb_limit INTEGER DEFAULT 20480, -- default 20 MB, can be changed in control plane
    estimated_db_size_kb INTEGER, -- the site DB's size on disk, measured by the resource check and with its last write
    -- each table of the site's DB by name, with the bytes it and its indexes take and its records, as last measured from
    -- SQLite's dbstat, so a page shows them without reading the whole DB
    table_sizes_json TEXT CHECK(length(table_sizes_json) <= 16384 AND json_valid(table_sizes_json)),
    

    -- last_seen_at is default null because on creation they have not been seen, will be deleted quickly is not logged in sooner
    last_seen_at INTEGER,
    last_write_at INTEGER,
    -- the site's data written up to then, by its updated_at, is in its summaries, as the stale rebuild last checked
    summaries_checked_at INTEGER,

    -- States
    state TEXT NOT NULL CHECK (state IN ('live', 'disabled', 'archived', 'deleted')) DEFAULT 'live',
    -- live is all good, disabled no-login DB still accessible, archived means remove copy available, deleted means no rehydration available
    disabled_at INTEGER,
    archived_at INTEGER,
    deleted_at INTEGER,

    ---------------
    -- SITE SETTINGS
    ---------------
    units TEXT CHECK (units IN ('mmol/L', 'mg/dL')), -- mg/dL, mmol/L
    -- how the Nightscout API writes each units, as some clients read only one spelling
    units_mgdl_label TEXT CHECK (units_mgdl_label IN ('mg/dL', 'mg/dl')),
    units_mmol_label TEXT CHECK (units_mmol_label IN ('mmol/L', 'mmol')),

    -- DATE
    timezone TEXT CHECK(length(timezone) <= 64),
    language TEXT CHECK(length(language) <= 12),
    date_format TEXT CHECK(length(date_format) <= 24),
    time_format TEXT CHECK(length(time_format) <= 24),
    week_start_day TEXT CHECK (length(week_start_day) <= 16),

    -- INSULIN
    bolus_dia_hours REAL,
    bolus_onset_minutes REAL,
    bolus_peak_minutes REAL,

    basal_dia_hours REAL,
    basal_onset_minutes REAL,
    basal_peak_minutes REAL,

    default_insulin_sensitivity REAL,
    default_carb_ratio REAL,
    default_basal_rate REAL,

    -- CARBS
    carb_max_absorption_hours REAL,
    carb_min_5m_impact REAL,


    -- ALERTS
    alarm_urgent_low_enabled INTEGER,
    range_urgent_low_mgdl INTEGER,

    alarm_low_enabled INTEGER,
    range_low_mgdl INTEGER,
    alarm_low_minutes INTEGER,
    alarm_low_prediction_enabled INTEGER,
    alarm_low_prediction_minutes INTEGER, -- how far ahead the current trend is followed to the low alarm's glucose

    alarm_high_enabled INTEGER,
    range_high_mgdl INTEGER,
    alarm_high_minutes INTEGER,
    
    alarm_urgent_high_enabled INTEGER,
    range_urgent_high_mgdl INTEGER,

    alarm_fast_rise_enabled INTEGER,
    alarm_fast_rise_mgdl INTEGER,
    alarm_fast_rise_minutes INTEGER,

    alarm_fast_drop_enabled INTEGER,
    alarm_fast_drop_mgdl INTEGER,
    alarm_fast_drop_minutes INTEGER,

    alarm_no_data_enabled INTEGER,
    alarm_no_data_minutes INTEGER,

    alarm_device_battery_low_enabled INTEGER,
    alarm_device_battery_low_percent INTEGER,
    alarm_device_battery_low_minutes INTEGER,
    alarm_device_battery_low_devices_json TEXT CHECK(length(alarm_device_battery_low_devices_json) <= 1024 AND json_valid(alarm_device_battery_low_devices_json)), -- the devices whose batteries and statuses are checked

    alarm_low_reservoir_enabled INTEGER,
    alarm_low_reservoir_units  INTEGER,

    -- NOTIFICATIONS
    notifications_enabled INTEGER

) STRICT;


-- SITE TABLES
CREATE TABLE control_identities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    name TEXT NOT NULL UNIQUE CHECK(length(name) <= 32),

    password_hash TEXT NOT NULL CHECK(length(password_hash) <= 128),
    totp_secret_enc TEXT NOT NULL CHECK(length(totp_secret_enc) <= 128),
    role TEXT NOT NULL CHECK(length(role) <= 32),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    last_seen_at INTEGER,
    last_write_at INTEGER,
    
    updated_by_identity_id INTEGER, -- id of identity that updated this record
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32), -- denormalized name for easier lookup (also survives delete)

    created_by_identity_id INTEGER NOT NULL, -- id of identity that created this record
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32),

    deleted_at INTEGER, -- soft delete here, but on delete will wipe out password and such

    -- credentials_version counts its password and TOTP changes: a JWT carries it, and one from an older version is refused
    credentials_version INTEGER NOT NULL DEFAULT 0
) STRICT;

-- This table is like users and devices
CREATE TABLE site_identities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id TEXT NOT NULL REFERENCES sites(id),

    name TEXT NOT NULL CHECK(length(name) <= 32),

    password_hash TEXT NOT NULL CHECK(length(password_hash) <= 128),
    totp_secret_enc TEXT NOT NULL CHECK(length(totp_secret_enc) <= 128),
    role TEXT NOT NULL CHECK(length(role) <= 32),
    -- client is the app a device was added for, e.g. aaps or loop, or other, so its settings can be changed by app;
    -- null for a user or a site's first devices
    client TEXT CHECK(length(client) <= 32),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    last_seen_at INTEGER,
    last_write_at INTEGER,
    
    updated_by_identity_id INTEGER, -- id of identity that updated this record
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32), -- denormalized name for easier lookup (also survives delete)

    created_by_identity_id INTEGER NOT NULL, -- id of identity that created this record
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32),

    enabled INTEGER NOT NULL DEFAULT 1,
    -- credentials_version counts its password and TOTP changes: a JWT carries it, and one from an older version is refused
    credentials_version INTEGER NOT NULL DEFAULT 0,
    -- nightscout_token_digest finds the identity from the SHA1 of its Nightscout token, which xDrip+ sends as its
    -- api-secret. It is the SHA256 of that SHA1, so it is not what is sent.
    nightscout_token_digest TEXT CHECK(length(nightscout_token_digest) = 64),

    ---------------
    -- IDENTITY SETTINGS
    ---------------
    units TEXT CHECK (units IN ('mmol/L', 'mg/dL')), -- mg/dL, mmol/L
    -- how the Nightscout API writes each units, as some clients read only one spelling
    units_mgdl_label TEXT CHECK (units_mgdl_label IN ('mg/dL', 'mg/dl')),
    units_mmol_label TEXT CHECK (units_mmol_label IN ('mmol/L', 'mmol')),

    -- DATE
    timezone TEXT CHECK(length(timezone) <= 64),
    language TEXT CHECK(length(language) <= 12),
    date_format TEXT CHECK(length(date_format) <= 24),
    time_format TEXT CHECK(length(time_format) <= 24),
    week_start_day TEXT CHECK (length(week_start_day) <= 16),

    -- INSULIN
    bolus_dia_hours REAL,
    bolus_onset_minutes REAL,
    bolus_peak_minutes REAL,

    basal_dia_hours REAL,
    basal_onset_minutes REAL,
    basal_peak_minutes REAL,

    default_insulin_sensitivity REAL,
    default_carb_ratio REAL,
    default_basal_rate REAL,

    -- CARBS
    carb_max_absorption_hours REAL,
    carb_min_5m_impact REAL,


    -- ALERTS
    alarm_urgent_low_enabled INTEGER,
    range_urgent_low_mgdl INTEGER,

    alarm_low_enabled INTEGER,
    range_low_mgdl INTEGER,
    alarm_low_minutes INTEGER,
    alarm_low_prediction_enabled INTEGER,
    alarm_low_prediction_minutes INTEGER, -- how far ahead the current trend is followed to the low alarm's glucose

    alarm_high_enabled INTEGER,
    range_high_mgdl INTEGER,
    alarm_high_minutes INTEGER,
    
    alarm_urgent_high_enabled INTEGER,
    range_urgent_high_mgdl INTEGER,

    alarm_fast_rise_enabled INTEGER,
    alarm_fast_rise_mgdl INTEGER,
    alarm_fast_rise_minutes INTEGER,

    alarm_fast_drop_enabled INTEGER,
    alarm_fast_drop_mgdl INTEGER,
    alarm_fast_drop_minutes INTEGER,

    alarm_no_data_enabled INTEGER,
    alarm_no_data_minutes INTEGER,

    alarm_device_battery_low_enabled INTEGER,
    alarm_device_battery_low_percent INTEGER,
    alarm_device_battery_low_minutes INTEGER,
    alarm_device_battery_low_devices_json TEXT CHECK(length(alarm_device_battery_low_devices_json) <= 1024 AND json_valid(alarm_device_battery_low_devices_json)), -- the devices whose batteries and statuses are checked

    alarm_low_reservoir_enabled INTEGER,
    alarm_low_reservoir_units  INTEGER,

    -- NOTIFICATIONS
    notifications_enabled INTEGER,

    -- Constraints
    UNIQUE(site_id, name)
) STRICT;

-- CONNECTORS: a site's third party integrations, each syncing from one service

-- This is the table that the site can update and delete from
CREATE TABLE site_connectors (
    -- Nightscout/Dexcom
    site_id TEXT NOT NULL REFERENCES sites(id), -- only for nightscout, dexcom can leave null
    type TEXT NOT NULL CHECK(type IN ('nightscout', 'dexcom', 'libre', 'virtual')),
    
    url_or_username_enc TEXT CHECK(length(url_or_username_enc) <= 256),
    password_enc TEXT CHECK(length(password_enc) <= 256), -- or ns token

    -- Specific Info
    region TEXT CHECK(region IN ('US', 'OUS', 'GLOBAL', 'AE', 'AP', 'AU', 'CA', 'DE', 'EU', 'EU2', 'FR', 'JP', 'LA')), -- dexcom US or OUS, or a libre region in consts.LibreRegions

    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    enabled INTEGER DEFAULT 1, -- 0 or 1 for now, can be used to disable without deleting
    first_successful_sync_time INTEGER, -- when a sync first worked since the last save; while NULL a failed sync disables the source

    -- BACKFILL, never reset so a source that backfills too much can be slowed or blocked
    backfill_started_at INTEGER, -- set while a backfill is queued or running; another is refused until it finishes or an hour passes
    last_backfill_from INTEGER, -- the oldest time the last backfill fetched from
    last_backfill_to INTEGER, -- and the time it fetched up to
    last_backfill_finished_at INTEGER,
    last_backfill_duration INTEGER, -- in ms
    last_backfill_records INTEGER, -- new records it added
    last_backfill_error TEXT CHECK(length(last_backfill_error) <= 1024),
    total_backfills INTEGER NOT NULL DEFAULT 0,

    -- REPORTING

    -- Frequency of syncs used to punish bad servers or behaviour
    sync_interval_seconds_reason TEXT NOT NULL DEFAULT 'default' CHECK(length(sync_interval_seconds_reason) <= 256), -- reason for the current sync interval, can be default, user_request, error_response, etc. Used to determine if we should change the sync interval after a sync attempt
    sync_interval_seconds INTEGER NOT NULL DEFAULT 60, -- how often to sync in seconds

    -- Timing of syncs
    last_attempted_sync INTEGER,
    last_successful_sync INTEGER,
    last_failed_sync INTEGER,

    -- Counters
    total_syncs INTEGER DEFAULT 0,
    total_successful_syncs INTEGER DEFAULT 0,
    total_failed_syncs INTEGER DEFAULT 0,
    total_records INTEGER NOT NULL DEFAULT 0, -- the records its syncs and backfills have added in all
    last_sync_records INTEGER, -- the records its last successful sync added

    consecutive_successful_syncs INTEGER DEFAULT 0,
    consecutive_failed_syncs INTEGER DEFAULT 0,

    -- Previous Error Text
    previous_sync_errors TEXT CHECK(length(previous_sync_errors) <= 5012 AND json_valid(previous_sync_errors)), -- stored to show to user as feedback

    -- Timing EMAs to then determine frequency on
    total_ema REAL,
    success_ema REAL, -- https://en.wikipedia.org/wiki/Exponential_smoothing
    failure_ema REAL,
    
    -- Duration of last sync in ms
    last_sync_duration INTEGER,

    -- Unique contraint of site_id and type
    UNIQUE(site_id, type)
) STRICT;




-- PUSH NOTIFICATIONS

-- A credential pushes are sent with: an Apple push (APNs) key or a Firebase (FCM) service account. One with a site_id
-- is that site admin's, used only for its identities, such as their own build of Loop; one without is a control
-- admin's, used for every site, such as a published app.
CREATE TABLE push_credentials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id TEXT REFERENCES sites(id), -- NULL is for every site

    name TEXT NOT NULL CHECK(length(name) <= 64),
    provider TEXT NOT NULL CHECK(provider IN ('apns', 'fcm')),
    team_or_project_id TEXT NOT NULL CHECK(length(team_or_project_id) <= 64), -- APNs team ID, or Firebase project ID
    key_id TEXT CHECK(length(key_id) <= 32), -- APNs key ID; FCM has none
    secret_enc TEXT CHECK(length(secret_enc) <= 8192), -- the .p8 key or service account JSON, encrypted

    apps_json TEXT NOT NULL CHECK(length(apps_json) <= 1024 AND json_valid(apps_json)), -- the apps it pushes to, e.g. ["loop"]
    bundle_ids_json TEXT CHECK(length(bundle_ids_json) <= 2048 AND json_valid(bundle_ids_json)), -- the only bundle IDs or package names it pushes to; NULL is any

    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
) STRICT;

CREATE INDEX push_credentials_site ON push_credentials (site_id, provider);

-- An app on a device that an identity has pushes sent to. Its token addresses that app on that device, and changes
-- when the app is installed again; one is registered to the last of its site's identities that sent it, and another
-- site registering it has its own row, so no site can take another's.
CREATE TABLE push_registrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id TEXT NOT NULL REFERENCES sites(id),
    identity_id INTEGER NOT NULL REFERENCES site_identities(id) ON DELETE CASCADE,

    app TEXT NOT NULL CHECK(length(app) <= 32), -- consts.PushApp, e.g. loop
    provider TEXT NOT NULL CHECK(provider IN ('apns', 'fcm')),
    bundle_id TEXT NOT NULL CHECK(length(bundle_id) <= 255), -- iOS bundle ID or Android package name
    production INTEGER NOT NULL DEFAULT 1 CHECK(production IN (0, 1)), -- APNs production, else sandbox (a development build)

    token_enc TEXT NOT NULL CHECK(length(token_enc) <= 2048), -- encrypted
    token_hash TEXT NOT NULL CHECK(length(token_hash) = 64), -- SHA-256 of the token, in lower case, to find it without decrypting

    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    last_pushed_at INTEGER,
    last_error TEXT CHECK(length(last_error) <= 256),

    UNIQUE (site_id, token_hash)
) STRICT;

CREATE INDEX push_registrations_identity ON push_registrations (site_id, identity_id);

-- A passkey a site's identity, or a control admin, signs in with. Only what checking a sign in needs is kept: the
-- credential's ID, and in credential_json its public key, flags and sign count, without the authenticator's model or
-- attestation. No name, email or device of the person is kept, and the user handle given to the authenticator is
-- made from the IDs. A site identity's has its site and identity, and a control admin's only its control identity, as
-- each query that adds one sets.
CREATE TABLE passkeys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id TEXT REFERENCES sites(id),
    identity_id INTEGER REFERENCES site_identities(id) ON DELETE CASCADE,
    control_identity_id INTEGER REFERENCES control_identities(id) ON DELETE CASCADE,
    credential_id TEXT NOT NULL CHECK(length(credential_id) <= 1366), -- base64url of the ID, at most 1023 bytes
    credential_json TEXT NOT NULL CHECK(length(credential_json) <= 4096 AND json_valid(credential_json)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    last_used_at INTEGER,
    UNIQUE (site_id, credential_id)
) STRICT;

CREATE INDEX passkeys_identity ON passkeys (site_id, identity_id);
CREATE UNIQUE INDEX passkeys_control ON passkeys (credential_id) WHERE control_identity_id IS NOT NULL;

-- SERVER SETTINGS
-- The server's one row of settings, which its control admin changes
CREATE TABLE server_settings (
    id INTEGER PRIMARY KEY CHECK(id = 1),

    -- Whether the public can create sites from the home page; a control admin always can
    public_site_creation INTEGER NOT NULL DEFAULT 1 CHECK(public_site_creation IN (0, 1)),
    -- The rate limits a control admin set, by name, each {"per_minute": n, "burst": n}, used in place of the env's
    rate_limits_json TEXT NOT NULL DEFAULT '{}' CHECK(length(rate_limits_json) <= 4096 AND json_valid(rate_limits_json)),

    updated_at INTEGER,
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32)
) STRICT;

INSERT INTO server_settings (id) VALUES (1);
