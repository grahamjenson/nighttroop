# Truncating old data: a proposal

**Status:** proposal, for comment. Nothing is built beyond what "Today" says. The numbers (ages, resolutions, sizes)
are starting points to argue about, not decisions.

## The problem

A site keeps everything at the resolution sent, forever, in a SQLite DB with a default limit of **20 MB**
(`sites.max_db_size_kb_limit`). A looping site writes a lot every day:

| Per day, a typical looping site | Rows | Rough size with indexes |
| --- | --- | --- |
| CGM readings, one source (every 5 min) | 288 | ~90 KB |
| CGM readings, a second source (e.g. Dexcom and Nightscout both connected) | +288 | +~90 KB |
| Device reports (every 5 min, before thinning) | 288 | ~150 KB, much of it loop status JSON |
| Temp basals and SMBs (AAPS / Trio) | 100–300 | ~60 KB |
| Carbs, manual boluses, notes, site changes | 5–20 | ~5 KB |

- Estimates: measure them with SQLite's `dbstat` on a real site before we pick numbers.
- ~155 KB a day is never cut back (readings, doses, basals and the rest), plus device reports (thinned, deleted after
  30 days, ~2 MB at any time). Thus 20 MB fills in **about 4 months**, or about 2½ with a second CGM source.
- Today only device reports are cut back. Everything else grows until a control admin raises the limit, or the site
  is truncated by hand.

## Goals

- **Keep data forever, not at full resolution.** Last Tuesday's reading matters to the minute; one from three years
  ago only in that month's AGP and time in range.
- **Normalise, thin, summarise, then downsample the summary.** Each step loses less than the next, and runs only on
  data too old to need finer resolution.
- **Delete raw rows only after their summary is built, checked and frozen (final).**
- **Say what is lost:** what is deleted, at what age, what it becomes, and what is lost for good.

## Who reads what, and how far back

- **Loops and followers** (AAPS, Trio, Loop, xDrip+, LoopFollow, Nightguard): hours to a couple of days, through the
  Nightscout API. v3 `history` syncs since the app's last `lastModified`, usually minutes or hours ago.
- **The site's pages:** 6–48 hours on the summary, up to a year on the reports (from kept summaries,
  `summary_period`, not raw rows).
- **People looking back** (clinic visits, "how was last year"): daily or weekly numbers and AGPs.

Thus raw rows matter for days to weeks, summaries for years.

## Tiers

| Tier | Age (proposed default) | What is kept |
| --- | --- | --- |
| **Full** | 0–14 days | Everything as sent, as today (device reports already thinned, see below) |
| **Reduced** | 14–90 days | Rows kept, but duplicates removed, columns nobody reads cleared, high-volume automatic records rolled up by hour |
| **Summarised** | 90 days – 2 years | The day's summary, plus the few, meaningful events (meals, manual boluses, finger pricks, site/sensor changes, profile switches) kept as rows |
| **Archive** | 2 years + | A downsampled day summary (15 min glucose, hourly insulin, carbs and IOB/COB), weeks and months with their AGP, and the meaningful events |

Each age is a setting (env var, documented in `docs/config.md`). A site over its size limit can be compacted sooner
(see [Size pressure](#size-pressure)).

## Budget at 20 MB

A looping site with one CGM source (same estimates, to be measured):

| Tier | Days | Per day | Total |
| --- | --- | --- | --- |
| Full, 0–14 days | 14 | ~250 KB (device reports at full resolution) | ~3.5 MB |
| Reduced, 14–90 days | 76 | ~70 KB (one reading per glucose, reduced columns, hourly roll-ups, device reports thinned then gone) | ~5.3 MB |
| Summarised, 90 days – 2 years | 640 | ~12 KB (the day's summary ~7 KB, plus kept events ~5 KB) | ~7.7 MB |
| **At 2 years** | | | **~16.5 MB** |
| Archive, each year after | 365 | ~6 KB (the day's summary ~1 KB, plus kept events ~5 KB) | **+~2.2 MB a year** |

- A 20 MB site lasts about 2 years before the archive tier, then a couple more. Growth is then mostly **kept events**
  (meals, manual boluses and their treatment groups, ~5 KB a day), not summaries.
- Thus "forever" at 20 MB also needs thinner events after some years (example: drop treatment groups, keep notes), or
  day summaries that give way to weeks and months (see Open questions).

## Per record type

Each section gives **Today**, the **Proposal**, and **Lost**: what cannot be recovered after the step runs.

### Glucose readings (`glucose_readings`)

The biggest table by count: a CGM reading every 5 minutes (every 1 minute for some Libre setups), twice that when two
sources send the same glucose (example: a Dexcom connector and Nightscout).

**Today:** kept forever, every source. `utils/glucose.CleanReadings` picks one reading per glucose when sources
overlap (a meter's first, then direct, then Nightscout, Dexcom, Libre, virtual), but only on read. Duplicates stay in
the DB.

**Proposal:**

1. **Normalise (after 2 days).** Run `CleanReadings` over each day, and hard delete "the others". Two days lets
   connectors and backfills (which re-fetch the last 2 days) fill gaps first.
2. **Reduce (after 14 days).** Keep `notes`, `reason`, the identity that wrote it and `reading_type`. Clear:
   - the CGM raw signal (`noise`, `filtered`, `unfiltered`, `rssi`), which only live xDrip+ calibration uses;
   - `delta`, `direction` and `units`, which neighbouring readings and the stored mg/dL give again;
   - `created_by_app_name`, `created_by_device_name` and `updated_by_*`.
3. **Summarise (after 90 days).** Delete a day's CGM readings (`sgv`) once its `summary_period` row is frozen. That
   row already holds the day's time in range, and the glucose in each 5 minutes from midnight (288 values, the newest
   clean reading in each, whole mg/dL). First it gains, from every clean reading:
   - count, min, max and mean;
   - standard deviation (for GMI and CV);
   - time in < 54, 54–70, 70–180, 180–250 and > 250 mg/dL;
   - sources used.
4. **Keep finger pricks (`mbg`) as rows forever**, reduced: a few a day, each a real measurement.
5. **Downsample (after 2 years).** The day row's glucose becomes 15-minute (96 means). Weeks and months keep their AGP.

**Lost:**

- After 2 days: other sources' readings of the same glucose, and which source sent each.
- After 14 days: raw signal values, each reading's app and device, and delta and direction as sent (they can be
  worked out again).
- After 90 days: each reading's exact time and value (a 1-minute CGM loses 4 of every 5), and notes on CGM readings
  (rare; they could become note events first, see Open questions).
- After 2 years: the 5-minute shape in each 15 minutes. Time in range and the statistics stay exact.

### Insulin doses (`insulin_doses`)

Manual boluses: a few a day. Automatic doses (AAPS and Trio SMBs, Loop's automatic boluses): up to one every 5 minutes.
**Today:** kept forever.

**Proposal:**

1. **Reduce (after 14 days).** Clear `created_by_app_name`, `created_by_device_name` and `updated_by_*`. Of what only
   some apps send, keep only `automatic` and the combo bolus split.
2. **Roll up automatic doses (after 14 days).** A day's automatic doses (`automatic`, or `dose_type` SMB) become 24
   **hourly insulin** values on its summary row. The total daily dose, a sum, is unchanged.
3. **Keep manual boluses as rows forever**, reduced: few, and each a person's decision.
4. **Downsample (after 2 years):** nothing more; both are small.

**Lost:** after 14 days, each automatic dose's exact time, size and ID. An app that syncs history past 14 days would
not see them, but we know of none.

### Temp basals (`temp_basals`)

Every 5–30 minutes on a loop, each with a rate and duration. **Today:** kept forever.

**Proposal:**

1. **Roll up (after 14 days)** into 24 **hourly basal delivered** values, from the temp basals and the scheduled basal
   in force. The summary already works this out per minute (`summaryMinutes`) for the total daily dose. Also keep the
   day's minutes suspended (`type = PUMP_SUSPEND`) or on a zero temp.
2. Delete the temp basal rows.

**Lost:** each temp basal's start, rate, duration, and type other than suspended or zero. Only hourly totals remain.

### Food intakes (`food_intakes`)

A few a day. **Today:** kept forever. **Proposal:** keep as rows forever.

- After 14 days, clear app, device and `updated_by_*`.
- Keep carbs, name, food type, absorption time, notes and macros (`fat`, `protein`, `fibre`, `energy`): small, and
  what people look back for.
- Add 24 **hourly carbs** to the day's summary row, so charts of old days do not read rows.

**Lost:** which app and device entered each meal.

### Device reports (`device_reports`)

One every 5 minutes per device: phone battery, pump reservoir and battery, the loop's IOB and COB, and its full status
(predictions, reason), most of their size.

**Today** (`data/sql/sitedb/queries/truncate.sql`, every 10 minutes):

- The loop status and pump as sent stay only on each device's newest report.
- Reports thin to every 10 minutes after an hour, 20 after a week, and hourly after two weeks.
- All are deleted after 30 days.
- The day's summary keeps the loop's IOB (total, basal, treatment) and COB every 5 minutes.

**Proposal:**

1. Keep today's thinning and 30-day delete.
2. First add to the day's summary row: the pump reservoir at day start and end, the lowest pump and phone battery,
   and a short list of reporting devices (names, pump manufacturer and model).
3. **Downsample (after 2 years):** IOB and COB become hourly (24 values each), without the basal and treatment split.

**Lost:** after 30 days, all per report except the above. After 2 years, IOB and COB within each hour, and the split.

**Note:** A day summarised after its reports thinned (over two weeks old) has only hourly IOB and COB. Thus summaries
must be frozen, not rebuilt from thinned or deleted data (see [Freezing summaries](#freezing-summaries)).

### Event groups (`event_groups`)

The Nightscout treatment that ties a dose, carbs, a reading and so on together, with `entered_by`, notes and the app's
own IDs (`uuid`, `sent_hash`) to rebuild it as sent.

**Today:** kept forever. Its members cannot be deleted while it refers to them (`ON DELETE RESTRICT`).

**Proposal:**

- Members rolled up or summarised (automatic doses, temp basals, CGM readings): delete the group first, in the same
  transaction.
- Members kept (meals, manual boluses, site changes): keep the group. After 90 days clear `uuid` and `sent_hash`;
  keep `group_type`, `entered_by`, notes and duration.

**Lost:** after 90 days, the app's IDs, so a re-upload of a 90-day-old treatment is not a known duplicate (no app we
know re-uploads that far back). Also the Nightscout shape of rolled-up treatments.

### Temp targets, profile switches, activities, notifications

A few a day, and what people look back for: target or profile in force, exercise, site and sensor changes,
announcements. **Today:** kept forever.

**Proposal:**

- **Temp targets and activities:** rows forever; clear app, device and `updated_by_*` after 14 days. Site, sensor,
  pump and battery changes (event groups with no events) are kept as above: they give sensor and site lifetimes.
- **Profile switches:** keep. AAPS sends an "effective profile switch" on any change, each with the whole profile's
  JSON (up to 5 × 2 KB). After 14 days, either:
  - delete a switch whose profile (basal rates, ratios, sensitivities and targets, after multiplier and timeshift)
    equals the one in force before it; or
  - store each distinct profile once by hash, as `loop_settings` does (`docs/loop_settings.md`), for the switch to
    refer to.
- **Notification events** (announcements, Loop remote action errors): delete after 1 year, unless a person wrote it
  (`created_by_identity_name` is a user, not a device).

**Lost:** repeated switches to the same profile and their times; device announcements over a year old.

### Loop settings (`loop_settings`)

Each distinct loop configuration, once, by hash. **Today:** truncating a site (My Account's or control's Truncate)
deletes sets no device report or profile preset refers to, except each app's last seen, kept forever
(`docs/loop_settings.md`). The 10-minute job does not. **Proposal:** the nightly compaction deletes them too.
**Lost:** old configurations nothing refers to.

### Presets (`food_presets`, `profile_presets`)

User-managed lists, not a time series. **Proposal:** not truncated. Loop uploads its profile store often, but each
store is one row per profile name and store (`UNIQUE (name, profile_store_id)`). If that grows, keep each name's
newest store, and the store any kept profile switch was made from.

### Soft deleted rows (any table)

A delete from the site or the Nightscout API sets `deleted_at`, so apps that sync v3 history see `isValid: false` and
remove their copy. **Proposal:** hard delete rows soft deleted over 30 days ago, far longer than any app's sync
interval. **Lost:** that the row existed. An app not synced in 30 days keeps its copy.

### Summaries (`summary_period`)

The day, week, fortnight and month rows the reports read. Today they are built up to a year back, and rebuilt on a
timezone or week-start change and on Regenerate. A day row is ~6–7 KB (288 glucose values, plus 4 IOB/COB arrays of
288): ~2.5 MB a year, an eighth of a 20 MB site. So summaries need their own downsampling:

| Age | Day row | Week, fortnight, month rows |
| --- | --- | --- |
| 0–2 years | glucose and IOB/COB every 5 min, hourly insulin, basal and carbs, the statistics above (~7 KB) | kept, with AGP |
| 2 years + | glucose every 15 min, IOB and COB hourly, hourly insulin, basal and carbs, the statistics (~1 KB) | kept, with AGP |

That is ~0.4 MB a year after two years. If too much over a decade, drop day rows after N years and keep weeks and
months with their AGPs and statistics (~0.1 MB a year). See Open questions.

## Freezing summaries

A day's summary cannot be rebuilt once its raw rows start to go. So:

- `summary_period` gains a `frozen_at`.
- A day is frozen when it first enters the **Reduced** tier, after a final full-resolution build. Weeks, fortnights
  and months are frozen once all their days are.
- **Regenerate summaries** and timezone and week-start rebuilds skip frozen rows.
- A frozen day keeps the timezone it was built in (on the row), even if the site changes timezone.
- Raw rows are deleted only after the summary is frozen. Tests check the day's total daily dose, carbs and time in
  range are unchanged by compaction.

## How it runs

- A **nightly compaction job** runs per site, in the site's timezone, after the daily summary builder has built
  yesterday.
- It works one day at a time, oldest first, each day in one transaction, so it can stop and resume anywhere.
- Steps, in order: freeze, normalise glucose, reduce columns, roll up, summarise and delete, purge soft deletes,
  downsample old summaries, delete unreferenced loop settings.
- Each run records a **`site_compacted` event** (rows deleted or changed per table), shown in Events to the site's
  admins and control.
- The 10-minute device report truncation stays as it is.
- **Download Data** gets the compacted data, as the DB is all there is. A later PR could add each day's summary as a
  CSV beside the DB in the zip, which is why it is a zip.

### Size pressure

- A site over its limit is compacted with shorter ages (example: half of each), down to a floor: never less than 2
  days of full data.
- Control's **Truncate** button runs the same compaction now, with those shorter ages, and vacuums the DB.

## What a person sees

- Data pages (glucose, insulin and so on) show only kept rows, with a note under the table: older data is summarised
  in Reports.
- Reports (AGP, TDD, Summaries) are unchanged: they already read summaries.
- A day past the **Summarised** age has no readings to hover in its day report, only its 5-minute (or 15-minute) line.

## Open questions

1. **Ages:** are 2 days (normalise), 14 days (reduce and roll up), 90 days (summarise) and 2 years (downsample) right?
   Should a paid or larger site get longer full-resolution windows?
2. **Lasting at 20 MB:** ~16.5 MB at 2 years, then ~2.2 MB a year, mostly kept events. After N years, thin events
   (drop treatment groups, keep notes), drop day rows for weeks and months, or both? Or keep Full longer (30 days costs
   ~4 MB more)?
3. **Raw archive to S3:** write each month's raw rows to a compressed file in the site's S3 folder before deletion, so
   nothing is truly lost? It costs storage and a restore path, but "summarised" would mean "off the site DB", not
   "gone".
4. **Notes on summarised CGM readings:** should a CGM reading with a note become a note event, so the note survives?
5. **Profile switches:** delete repeats, or store each profile once by hash (more work, keeps every switch)?
6. **Notifications:** delete device ones after a year, or keep everything?
7. **Hourly vs 15-minute roll-ups** of automatic insulin and basal: hourly is ~10× smaller. Enough to see how the loop
   behaved, or keep 15 minutes for the first few months?
8. **Day row statistics** to keep before raw readings go? Proposed: count, min, max, mean, SD, time in the five bands,
   and sources.
