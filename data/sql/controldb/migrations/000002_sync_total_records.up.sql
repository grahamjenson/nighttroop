-- The records a sync source's syncs and backfills have added in all, shown on its sync report
ALTER TABLE site_sync_sources ADD COLUMN total_records INTEGER NOT NULL DEFAULT 0;
