---------------------------------------------------
-- SITE PATIENT DATA Tables
---------------------------------------------------

-----------
-- EVENTS (they happen at a time on a time line)
-----------

CREATE TABLE glucose_readings (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    
    -- UNIT IS ALWAYS MG/DL BECAUSE IT IS HIGHER RESOLUTION
    mgdl REAL NOT NULL CHECK(mgdl > 0 AND mgdl <= 1000),
    reading_type TEXT NOT NULL CHECK(reading_type IN ('sgv', 'mbg')) DEFAULT 'sgv', -- sgv, mbg etc
    units TEXT CHECK(units IN ('mmol/L', 'mg/dL')), -- mg/dL, mmol/L -- the reading is always store in mgdl, but the original unit CAN also be stored
    measurement_method TEXT CHECK(length(measurement_method) <= 32), -- cgm, meter, manual, etc

    delta REAL CHECK(delta BETWEEN -1000 AND 1000),
    direction TEXT CHECK(length(direction) <= 16),

    -- What only some apps send, NULL for most readings: a CGM's raw signal, from xDrip+ and its followers (noise,
    -- filtered, unfiltered, rssi; raw values can be in the hundreds of thousands), a noise or rssi of 0 meaning none
    -- and stored as NULL, though filtered and unfiltered are kept as sent, 0 too, as AAPS sends them,
    -- and an mbg AAPS calibrated its sensor with (is_calibration), and the sensor's reading then (sensor_mgdl_at_pairing)
    noise REAL CHECK(noise BETWEEN -1000 AND 1000),
    filtered REAL CHECK(filtered BETWEEN 0 AND 10000000),
    unfiltered REAL CHECK(unfiltered BETWEEN 0 AND 10000000),
    rssi REAL CHECK(rssi BETWEEN -1000 AND 1000),
    is_calibration INTEGER CHECK(is_calibration IN (0, 1)),
    sensor_mgdl_at_pairing REAL CHECK(sensor_mgdl_at_pairing > 0 AND sensor_mgdl_at_pairing <= 1000),

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0 CHECK(utc_offset_minutes BETWEEN -1440 AND 1440),
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)) CHECK(occurred_at BETWEEN 0 AND 253402300799999), -- before the year 10000, the last a date string can have
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_id INTEGER,
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_id INTEGER, -- the name can be reused after an identity is deleted, the id cannot
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
) STRICT;

CREATE VIEW glucose_readings_vw AS SELECT glucose_readings.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN glucose_readings ON TRUE;


CREATE TABLE insulin_doses (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    dose REAL NOT NULL CHECK(dose BETWEEN -500 AND 500), -- units, negative takes insulin back off IOB
    dose_type TEXT CHECK(length(dose_type) <= 32),
    

    -- What only some apps send, NULL for most doses: a combo bolus's split, the fraction given at once (split_now,
    -- Nightscout's splitNow / 100) and over time (split_ext, its splitExt / 100), whether a loop gave it
    -- (automatic, Loop's, as LoopFollow shows it as an SMB), and whether it is basal insulin (is_basal_insulin, AAPS's)
    split_now REAL CHECK(split_now BETWEEN 0 AND 1),
    split_ext REAL CHECK(split_ext BETWEEN 0 AND 1),
    automatic INTEGER CHECK(automatic IN (0, 1)),
    is_basal_insulin INTEGER CHECK(is_basal_insulin IN (0, 1)),

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0 CHECK(utc_offset_minutes BETWEEN -1440 AND 1440),
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)) CHECK(occurred_at BETWEEN 0 AND 253402300799999), -- before the year 10000, the last a date string can have
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_id INTEGER,
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_id INTEGER,
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
) STRICT;

CREATE VIEW insulin_doses_vw AS SELECT insulin_doses.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN insulin_doses ON TRUE;

CREATE TABLE food_intakes (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    name TEXT CHECK(length(name) <= 64), -- this could be a reference to a preset

    food_type TEXT CHECK(length(food_type) <= 32), -- Type of food

    absorption_time_minutes INTEGER CHECK(absorption_time_minutes BETWEEN 0 AND 1440),

    carbs REAL CHECK(carbs BETWEEN -1000 AND 1000), -- grams, negative takes back too much

    -- What only some apps send, NULL for most food: Nightscout's food fields, its category and subcategory, the grams
    -- of fat, protein and fibre, the energy (kcal or kJ), the portions, and the minutes bolused before eating (pre_bolus)
    category TEXT CHECK(length(category) <= 64),
    subcategory TEXT CHECK(length(subcategory) <= 64),
    fat REAL CHECK(fat BETWEEN -1000 AND 1000),
    protein REAL CHECK(protein BETWEEN -1000 AND 1000),
    fibre REAL CHECK(fibre BETWEEN -1000 AND 1000),
    energy REAL CHECK(energy BETWEEN -20000 AND 20000),
    portions REAL CHECK(portions BETWEEN -1000 AND 1000),
    pre_bolus REAL CHECK(pre_bolus BETWEEN -1440 AND 1440),

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0 CHECK(utc_offset_minutes BETWEEN -1440 AND 1440),
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)) CHECK(occurred_at BETWEEN 0 AND 253402300799999), -- before the year 10000, the last a date string can have
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_id INTEGER,
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_id INTEGER,
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
) STRICT;

CREATE VIEW food_intakes_vw AS SELECT food_intakes.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN food_intakes ON TRUE;

CREATE TABLE activities (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    name TEXT CHECK(length(name) <= 64),
    activity_type TEXT NOT NULL CHECK(length(activity_type) <= 32) default 'exercise',

    duration_minutes INTEGER CHECK(duration_minutes BETWEEN 0 AND 10080), -- at most a week

    -- What only some apps send, NULL for most activities: what a fitness tracker or watch measured of it (calories,
    -- steps, heart_rate), such as xDrip+'s, a heart rate of 0 meaning none and stored as NULL
    calories INTEGER CHECK(calories BETWEEN 0 AND 50000),
    steps INTEGER CHECK(steps BETWEEN 0 AND 1000000),
    heart_rate INTEGER CHECK(heart_rate BETWEEN 0 AND 300),

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0 CHECK(utc_offset_minutes BETWEEN -1440 AND 1440),
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)) CHECK(occurred_at BETWEEN 0 AND 253402300799999), -- before the year 10000, the last a date string can have
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_id INTEGER,
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_id INTEGER,
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
) STRICT;

CREATE VIEW activities_vw AS SELECT activities.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN activities ON TRUE;

CREATE TABLE notification_events (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    notification_type TEXT CHECK(length(notification_type) <= 32) NOT NULL DEFAULT 'announcement', -- note, warning, error, critical, info
    title TEXT CHECK(length(title) <= 128) NOT NULL,
    severity INTEGER CHECK(severity >= 0 AND severity <= 5) NOT NULL DEFAULT 0, -- 0 is lowest, 5 is highest
    content TEXT CHECK(length(content) <= 1024),

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0 CHECK(utc_offset_minutes BETWEEN -1440 AND 1440),
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)) CHECK(occurred_at BETWEEN 0 AND 253402300799999), -- before the year 10000, the last a date string can have
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_id INTEGER,
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_id INTEGER,
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
) STRICT;

CREATE VIEW notification_events_vw AS SELECT notification_events.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN notification_events ON TRUE;

-- temp_basals, temp_targets and profile_switches are all ranges over a time period so start is occured_at until duration (unless another overide happens)
-- activity also has a duration
CREATE TABLE temp_basals (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    duration_minutes INTEGER NOT NULL CHECK(duration_minutes BETWEEN 0 AND 10080), -- at most a week
    rate REAL NOT NULL CHECK(rate BETWEEN 0 AND 100), -- the units an hour given while it runs
    multiplier REAL CHECK(multiplier BETWEEN 0 AND 11), -- times the scheduled basal at its start, so 0.5 is half and 0 a zero temp; optional, but rate is not
    type TEXT CHECK(length(type) <= 64), -- AAPS's kind of temp basal, e.g. NORMAL, or PUMP_SUSPEND for a suspended pump
    
    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0 CHECK(utc_offset_minutes BETWEEN -1440 AND 1440),
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)) CHECK(occurred_at BETWEEN 0 AND 253402300799999), -- before the year 10000, the last a date string can have
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_id INTEGER,
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_id INTEGER,
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
) STRICT;

CREATE VIEW temp_basals_vw AS SELECT temp_basals.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN temp_basals ON TRUE;

CREATE TABLE temp_targets (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    duration_minutes INTEGER NOT NULL CHECK(duration_minutes BETWEEN 0 AND 10080), -- at most a week, 0 cancels
    high_target_mgdl REAL NOT NULL CHECK(high_target_mgdl BETWEEN 0 AND 1000), -- a cancel can have no targets
    low_target_mgdl REAL NOT NULL CHECK(low_target_mgdl BETWEEN 0 AND 1000),
    insulin_needs_scale_factor REAL CHECK(insulin_needs_scale_factor BETWEEN 0 AND 10),

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0 CHECK(utc_offset_minutes BETWEEN -1440 AND 1440),
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)) CHECK(occurred_at BETWEEN 0 AND 253402300799999), -- before the year 10000, the last a date string can have
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_id INTEGER,
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_id INTEGER,
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
) STRICT;

CREATE VIEW temp_targets_vw AS SELECT temp_targets.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN temp_targets ON TRUE;

-- Profile Switches are tricksy, they have a multiplier, but the values stored in here already contain that multiplier
CREATE TABLE profile_switches (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    name TEXT NOT NULL CHECK(length(name) <= 64),
    
    -- Profile Switch Specific
    duration_minutes INTEGER NOT NULL default 0 CHECK(duration_minutes BETWEEN 0 AND 525600), -- at most a year, and 0 means it is forever unless another profile overrides it
    multiplier REAL default 1 CHECK(multiplier BETWEEN 0 AND 10), -- times the profile's insulin, Nightscout's percentage / 100
    timeshift INTEGER default 0 CHECK(timeshift BETWEEN -1440 AND 1440), -- in minutes, e.g. 60 means the profile is shifted forward by 1 hour

    -- What only AAPS sends, NULL for most switches: the AAPS switch it was made from, which its followers read an
    -- effective profile switch (a Note) by, its name and customized name, its timeshift in milliseconds within a day,
    -- its percentage, how long it was to last in milliseconds, at most a year, and when it ended
    original_profile_name TEXT CHECK(length(original_profile_name) <= 64),
    original_customized_name TEXT CHECK(length(original_customized_name) <= 64),
    original_timeshift_ms INTEGER CHECK(original_timeshift_ms BETWEEN -86400000 AND 86400000),
    original_percentage INTEGER CHECK(original_percentage BETWEEN 0 AND 1000),
    original_duration_ms INTEGER CHECK(original_duration_ms BETWEEN 0 AND 31536000000),
    original_end INTEGER CHECK(original_end BETWEEN 0 AND 253402300799999),

    -- Profile Stuff
    dia REAL CHECK(dia BETWEEN 0 AND 24), -- in hours
    units TEXT CHECK(units IN ('mmol/L', 'mg/dL')), -- mg/dL, mmol/L used for targets
    timezone TEXT CHECK(length(timezone) <= 32),
    carbs_hr REAL CHECK(carbs_hr BETWEEN 0 AND 500),
    delay REAL CHECK(delay BETWEEN 0 AND 1440),

    carb_ratios_json TEXT CHECK(length(carb_ratios_json) <= 2048                     AND json_valid(carb_ratios_json)), -- 
    insulin_sensitivities_json TEXT CHECK(length(insulin_sensitivities_json) <= 2048 AND json_valid(insulin_sensitivities_json)), -- { [1,2] },
    basal_rates_json TEXT CHECK(length(basal_rates_json) <= 2048                     AND json_valid(basal_rates_json)), -- {"value": "110", "time": "00:00"},
    targets_high_json TEXT CHECK(length(targets_high_json) <= 2048                   AND json_valid(targets_high_json)), -- { [high, low] }
    targets_low_json TEXT CHECK(length(targets_low_json) <= 2048                     AND json_valid(targets_low_json)), -- { [high, low] }
    
    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0 CHECK(utc_offset_minutes BETWEEN -1440 AND 1440),
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)) CHECK(occurred_at BETWEEN 0 AND 253402300799999), -- before the year 10000, the last a date string can have
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_id INTEGER,
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_id INTEGER,
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
) STRICT;

CREATE VIEW profile_switches_vw AS SELECT profile_switches.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN profile_switches ON TRUE;

-- EventGroup is when multiple events happen together, like a meal with insulin and glucose reading
CREATE TABLE event_groups (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    group_type TEXT NOT NULL CHECK(length(group_type) > 0 AND length(group_type) <= 32),

    glucose_reading_id TEXT UNIQUE REFERENCES glucose_readings(id) ON DELETE RESTRICT,
    insulin_dose_id TEXT UNIQUE REFERENCES insulin_doses(id) ON DELETE RESTRICT,
    food_intake_id TEXT UNIQUE REFERENCES food_intakes(id) ON DELETE RESTRICT,
    temp_basal_id TEXT UNIQUE REFERENCES temp_basals(id) ON DELETE RESTRICT,
    temp_target_id TEXT UNIQUE REFERENCES temp_targets(id) ON DELETE RESTRICT,
    profile_switch_id TEXT UNIQUE REFERENCES profile_switches(id) ON DELETE RESTRICT,

    activity_id TEXT UNIQUE REFERENCES activities(id) ON DELETE RESTRICT,
    notification_event_id TEXT UNIQUE REFERENCES notification_events(id) ON DELETE RESTRICT,

    -- who the app says entered it, as Nightscout keeps it, e.g. loop://iPhone or xdrip
    entered_by TEXT CHECK(length(entered_by) <= 64),
    -- What only some apps send, NULL for most treatments: the app's own id for it, given back as sent, e.g. xDrip+'s
    -- uuid, and AAPS's running mode (an OpenAPS Offline), e.g. DISCONNECTED_PUMP, whether the loop forced it
    -- (auto_forced) and why (reasons), how long it was to last in milliseconds, at most a year (original_duration_ms),
    -- Loop's "indefinite" for an override with no end (duration_type), and a hash of the JSON it was last sent as, to
    -- tell one sent again with changes from one sent the same (sent_hash), and the pump's ID, serial and type (pump_id,
    -- pump_serial, pump_type), as AAPS sends with every treatment, and a sensor change's sensor code and transmitter
    -- (sensor_code, transmitter_id), the insulin AAPS says a bolus or profile switch was given with (icfg_json), and
    -- AAPS's Bolus Wizard calculation, the JSON text it sends and reads back to restore the record (bolus_calculator_result_json),
    -- and Loop's id for a carb entry or dose, which it finds its own by (sync_identifier), the insulin it says a dose
    -- is of (insulin_type), when a carb entry was entered (user_entered_at) and a bolus's insulin not absorbed (unabsorbed)
    uuid TEXT CHECK(length(uuid) <= 64),
    mode TEXT CHECK(length(mode) <= 64),
    auto_forced INTEGER CHECK(auto_forced IN (0, 1)),
    reasons TEXT CHECK(length(reasons) <= 1024),
    original_duration_ms INTEGER CHECK(original_duration_ms BETWEEN 0 AND 31536000000),
    duration_type TEXT CHECK(length(duration_type) <= 16),
    sent_hash TEXT CHECK(length(sent_hash) <= 32),
    pump_id TEXT CHECK(length(pump_id) <= 32),
    pump_serial TEXT CHECK(length(pump_serial) <= 32),
    pump_type TEXT CHECK(length(pump_type) <= 64),
    sensor_code TEXT CHECK(length(sensor_code) <= 64),
    transmitter_id TEXT CHECK(length(transmitter_id) <= 64),
    icfg_json TEXT CHECK(length(icfg_json) <= 1024 AND json_valid(icfg_json)),
    bolus_calculator_result_json TEXT CHECK(length(bolus_calculator_result_json) <= 4096 AND json_valid(bolus_calculator_result_json)),
    sync_identifier TEXT CHECK(length(sync_identifier) <= 256),
    insulin_type TEXT CHECK(length(insulin_type) <= 64),
    user_entered_at INTEGER CHECK(user_entered_at BETWEEN 0 AND 253402300799999),
    unabsorbed REAL CHECK(unabsorbed BETWEEN 0 AND 500),

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    duration_minutes INTEGER CHECK(duration_minutes BETWEEN 0 AND 525600), -- at most a year; this is needed for notes, and is common enough
    -- the duration to the millisecond, as sent (durationInMilliseconds), at most a year, as AAPS ends a temp target early
    -- to the millisecond and reads it back before the minutes
    duration_ms INTEGER CHECK(duration_ms BETWEEN 0 AND 31536000000),

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0 CHECK(utc_offset_minutes BETWEEN -1440 AND 1440),
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)) CHECK(occurred_at BETWEEN 0 AND 253402300799999), -- before the year 10000, the last a date string can have
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_id INTEGER,
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_id INTEGER,
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
) STRICT;


------------
-- Reports Reports are not events, they are a snapshot of the state of a device at a point in time
------------
CREATE TABLE device_reports (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    
    -- uploader
    -- device is created_by_device_name
    device_battery_percent REAL CHECK(device_battery_percent BETWEEN 0 AND 100),
    is_device_charging INTEGER DEFAULT 0 CHECK(is_device_charging IN (0, 1)),

    -- PUMP
    pump_state TEXT CHECK(length(pump_state) <= 64),
    pump_reservoir REAL CHECK(pump_reservoir BETWEEN 0 AND 10000), -- units
    pump_battery_percent REAL CHECK(pump_battery_percent BETWEEN 0 AND 100),
    pump_charging_state TEXT CHECK(length(pump_charging_state) <= 32),

    -- Extracted Loop Settings that will be useful for reporting and analysis
    loop_iob REAL CHECK(loop_iob BETWEEN -1000 AND 1000),
    loop_iob_basal REAL CHECK(loop_iob_basal BETWEEN -1000 AND 1000),
    loop_iob_treatment REAL CHECK(loop_iob_treatment BETWEEN -1000 AND 1000),
    loop_cob REAL CHECK(loop_cob BETWEEN -1000 AND 2000),

    -- The loop_settings its loop's configuration is (AAPS's), by its ID. Not a foreign key: deleting the settings keeps
    -- the report, which is read back without them.
    configuration_id TEXT CHECK(length(configuration_id) = 64),

    -- The loop's status as sent, its openaps or loop object (which loop_name says), with its IOB, COB, reason and
    -- predicted glucoses, most of a looping site's DB. It is kept only on its device's newest report with one.
    loop_status_json TEXT CHECK(length(loop_status_json) <= 8192 AND json_valid(loop_status_json)),
    -- Which loop it is, whose status loop_status_json is
    loop_name TEXT CHECK(loop_name IN ('openaps', 'loop')),

    -- What only some apps send, NULL for most reports: the pump's time when it reported, its manufacturer, model and
    -- ID, whether it was suspended or bolusing, which followers show, and Loop's override in force
    pump_clock INTEGER CHECK(pump_clock BETWEEN 0 AND 253402300799999),
    pump_manufacturer TEXT CHECK(length(pump_manufacturer) <= 64),
    pump_model TEXT CHECK(length(pump_model) <= 64),
    pump_id TEXT CHECK(length(pump_id) <= 64),
    pump_suspended INTEGER CHECK(pump_suspended IN (0, 1)),
    pump_bolusing INTEGER CHECK(pump_bolusing IN (0, 1)),
    override_json TEXT CHECK(length(override_json) <= 2048 AND json_valid(override_json)),
    -- The pump and uploader as sent, for their fields without a column, like AAPS's extended. The pump's, the larger,
    -- is kept, like the loop's status, only on its device's newest report with one.
    pump_json TEXT CHECK(length(pump_json) <= 4096 AND json_valid(pump_json)),
    uploader_json TEXT CHECK(length(uploader_json) <= 1024 AND json_valid(uploader_json)),


    -- Default Stuff

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0 CHECK(utc_offset_minutes BETWEEN -1440 AND 1440),
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)) CHECK(occurred_at BETWEEN 0 AND 253402300799999), -- before the year 10000, the last a date string can have
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),


    updated_by_identity_id INTEGER,
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    created_by_identity_id INTEGER,
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
) STRICT;


-- loop_settings are each different set of a loop app's settings, once, however often they are sent: Loop's
-- (loopSettings) and Trio's (its remote settings) from their profiles, and AAPS's configuration from its device
-- statuses. id is the SHA-256 of the app and the settings, in hex, so the same settings are always the same row, and a
-- profile or device report refers to the row it was sent with by it. That reference is not a foreign key: a row can be
-- deleted and the records sent with it are kept, read back without them. See docs/loop_settings.md.
CREATE TABLE loop_settings (
    id TEXT PRIMARY KEY CHECK(length(id) = 64),
    app TEXT NOT NULL CHECK(app IN ('loop', 'trio', 'aaps')),
    source TEXT NOT NULL CHECK(source IN ('profile', 'devicestatus')),
    settings_json TEXT NOT NULL CHECK(length(settings_json) <= 8192 AND json_valid(settings_json)),
    first_seen_at INTEGER NOT NULL CHECK(first_seen_at BETWEEN 0 AND 253402300799999),
    last_seen_at INTEGER NOT NULL CHECK(last_seen_at BETWEEN 0 AND 253402300799999),
    created_by_device_name TEXT CHECK(length(created_by_device_name) <= 64),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
) STRICT;
CREATE INDEX loop_settings_app_last_seen_at_idx ON loop_settings(app, last_seen_at);

----------
-- PRESETS are user defined templates for the event types
----------

CREATE TABLE food_presets (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    name TEXT NOT NULL CHECK(length(name) <= 64),

    food_type TEXT CHECK(length(food_type) <= 32),     

    absorption_time_minutes INTEGER CHECK(absorption_time_minutes BETWEEN 0 AND 1440),

    carbs REAL CHECK(carbs BETWEEN -1000 AND 1000), -- grams, negative takes back too much

    -- What only some apps send, NULL for most foods: Nightscout's food fields, its category and subcategory, the grams
    -- of fat, protein and fibre, the energy (kcal or kJ), the portions, and the minutes bolused before eating (pre_bolus)
    category TEXT CHECK(length(category) <= 64),
    subcategory TEXT CHECK(length(subcategory) <= 64),
    fat REAL CHECK(fat BETWEEN -1000 AND 1000),
    protein REAL CHECK(protein BETWEEN -1000 AND 1000),
    fibre REAL CHECK(fibre BETWEEN -1000 AND 1000),
    energy REAL CHECK(energy BETWEEN -20000 AND 20000),
    portions REAL CHECK(portions BETWEEN -1000 AND 1000),
    pre_bolus REAL CHECK(pre_bolus BETWEEN -1440 AND 1440),

    -- Default Stuff

    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_id INTEGER,
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_id INTEGER,
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
) STRICT;

-- e.g. this is just a list of available profiles and user settings to choose from
CREATE TABLE profile_presets (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    name TEXT NOT NULL CHECK(length(name) <= 64),

    -- Store stuff
    profile_store_id TEXT NOT NULL,
    is_default INTEGER NOT NULL DEFAULT 0, -- 0 for false, 1 for true

    -- Profile Stuff
    dia REAL CHECK(dia BETWEEN 0 AND 24), -- in hours
    units TEXT CHECK(units IN ('mmol/L', 'mg/dL')), -- mg/dL, mmol/L
    timezone TEXT CHECK(length(timezone) <= 32), -- e.g. UTC, America/New_York, etc
    carbs_hr REAL CHECK(carbs_hr BETWEEN 0 AND 500),
    delay REAL CHECK(delay BETWEEN 0 AND 1440),

    carb_ratios_json TEXT CHECK(length(carb_ratios_json) <= 2048 AND json_valid(carb_ratios_json)), -- 
    insulin_sensitivities_json TEXT CHECK(length(insulin_sensitivities_json) <= 2048 AND json_valid(insulin_sensitivities_json)), -- { [1,2] },
    basal_rates_json TEXT CHECK(length(basal_rates_json) <= 2048 AND json_valid(basal_rates_json)), -- {"value": "110", "time": "00:00"},
    targets_high_json TEXT CHECK(length(targets_high_json) <= 2048 AND json_valid(targets_high_json)), -- { [high, low] }
    targets_low_json TEXT CHECK(length(targets_low_json) <= 2048 AND json_valid(targets_low_json)), -- { [high, low] }

    -- loop settings
    -- The loop_settings the profile was sent with: Loop's settings, and Trio's remote control settings
    loop_settings_id TEXT CHECK(length(loop_settings_id) = 64),
    remote_settings_id TEXT CHECK(length(remote_settings_id) = 64),
    -- Who the app said uploaded it (enteredBy), e.g. Loop, as tools pick a loop's profiles by it
    entered_by TEXT CHECK(length(entered_by) <= 64),

    -- Default Stuff

    utc_offset_minutes INTEGER DEFAULT 0 CHECK(utc_offset_minutes BETWEEN -1440 AND 1440),
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)) CHECK(occurred_at BETWEEN 0 AND 253402300799999), -- before the year 10000, the last a date string can have
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_id INTEGER,
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_id INTEGER,
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64),

    -- Constraints
    UNIQUE (name, profile_store_id)
) STRICT;



-- general_settings are the documents apps keep in the Nightscout settings collection under the keys Nighttroop keeps
-- for them, shared by every identity of the site: its key, the kind of key it is (AAPS's running config, aaps and
-- aaps-state, or a message between AAPS and AAPSClient, aaps_clientcontrol_...), the document as sent, and who first
-- sent it
CREATE TABLE general_settings (
    key TEXT PRIMARY KEY CHECK(length(key) <= 128),
    type TEXT NOT NULL CHECK(type IN ('aaps', 'aaps_clientcontrol')),
    value_json TEXT NOT NULL CHECK(length(value_json) <= 262144 AND json_valid(value_json)),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
) STRICT;

--
-- INDEXES
--

-- summary_period is each day, week, fortnight and month of the site's data, named by its first date and from midnight
-- in the site's timezone, weeks from its week start day and fortnights every other one from the first of 1970. A day is
-- built from its readings, insulin and food, and a week, fortnight or month from its days. days is how many days had data, and the daily dose and carbs are a day's totals, or the mean
-- of a week's or month's days with data, as is its time in range. Glucose is in mg/dL every 5 minutes from midnight,
-- 288 values (276 or 300 for a day the clocks change), null when there was no reading: for a week or month, the
-- mean of its days' at that time of day. A week or month also has its AGP: the 5th, 25th, 50th, 75th and 95th
-- percentiles of its days' glucose in the 15 minutes around each 5 minutes from midnight, 288 of them, null when none
-- had a reading then.
CREATE TABLE summary_period (
    period TEXT NOT NULL CHECK(period IN ('day', 'week', 'fortnight', 'month')),
    date TEXT NOT NULL CHECK(length(date) = 10),
    timezone TEXT NOT NULL CHECK(length(timezone) <= 64),
    start_at INTEGER NOT NULL,
    end_at INTEGER NOT NULL,

    days INTEGER NOT NULL,
    time_in_range REAL,
    total_daily_dose REAL NOT NULL,
    -- The basal of the daily dose, the basal in force through the day, and a period's the mean of its days', as its daily
    -- dose is; the rest of the dose is boluses
    total_daily_basal REAL NOT NULL DEFAULT 0 CHECK(total_daily_basal >= 0),
    total_daily_carbs REAL NOT NULL,
    glucose_mgdl_json TEXT NOT NULL CHECK(length(glucose_mgdl_json) <= 2048 AND json_valid(glucose_mgdl_json)),
    agp_mgdl_json TEXT CHECK(agp_mgdl_json IS NULL OR (length(agp_mgdl_json) <= 8192 AND json_valid(agp_mgdl_json))),
    -- A day's glucose readings, those from a meter, and its 5 minutes with a CGM reading, and a period's its days' added
    -- up. Its glucose mean and standard deviation, in mg/dL, are of its readings, and a period's of all its days'. The
    -- percent of its readings very low (below 54 mg/dL), low, high and very high (above 250 mg/dL), with time_in_range
    -- adding up to 100, are a period's the means of its days', as time_in_range is.
    glucose_readings INTEGER NOT NULL DEFAULT 0 CHECK(glucose_readings >= 0),
    meter_readings INTEGER NOT NULL DEFAULT 0 CHECK(meter_readings >= 0),
    cgm_slots INTEGER NOT NULL DEFAULT 0 CHECK(cgm_slots >= 0),
    glucose_mean_mgdl REAL,
    glucose_sd_mgdl REAL,
    time_very_low REAL,
    time_low REAL,
    time_high REAL,
    time_very_high REAL,
    -- The loop's IOB (total, basal and treatment) and COB every 5 minutes from midnight, as glucose is: a day's newest
    -- loop device report in each 5 minutes, null when there was none, and a week's or month's the mean of its days'
    loop_iob_json TEXT NOT NULL CHECK(length(loop_iob_json) <= 4096 AND json_valid(loop_iob_json)),
    loop_iob_basal_json TEXT NOT NULL CHECK(length(loop_iob_basal_json) <= 4096 AND json_valid(loop_iob_basal_json)),
    loop_iob_treatment_json TEXT NOT NULL CHECK(length(loop_iob_treatment_json) <= 4096 AND json_valid(loop_iob_treatment_json)),
    loop_cob_json TEXT NOT NULL CHECK(length(loop_cob_json) <= 4096 AND json_valid(loop_cob_json)),
    -- When it was marked to be built again, as a record in it was deleted or moved to another day, so it is built again
    -- after a restart too; null while it is up to date
    stale_at INTEGER,

    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    PRIMARY KEY (period, date)
) STRICT;
-- The summaries a time is in, as marking them stale looks them up, and those stale
CREATE INDEX summary_period_end_at_idx ON summary_period(end_at);
CREATE INDEX summary_period_stale_idx ON summary_period(period, start_at) WHERE stale_at IS NOT NULL;

-- Every All*/Search* query orders by occurred_at DESC with a LIMIT, so an
-- index on it lets SQLite walk rows newest first and stop instead of scanning
-- and sorting the whole table (the Nightscout API dedups every create this way).
CREATE INDEX glucose_readings_occurred_at_idx ON glucose_readings(occurred_at);
CREATE INDEX insulin_doses_occurred_at_idx ON insulin_doses(occurred_at);
CREATE INDEX food_intakes_occurred_at_idx ON food_intakes(occurred_at);
CREATE INDEX activities_occurred_at_idx ON activities(occurred_at);
CREATE INDEX notification_events_occurred_at_idx ON notification_events(occurred_at);
CREATE INDEX temp_basals_occurred_at_idx ON temp_basals(occurred_at);
CREATE INDEX temp_targets_occurred_at_idx ON temp_targets(occurred_at);
CREATE INDEX profile_switches_occurred_at_idx ON profile_switches(occurred_at);
CREATE INDEX event_groups_occurred_at_idx ON event_groups(occurred_at);
CREATE INDEX device_reports_occurred_at_idx ON device_reports(occurred_at);
CREATE INDEX profile_presets_occurred_at_idx ON profile_presets(occurred_at);
CREATE INDEX food_presets_created_at_idx ON food_presets(created_at);

-- LastModifiedCollections takes the newest updated_at of each collection
CREATE INDEX glucose_readings_updated_at_idx ON glucose_readings(updated_at);
-- The stale rebuild finds the summary data written since it last checked
CREATE INDEX insulin_doses_updated_at_idx ON insulin_doses(updated_at);
CREATE INDEX food_intakes_updated_at_idx ON food_intakes(updated_at);
CREATE INDEX temp_basals_updated_at_idx ON temp_basals(updated_at);
CREATE INDEX temp_targets_updated_at_idx ON temp_targets(updated_at);
CREATE INDEX profile_switches_updated_at_idx ON profile_switches(updated_at);
CREATE INDEX activities_updated_at_idx ON activities(updated_at);
CREATE INDEX event_groups_updated_at_idx ON event_groups(updated_at);
CREATE INDEX device_reports_updated_at_idx ON device_reports(updated_at);
CREATE INDEX food_presets_updated_at_idx ON food_presets(updated_at);
CREATE INDEX profile_presets_updated_at_idx ON profile_presets(updated_at);

-- Type lookups and deletes by type, a device's reports, and a profile store's presets
-- A type's readings by occurred_at let each alarm check find its newest CGM readings without scanning and sorting them all.
CREATE INDEX glucose_readings_reading_type_occurred_at_idx ON glucose_readings(reading_type, occurred_at);
-- So are a type's newest treatments, which a delete by type and each Nightscout properties request find (a sensor or
-- site change is a treatment), rather than sorting all of a type, or reading every treatment to find none of a type
CREATE INDEX event_groups_group_type_occurred_at_idx ON event_groups(group_type, occurred_at);
CREATE INDEX device_reports_device_occurred_at_idx ON device_reports(created_by_device_name, occurred_at);
CREATE INDEX profile_presets_profile_store_id_idx ON profile_presets(profile_store_id);

-- A settings search or history finds the general settings changed since
CREATE INDEX general_settings_updated_at_idx ON general_settings(updated_at);
