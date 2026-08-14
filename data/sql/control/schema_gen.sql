CREATE UNIQUE INDEX dedupe ON jobs (deduping_key)
WHERE
    deduping_key != ''
    AND job_status IN ('queued', 'fetched');

CREATE INDEX todo ON jobs (queue, job_status, execute_after)
WHERE
    job_status = 'queued'
    OR job_status = 'fetched';

CREATE TABLE data_source_block_list(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    username TEXT CHECK(length(username) <= 32), -- in case it is a dexcom account or something
    url TEXT CHECK(length(url) <= 256),
    reason TEXT CHECK(length(reason) <= 256)
);

CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT,
    details TEXT,
    start INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    end INTEGER
);

CREATE TABLE identities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subdomain TEXT NOT NULL CHECK(length(subdomain) <= 32),
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

    -- IDENTITY SETTINGS
    units TEXT CHECK (units IN ('mmol/L', 'mg/dL')), -- mg/dL, mmol/L

    -- DATE
    timezone TEXT CHECK(length(timezone) <= 64),
    language TEXT CHECK(length(language) <= 12),
    date_format TEXT CHECK(date_format IN ('YYYY-MM-DD', 'DD/MM/YYYY', 'MM/DD/YYYY')),
    time_format TEXT CHECK(time_format IN ('HH:mm:ss', 'hh:mm:ss a')),
    week_start_day TEXT CHECK (week_start_day IN ('monday', 'sunday')),

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
    alarm_urgent_low_enabled INTEGER DEFAULT 1,
    alarm_urgent_low INTEGER DEFAULT 56,

    alarm_low_enabled INTEGER DEFAULT 1,
    alarm_low INTEGER DEFAULT 72,
    alarm_low_minutes INTEGER DEFAULT 5,
    alarm_low_prediction_enabled INTEGER DEFAULT 1,

    alarm_high_enabled INTEGER DEFAULT 1,
    alarm_high INTEGER DEFAULT 180,
    alarm_high_minutes INTEGER DEFAULT 30,
    
    alarm_urgent_high_enabled INTEGER DEFAULT 1,
    alarm_urgent_high INTEGER DEFAULT 270,

    alarm_fast_rise_enabled INTEGER DEFAULT 1,
    alarm_fast_rise INTEGER DEFAULT 18,
    alarm_fast_rise_minutes INTEGER DEFAULT 15,

    alarm_fast_drop_enabled INTEGER DEFAULT 1,
    alarm_fast_drop INTEGER DEFAULT 18,
    alarm_fast_drop_minutes INTEGER DEFAULT 15,

    alarm_no_data_enabled INTEGER DEFAULT 1,
    alarm_no_data_minutes INTEGER DEFAULT 20,

    alarm_battery_low_enabled INTEGER DEFAULT 1,
    alarm_battery_low_percent INTEGER DEFAULT 20,
    alarm_battery_low_pwd_enabled INTEGER DEFAULT 1,
    alarm_battery_low_pwd_percent INTEGER DEFAULT 20,

    alarm_low_reservoir_enabled INTEGER DEFAULT 1,
    alarm_low_reservoir_units  INTEGER DEFAULT 25,

    -- NOTIFICATIONS
    notifications_enabled INTEGER DEFAULT 0,

    -- Constraints
    UNIQUE(subdomain, name)
);

CREATE TABLE job_stats_daily (
    day_queue_job TEXT NOT NULL, -- YYYY-MM-DD::queue::job

    total_jobs INTEGER NOT NULL DEFAULT 0,
    total_success INTEGER NOT NULL DEFAULT 0,
    total_failed INTEGER NOT NULL DEFAULT 0,

    duration_sum REAL NOT NULL DEFAULT 0,
    duration_squared_sum REAL NOT NULL DEFAULT 0,

    errors TEXT  NOT  NULL  DEFAULT  "[]" CHECK(length(errors) <= 1024),

    -- liteq jobs use epoch in seconds
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch() AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch() AS INTEGER)),

    PRIMARY KEY (day_queue_job)
);

CREATE TABLE jobs (
    id INTEGER NOT NULL,
    queue TEXT NOT NULL CHECK(length(queue) <= 64),
    job TEXT NOT NULL CHECK(length(job) <= 256),
    job_status TEXT NOT NULL DEFAULT 'queued' CHECK(length(job_status) <= 32),
    execute_after INTEGER NOT NULL DEFAULT 0,
    remaining_attempts INTEGER NOT NULL DEFAULT 1,
    consumer_fetched_at INTEGER NOT NULL DEFAULT 0,
    finished_at INTEGER NOT NULL DEFAULT 0,
    deduping_key TEXT NOT NULL DEFAULT '' CHECK(length(deduping_key) <= 256),
    errors TEXT NOT NULL DEFAULT "[]" CHECK(length(errors) <= 1024),
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE site_sync_sources (
    -- Nightscout/Dexcom
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subdomain TEXT NOT NULL CHECK(length(subdomain) <= 32), -- only for nightscout, dexcom can leave null
    type TEXT NOT NULL CHECK(type IN ('NIGHTSCOUT', 'DEXCOM')),

    url_or_username TEXT NOT NULL CHECK(length(url_or_username) <= 256), -- url for nightscout, username for dexcom
    password TEXT NOT NULL CHECK(length(password) <= 256), -- or ns token
    dexcom_region TEXT NOT NULL CHECK(dexcom_region IN ('USA', 'OTHER')) DEFAULT 'OTHER', -- USA or OTHER

    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    enabled INTEGER DEFAULT 1, -- 0 or 1 for now, can be used to disable without deleting

    unique_hash TEXT NOT NULL CHECK(length(unique_hash) <= 512), 

    -- Unique contraint of subdomain and type
    UNIQUE(subdomain, type)
);

CREATE TABLE sites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subdomain TEXT UNIQUE NOT NULL CHECK(length(subdomain) <= 32),
    batch TEXT GENERATED ALWAYS AS (SUBSTR(subdomain, 1, 1)) VIRTUAL NOT NULL,
    
    backup_key_hash TEXT NOT NULL CHECK(length(backup_key_hash) <= 64), 
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    

    max_db_size_kb_limit INTEGER DEFAULT 3000, -- default 3Mb, can be changed in control plane
    

    -- last_seen_at is default null because on creation they have not been seen, will be deleted quickly is not logged in sooner
    last_seen_at INTEGER,
    last_write_at INTEGER,
    deleted_at INTEGER,

    -- SITE SETTINGS
    units TEXT CHECK (units IN ('mmol/L', 'mg/dL')) DEFAULT 'mmol/L', -- mg/dL, mmol/L

    -- DATE
    timezone TEXT CHECK(length(timezone) <= 64) DEFAULT 'Pacific/Auckland',
    language TEXT CHECK(length(language) <= 12) DEFAULT 'en-NZ',
    date_format TEXT CHECK(date_format IN ('YYYY-MM-DD', 'DD/MM/YYYY', 'MM/DD/YYYY')) DEFAULT 'YYYY-MM-DD',
    time_format TEXT CHECK(time_format IN ('HH:mm:ss', 'hh:mm:ss a')) DEFAULT 'HH:mm:ss',
    week_start_day TEXT CHECK (week_start_day IN ('monday', 'sunday')) DEFAULT 'monday',

    -- INSULIN
    bolus_dia_hours REAL DEFAULT 6,
    bolus_onset_minutes REAL DEFAULT 20,
    bolus_peak_minutes REAL DEFAULT 60,

    basal_dia_hours REAL DEFAULT 24,
    basal_onset_minutes REAL DEFAULT 60,
    basal_peak_minutes REAL DEFAULT 120,

    default_insulin_sensitivity REAL DEFAULT 10,
    default_carb_ratio REAL DEFAULT 10,
    default_basal_rate REAL DEFAULT 1,

    -- CARBS
    carb_max_absorption_hours REAL DEFAULT 6,
    carb_min_5m_impact REAL DEFAULT 10
);

CREATE TABLE sqlite_sequence(name,seq);

CREATE TABLE sync_source_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    -- We do not store the actual sync data here, that is on the user, but we need to store the results
    unique_hash TEXT UNIQUE NOT NULL CHECK(length(unique_hash) <= 512), -- <type>_<username>_<password>_<dexcom_region> for dexcom, <type>_<url>_<token> for nightscout

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
    previous_sync_errors TEXT CHECK(length(previous_sync_errors) <= 1024), -- stored to show to user as feedback

    -- Timing EMAs to then determine frequency on
    total_ema REAL,
    success_ema REAL, -- https://en.wikipedia.org/wiki/Exponential_smoothing
    failure_ema REAL,
    
    -- Duration of last sync in ms
    last_sync_duration INTEGER -- 1 for success, 0 for failure, null for never tried, used for quick checks without having to look at timestamps 
);

