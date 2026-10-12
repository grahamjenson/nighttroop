CREATE TABLE control_identities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    name TEXT NOT NULL UNIQUE CHECK(length(name) <= 32),

    password TEXT NOT NULL CHECK(length(password) <= 128),
    totp_secret TEXT NOT NULL CHECK(length(totp_secret) <= 128),
    role TEXT NOT NULL CHECK(length(role) <= 32),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    last_seen_at INTEGER,
    last_write_at INTEGER,
    
    updated_by_identity_id INTEGER, -- id of identity that updated this record
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32), -- denormalized name for easier lookup (also survives delete)

    created_by_identity_id INTEGER NOT NULL, -- id of identity that created this record
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32),

    deleted_at INTEGER -- soft delete here, but on delete will wipe out password and such
) STRICT;

CREATE TABLE data_source_block_list(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    username TEXT CHECK(length(username) <= 32), -- in case it is a dexcom account or something
    url TEXT CHECK(length(url) <= 256),
    reason TEXT CHECK(length(reason) <= 256)
) STRICT;

CREATE TABLE site_identities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id TEXT NOT NULL REFERENCES sites(id),

    name TEXT NOT NULL CHECK(length(name) <= 32),

    password TEXT NOT NULL CHECK(length(password) <= 128),
    totp_secret TEXT NOT NULL CHECK(length(totp_secret) <= 128),
    role TEXT NOT NULL CHECK(length(role) <= 32),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    last_seen_at INTEGER,
    last_write_at INTEGER,
    
    updated_by_identity_id INTEGER, -- id of identity that updated this record
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32), -- denormalized name for easier lookup (also survives delete)

    created_by_identity_id INTEGER NOT NULL, -- id of identity that created this record
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32),

    deleted_at INTEGER, -- soft delete here, but on delete will wipe out password and such
    enabled INTEGER NOT NULL DEFAULT 1,

    ---------------
    -- IDENTITY SETTINGS
    ---------------
    units TEXT CHECK (units IN ('mmol/L', 'mg/dL')), -- mg/dL, mmol/L

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
    alarm_urgent_low INTEGER,

    alarm_low_enabled INTEGER,
    alarm_low INTEGER,
    alarm_low_minutes INTEGER,
    alarm_low_prediction_enabled INTEGER,

    alarm_high_enabled INTEGER,
    alarm_high INTEGER,
    alarm_high_minutes INTEGER,
    
    alarm_urgent_high_enabled INTEGER,
    alarm_urgent_high INTEGER,

    alarm_fast_rise_enabled INTEGER,
    alarm_fast_rise INTEGER,
    alarm_fast_rise_minutes INTEGER,

    alarm_fast_drop_enabled INTEGER,
    alarm_fast_drop INTEGER,
    alarm_fast_drop_minutes INTEGER,

    alarm_no_data_enabled INTEGER,
    alarm_no_data_minutes INTEGER,

    alarm_battery_low_enabled INTEGER,
    alarm_battery_low_percent INTEGER,
    alarm_battery_low_pwd_enabled INTEGER,
    alarm_battery_low_pwd_percent INTEGER,

    alarm_low_reservoir_enabled INTEGER,
    alarm_low_reservoir_units  INTEGER,

    -- NOTIFICATIONS
    notifications_enabled INTEGER,

    -- Constraints
    UNIQUE(site_id, name)
) STRICT;

CREATE TABLE site_sync_sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- Nightscout/Dexcom
    site_id TEXT NOT NULL REFERENCES sites(id), -- only for nightscout, dexcom can leave null
    type TEXT NOT NULL CHECK(type IN ('NIGHTSCOUT', 'DEXCOM', 'VIRTUAL')),
    
    url_or_username TEXT NOT NULL CHECK(length(url_or_username) <= 256),
    password TEXT NOT NULL CHECK(length(password) <= 256), -- or ns token

    -- Specific Info
    region TEXT CHECK(region IN ('US', 'OUS')), -- dexcom at the moment


    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    enabled INTEGER DEFAULT 1, -- 0 or 1 for now, can be used to disable without deleting

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

    consecutive_successful_syncs INTEGER DEFAULT 0,
    consecutive_failed_syncs INTEGER DEFAULT 0,

    -- Previous Error Text
    previous_sync_errors TEXT CHECK(length(previous_sync_errors) <= 5012), -- stored to show to user as feedback

    -- Timing EMAs to then determine frequency on
    total_ema REAL,
    success_ema REAL, -- https://en.wikipedia.org/wiki/Exponential_smoothing
    failure_ema REAL,
    
    -- Duration of last sync in ms
    last_sync_duration INTEGER,

    -- Unique contraint of site_id and type
    UNIQUE(site_id, type)
) STRICT;

CREATE TABLE sites (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    batch TEXT GENERATED ALWAYS AS (SUBSTR(id, 1, 1)) VIRTUAL NOT NULL,
    
    backup_key_hash TEXT NOT NULL CHECK(length(backup_key_hash) <= 64), 
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    

    max_db_size_kb_limit INTEGER DEFAULT 3000, -- default 3Mb, can be changed in control plane
    

    -- last_seen_at is default null because on creation they have not been seen, will be deleted quickly is not logged in sooner
    last_seen_at INTEGER,
    last_write_at INTEGER,

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
    alarm_urgent_low INTEGER,

    alarm_low_enabled INTEGER,
    alarm_low INTEGER,
    alarm_low_minutes INTEGER,
    alarm_low_prediction_enabled INTEGER,

    alarm_high_enabled INTEGER,
    alarm_high INTEGER,
    alarm_high_minutes INTEGER,
    
    alarm_urgent_high_enabled INTEGER,
    alarm_urgent_high INTEGER,

    alarm_fast_rise_enabled INTEGER,
    alarm_fast_rise INTEGER,
    alarm_fast_rise_minutes INTEGER,

    alarm_fast_drop_enabled INTEGER,
    alarm_fast_drop INTEGER,
    alarm_fast_drop_minutes INTEGER,

    alarm_no_data_enabled INTEGER,
    alarm_no_data_minutes INTEGER,

    alarm_battery_low_enabled INTEGER,
    alarm_battery_low_percent INTEGER,
    alarm_battery_low_pwd_enabled INTEGER,
    alarm_battery_low_pwd_percent INTEGER,

    alarm_low_reservoir_enabled INTEGER,
    alarm_low_reservoir_units  INTEGER,

    -- NOTIFICATIONS
    notifications_enabled INTEGER

) STRICT;

CREATE TABLE sqlite_sequence(name,seq);

