-- table_sizes_json is each table of the site's DB by name, with the bytes it and its indexes take and its records, as
-- the resource check last measured them from SQLite's dbstat, so a page shows them without reading the whole DB
ALTER TABLE sites ADD COLUMN table_sizes_json TEXT CHECK(length(table_sizes_json) <= 16384 AND json_valid(table_sizes_json));
