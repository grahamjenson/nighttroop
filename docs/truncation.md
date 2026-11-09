# Truncating old data: a proposal

Status: **proposal, for comment**. Nothing here is built yet beyond what "Today" says. The numbers (ages, resolutions,
sizes) are starting points to argue about, not decisions.

## The problem

A site keeps everything at the resolution it was sent, forever, in a SQLite DB with a default limit of **20 MB**
(`sites.max_db_size_kb_limit`). A looping site writes a lot every day:

| Per day, a typical looping site | Rows | Rough size with indexes |
| --- | --- | --- |
| CGM readings, one source (every 5 min) | 288 | ~90 KB |
| CGM readings, a second source (e.g. Dexcom and Nightscout both connected) | +288 | +~90 KB |
| Device reports (every 5 min, before thinning) | 288 | ~150 KB, much of it loop status JSON |
| Temp basals and SMBs (AAPS / Trio) | 100–300 | ~60 KB |
| Carbs, manual boluses, notes, site changes | 5–20 | ~5 KB |

These sizes are estimates and should be measured with SQLite's `dbstat` on a real site before we pick numbers. Even so,
a looping site keeps ~155 KB a day that is never cut back (readings, doses, basals and the rest), plus device reports
(thinned and deleted after 30 days, ~2 MB at any time). That fills 20 MB in **about 4 months**, or about 2½ with a
second CGM source. Today only device reports are cut back (see below). Everything else grows until a control admin
raises the limit or the site is truncated by hand.

## What we want

- **Keep data forever, but not at full resolution forever.** A reading from last Tuesday matters to the minute; one from
  three years ago matters as part of that month's AGP and time in range.
- **Normalise, then thin, then summarise, then downsample the summary.** Each step loses less than the next, and each
  only runs on data old enough that nothing needs the finer resolution any more.
- **Never delete raw rows a summary has not been built and checked from.** A day's raw data is only removed once its
  summary is kept and marked as final ("frozen").
- **Be clear about what is lost.** Each record type below says what is deleted, at what age, what it becomes, and
  what can never be got back.

## Who reads what, and how far back

Old data does not need to stay in its raw form for these readers:

- **Loops and followers** (AAPS, Trio, Loop, xDrip+, LoopFollow, Nightguard) read the last few hours to a couple of
  days through the Nightscout API. v3 `history` syncs since a `lastModified` the app last saw, usually minutes or
  hours ago.
- **The site's pages** show the last 6–48 hours on the summary, and up to a year on the reports. The reports already
  read kept summaries (`summary_period`), not raw rows.
- **People looking back.** Clinic visits and "how was last year" need daily or weekly numbers and AGPs, not every
  reading.

So raw rows matter for days to weeks. Summaries matter for years.

## Budget at 20 MB

With the tiers and record types below, a looping site with one CGM source would roughly hold the following. These are
the same estimates as above, to be measured.

| Tier | Days | Per day | Total |
| --- | --- | --- | --- |
| Full, 0–14 days | 14 | ~250 KB (device reports at full resolution) | ~3.5 MB |
| Reduced, 14–90 days | 76 | ~70 KB (one reading per glucose, reduced columns, hourly roll-ups, device reports thinned then gone) | ~5.3 MB |
| Summarised, 90 days – 2 years | 640 | ~12 KB (the day's summary ~7 KB, plus kept events ~5 KB) | ~7.7 MB |
| **At 2 years** | | | **~16.5 MB** |
| Archive, each year after | 365 | ~6 KB (the day's summary ~1 KB, plus kept events ~5 KB) | **+~2.2 MB a year** |

So a 20 MB site lasts about 2 years before needing the archive tier, and then a couple more. After the first 2 years
the growth is mostly the **kept events** (meals, manual boluses and their treatment groups, ~5 KB a day), not the
summaries. Keeping data "forever" at 20 MB therefore also needs either the events to be thinner after some years
(e.g. their treatment groups dropped, notes kept), or the day summaries to give way to weeks and months (see Open
questions).

## Tiers

| Tier | Age (proposed default) | What is kept |
| --- | --- | --- |
| **Full** | 0–14 days | Everything as sent, as today (device reports already thinned, see below) |
| **Reduced** | 14–90 days | Rows kept, but duplicates removed, columns nobody reads cleared, high-volume automatic records rolled up by hour |
| **Summarised** | 90 days – 2 years | The day's summary, plus the few, meaningful events (meals, manual boluses, finger pricks, site/sensor changes, profile switches) kept as rows |
| **Archive** | 2 years + | A downsampled day summary (15 min glucose, hourly insulin, carbs and IOB/COB), weeks and months with their AGP, and the meaningful events |

Each age is a setting (env var, documented in `docs/config.md`), and a site over its size limit can be compacted
sooner (see "Size pressure").

## Per record type

Each section says what it is, how it is handled **today**, and the **proposal**. Each ends with **Lost**: what cannot
be recovered once the step has run.

### Glucose readings (`glucose_readings`)

The biggest table by count. A CGM reading every 5 minutes (every 1 minute for some Libre setups). Twice that when two
sources send the same glucose, e.g. a Dexcom connector and Nightscout.

**Today:** kept forever, every source. `utils/glucose.CleanReadings` already picks one reading per glucose when sources
overlap (a meter's first, then direct, then Nightscout, Dexcom, Libre, virtual). But it only does this when reading:
the duplicates stay in the DB.

**Proposal:**

1. **Normalise (after 2 days).** Run `CleanReadings` over each day and hard delete the readings it did not keep,
   "the others". A site then holds one reading per glucose. Two days gives connectors and backfills time to fill gaps
   (backfill re-fetches the last 2 days) before we pick the winners.
2. **Reduce (after 14 days).** Clear on each kept reading:
   - a CGM's raw signal (`noise`, `filtered`, `unfiltered`, `rssi`), which only xDrip+ calibration uses live;
   - `delta`, `direction` and `units`, as they can be worked out again from neighbouring readings and the stored mg/dL;
   - `created_by_app_name`, `created_by_device_name` and `updated_by_*`.

   `notes`, `reason`, the identity that wrote it and `reading_type` stay.
3. **Summarise (after 90 days).** CGM readings (`sgv`) for a day are deleted once the day's `summary_period` row is
   frozen. That row already holds the glucose in each 5 minutes from midnight (288 values, the newest clean reading in
   each, as whole mg/dL) and the day's time in range. Before we do this, the day row should also gain the statistics
   that can't be worked out from 5-minute values alone:
   - the count of readings;
   - min, max and mean;
   - standard deviation, from which GMI and CV are worked out;
   - time in each band: < 54, 54–70, 70–180, 180–250 and > 250 mg/dL;
   - which sources were used.

   All are computed from every clean reading, not the slots.
4. **Finger pricks (`mbg`) are kept as rows forever**, with the reduced columns. There are a few a day, and each is a
   real measurement someone took.
5. **Downsample (after 2 years).** The day row's 5-minute glucose becomes 15-minute (96 values, the mean of each 15
   minutes). Weeks and months keep their AGP.

**Lost:**

- After 2 days: the readings other sources sent of the same glucose, and which source sent each of those.
- After 14 days: CGM raw signal values, the app and device that sent each reading, delta and direction as sent (but
  they can be worked out again).
- After 90 days: each CGM reading's exact time and exact value. Each reading becomes one whole-mg/dL value per
  5 minutes, keeping the newest in each 5 minutes, so a 1-minute CGM loses 4 of every 5 readings. Notes on CGM
  readings are lost too (rare; they could be moved to a note event first, see Open questions).
- After 2 years: the 5-minute shape within each 15 minutes. Time in range and the other statistics stay exact, as they
  were computed from every reading.

### Insulin doses (`insulin_doses`)

Manual boluses are a few a day. Automatic doses (AAPS and Trio SMBs, Loop's automatic boluses) can be one every
5 minutes.

**Today:** kept forever.

**Proposal:**

1. **Reduce (after 14 days).** Clear `created_by_app_name`, `created_by_device_name`, `updated_by_*`, and keep
   only `automatic` and the combo bolus split of what only some apps send.
2. **Roll up automatic doses (after 14 days).** Automatic doses (`automatic`, or `dose_type` SMB) for
   a day are replaced by the day's **hourly insulin**: 24 values of automatic units given in each hour, on the day's
   summary row. The day's total daily dose is unchanged, as it is a sum.
3. **Manual boluses are kept as rows forever**, reduced. There are few, and each is a decision someone made.
4. **Downsample (after 2 years).** Nothing more: manual rows and hourly automatic insulin are already small.

**Lost:**

- After 14 days: each automatic dose's exact time and size, and its ID. An app that syncs history back past 14 days
  would not see them, though none we know of do.
- Kept: how much automatic insulin was given in each hour, and the daily total.

### Temp basals (`temp_basals`)

Every 5–30 minutes on a loop, each with a rate and a duration.

**Today:** kept forever.

**Proposal:**

1. **Roll up (after 14 days).** A day's temp basals become the day's **hourly basal delivered**: 24 values, the units
   delivered in each hour, from the temp basals and the scheduled basal in force. The summary already works this out
   per minute (`summaryMinutes`) for the total daily dose. Also kept: the minutes the pump was suspended
   (`type = PUMP_SUSPEND`) or on a zero temp, as a day total.
2. Delete the temp basal rows.

**Lost:** each temp basal's exact start, rate and duration, and its type other than suspended or zero. The basal the
loop chose minute by minute can't be redrawn, only each hour's total.

### Food intakes (`food_intakes`)

A few a day.

**Today:** kept forever.

**Proposal:** **keep as rows forever.** After 14 days, clear app, device and `updated_by_*`. Keep carbs, name, food type,
absorption time, notes, and the macros (`fat`, `protein`, `fibre`, `energy`), as they are small and what
people look back for. Add **hourly carbs** (24 values) to the day's summary row, so charts of old days don't need to
read rows.

**Lost:** which app and device entered each meal.

### Device reports (`device_reports`)

One every 5 minutes per device: phone battery, pump reservoir and battery, the loop's IOB and COB, and the loop's full
status (predictions, reason), which is most of their size.

**Today** (`data/sql/sitedb/queries/truncate.sql`, run every 10 minutes):
- the loop status and pump as sent are kept only on each device's newest report;
- reports are thinned to every 10 minutes after an hour, every 20 minutes after a week, and hourly after two weeks;
- every report is deleted after 30 days.

The day's summary already keeps the loop's IOB (total, basal, treatment) and COB every 5 minutes, from the reports.

**Proposal:**

1. Keep today's thinning and the 30-day delete.
2. Add to the day's summary row, before the reports go: the pump reservoir at the start and end of the day, and the
   lowest pump and phone battery in the day. Also the devices that reported (names, pump manufacturer and model), as
   a short list.
3. **Downsample (after 2 years).** The day row's IOB and COB become hourly (24 values each), and the basal and
   treatment split of IOB is dropped.

**Lost:**

- After 30 days: everything per report except the summary's IOB and COB every 5 minutes, and the day's reservoir,
  batteries and devices.
- After 2 years: IOB and COB within each hour, and the IOB split.

One catch: the summary is built from the reports, so a day summarised after its reports were thinned (more than
two weeks old) only has hourly IOB and COB. This is why summaries must be frozen and not rebuilt from thinned or
deleted data (see Freezing).

### Event groups (`event_groups`)

The Nightscout treatment that ties a dose, carbs, a reading and so on together, with `entered_by`, notes and the app's
own IDs (`uuid`, `sent_hash`) so the Nightscout API can rebuild the treatment as sent.

**Today:** kept forever. Its members can't be deleted while it refers to them (`ON DELETE RESTRICT`).

**Proposal:**
- A group whose members are rolled up or summarised (automatic doses, temp basals, CGM readings) is deleted first, in
  the same transaction.
- A group whose members are kept (meals, manual boluses, site changes) is kept. After 90 days the app's IDs and hashes
  (`uuid`, `sent_hash`) are cleared, keeping `group_type`, `entered_by`, notes and duration.

**Lost:**
- After 90 days: the app's own IDs, so a re-upload of a 90-day-old treatment would not be recognised as a duplicate.
  No app we know re-uploads that far back.
- The Nightscout shape of rolled-up treatments.

### Temp targets, profile switches, activities, notifications, component events

These are few a day, and what people look back for: what target or profile was in force, exercise, site and sensor
changes, and announcements.

**Today:** kept forever.

**Proposal:**
- **Temp targets, activities and component events** (site, sensor, pump and battery changes): **keep as rows
  forever**. Clear app, device and `updated_by_*` after 14 days. Component events are how sensor and site lifetimes
  are worked out over years.
- **Profile switches:** keep, but AAPS sends an "effective profile switch" whenever anything changes, each repeating
  the whole profile's JSON (up to 5 × 2 KB). After 14 days:
  - delete a switch whose profile (basal rates, ratios, sensitivities and targets, after its multiplier and timeshift)
    is the same as the switch in force before it;
  - or store each distinct profile once by its hash, as `loop_settings` does (`docs/loop_settings.md`), with the
    switch referring to it.
- **Notification events** (announcements, Loop remote action errors): delete after 1 year, unless a person wrote it
  (`created_by_identity_name` is a user, not a device).

**Lost:**
- Repeated switches to the same profile, and when each was sent.
- Announcements a device sent, after a year.

### Loop settings (`loop_settings`)

Each distinct loop configuration, once, by hash.

**Today:** kept forever. A P3 in `TODOs.md` says to clear the sets no record refers to when a site is
truncated.

**Proposal:** after device reports and profile presets referring to a settings row are deleted, delete the row,
unless it is the newest for its app. Keep the newest for each app forever.

**Lost:** old loop configurations nothing refers to.

### Presets (`food_presets`, `profile_presets`)

User-managed lists, not a time series.

**Proposal:** **not truncated.** Loop uploads its profile store often, but each store is one row per profile name and
store (`UNIQUE (name, profile_store_id)`). If that grows, keep each name's newest store and the store any kept profile
switch was made from.

### Soft deleted rows (any table)

A delete from the site or the Nightscout API sets `deleted_at` rather than removing the row. That way apps syncing v3
history see `isValid: false` and remove their copy.

**Proposal:** **hard delete rows soft deleted more than 30 days ago.** That is far longer than any app's sync interval.

**Lost:** that the row existed. An app that has not synced in 30 days would keep its copy of a deleted record.

### Summaries (`summary_period`)

The day, week, fortnight and month rows the reports read. Today they are built for up to a year back, and rebuilt on a
timezone or week-start change and on Regenerate.

A day row is about 6–7 KB today: 288 glucose values, plus 4 IOB/COB arrays of 288. That is ~2.5 MB a year, an eighth
of a 20 MB site each year. So summaries need their own downsampling to keep data "forever":

| Age | Day row | Week, fortnight, month rows |
| --- | --- | --- |
| 0–2 years | glucose and IOB/COB every 5 min, hourly insulin, basal and carbs, the statistics above (~7 KB) | kept, with AGP |
| 2 years + | glucose every 15 min, IOB and COB hourly, hourly insulin, basal and carbs, the statistics (~1 KB) | kept, with AGP |

That is ~0.4 MB a year after two years. If that is still too much over a decade or more, the next step would be to drop
day rows after N years and keep weeks and months with their AGPs and statistics (~0.1 MB a year). See Open questions.

## Freezing summaries

Once a day's raw rows start to go, its summary can't be built again from them. So:

- `summary_period` gains a `frozen_at`.
- A day is frozen when it first passes into the **Reduced** tier, after a final build from full-resolution data.
  Weeks, fortnights and months are frozen once all their days are.
- **Regenerate summaries** and the timezone and week-start rebuilds skip frozen rows rather than overwriting them with
  emptier ones. A frozen day keeps the timezone it was built in (already stored on the row). A site that changes
  timezone keeps its old days as they were.
- A day's raw rows are only deleted once its summary is frozen. Tests check that its total daily dose, carbs and time
  in range are the same before and after compaction.

## How it runs

- A **nightly compaction job** runs per site, after the daily summary builder has built yesterday, in the site's own
  timezone. It works one day at a time, oldest first, each day in one transaction, so it can stop and resume anywhere.
- Steps, in order: freeze, normalise glucose, reduce columns, roll up, summarise and delete, purge soft deletes,
  downsample old summaries, delete unreferenced loop settings.
- Each run records a **`site_compacted` event** with how many rows of each table it deleted or changed, which the
  site's admins and control see in Events.
- The existing 10-minute device report truncation stays as it is.
- **Size pressure:** a site over its limit is compacted with shorter ages (e.g. half of each), down to a floor (never
  less than 2 days of full data). Control's **Truncate** button runs the same compaction now, with those shorter ages,
  and vacuums the DB.
- **Download Data** gets the compacted data, as the DB is all there is. A later PR could add each day's summary as a
  CSV beside the DB in the zip, which is why it is a zip.

## What a person sees

- Data pages (glucose, insulin and so on) show rows only as far back as they are kept. A note under the table says
  older data is summarised in Reports.
- Reports (AGP, TDD, Summaries) are unchanged: they already read summaries.
- A day older than the **Summarised** age has no individual readings to hover in its day report, only its 5-minute (or
  15-minute) line.

## Open questions

1. **The ages.** Are 2 days (normalise), 14 days (reduce and roll up), 90 days (summarise) and 2 years (downsample)
   right? Should a paid or larger site get longer full-resolution windows?
2. **Lasting at 20 MB.** By the budget above a site reaches ~16.5 MB at 2 years and grows ~2.2 MB a year after, mostly
   from kept events. After N years, do we thin the events (drop their treatment groups, keep notes), drop day rows for
   weeks and months, or both? Or, with 20 MB, keep the Full tier longer (30 days costs ~4 MB more)?
3. **Raw archive to S3.** Should we write each month's raw rows to a compressed file in the site's S3 folder before
   deleting them, so nothing is truly lost? It costs storage and adds a restore path, but "summarised" would then mean
   "off the site DB" rather than "gone".
4. **Notes on summarised CGM readings.** Should a CGM reading with a note become a note event, so the note survives
   summarising?
5. **Profile switches:** delete repeats, or store each profile once by hash (more work, keeps every switch)?
6. **Notifications:** delete device ones after a year, or keep everything?
7. **Hourly vs 15-minute roll-ups** for automatic insulin and basal: hourly is ~10× smaller. Is it enough for looking
   back at how the loop behaved, or should the first few months keep 15 minutes?
8. **Which statistics** must the day row keep before raw readings go? Proposed: count, min, max, mean, SD, time in the
   five bands, and sources.
