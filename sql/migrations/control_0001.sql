-- unique ID
-- subdomain is the unique identifier for the user and is used in the URL
-- pool is the first charater of the subdomain
-- backup key is provided to the user on creation, used as proof of ownership in case of lost information
-- only soft delete users


CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subdomain TEXT UNIQUE NOT NULL,
    batch TEXT GENERATED ALWAYS AS (SUBSTR(subdomain, 1, 1)) VIRTUAL,
    
    backup_key_hash TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    deleted_at TIMESTAMP
);

CREATE TABLE admins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    totp_secret TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    TYPE TEXT,
    details TEXT,
    start INTEGER,
    end INTEGER,

    timestamp INTEGER GENERATED ALWAYS AS (START) VIRTUAL,
    starttime TEXT GENERATED ALWAYS AS (datetime(START / 1000, 'unixepoch')) VIRTUAL,
    endtime TEXT GENERATED ALWAYS AS (datetime(END / 1000, 'unixepoch')) VIRTUAL
);

