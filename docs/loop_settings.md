# Loop settings

A loop app's settings (examples: Loop's maximum bolus, Trio's remote control keys) change rarely, but come in on
records sent often:

| App  | Sent in                                     | Field                                          | How often                               |
|------|---------------------------------------------|------------------------------------------------|-----------------------------------------|
| Loop | a profile (`/api/v1/profile`)               | `loopSettings`                                 | when its therapy settings are saved     |
| Trio | a profile                                   | its remote settings, at the top of the profile | when a schedule is saved                |
| AAPS | each device status (`/api/v1/devicestatus`) | `configuration`                                | about hourly, and `{}` in the others    |

`loop_settings` keeps each different set once, and each record refers to its set. (They were copied onto every
`profile_presets` row and device report.)

**Note:** A device status's `openaps` and `loop` objects are not settings. They are the loop's state at that cycle:
IOB, COB, predictions, what it enacted, and the cycle's effective ISF, CR and targets.

They change every cycle, and stay in `device_reports.loop_status_json`. After a day, the truncate every 10 minutes
removes their predictions, and keeps the rest.

## The table

`loop_settings` has a row for each different set of an app's settings:

- `id`: SHA-256, in hex, of the app and settings. The same settings from the same app are always one row, whenever
  and however often sent.
- `app`: `loop`, `trio` or `aaps`.
- `source`: `profile` or `devicestatus`.
- `settings_json`: the settings as the app's model writes them. The model drops unknown fields and orders known ones,
  and Go sorts map keys, so the same settings are always the same text. A Go upgrade that wrote numbers differently
  could add a second row for the same settings, but never mix two sets.
- `first_seen_at`, `last_seen_at`: the earliest and latest times of the records they came in.
- `created_by_device_name`: the device that sent them first.

## Writing: from a Nightscout write

`recordLoopSettings` (`server/nightscoutapi/nscollections/loop_settings.go`) runs when a profile or device status is
created, in the same transaction:

1. It writes the settings through the app's model, and calculates their ID.
2. It upserts the row: a new set is added, a known set only widens when it was seen. Thus repeat uploads, retries,
   backfills and out of order records land on one row, and A, then B, then A again is two rows, not three.
3. The record keeps the ID: `profile_presets.loop_settings_id` and `remote_settings_id`, and
   `device_reports.configuration_id`.

Empty (`{}`, as in most AAPS statuses) or missing settings refer to no row. Settings over 8 KB are not kept, but the
record is.

## Reading: putting the records back together

- A Nightscout read gives each record as sent: a profile's `loopSettings` and Trio remote settings, and a device
  status's `configuration`.
- The converters (`DBToNightscoutProfileStore`, `DBToNightscoutDeviceStatus`) find the row by its ID: one primary key
  lookup for each record that has one.
- A profile sent with no Loop or Trio settings gets Loop's last seen set (`SearchLoopSettings`, limit 1), if that set
  is not newer than the profile. Thus LoopFollow can control Loop after AAPS or the careportal sends a profile.
- The reference is not a foreign key. A set deleted on the site's Loop Settings page keeps its records. They read back
  as if no settings were sent, and nothing else changes.

## The site

The Data hub's Presets group has a Loop Settings page (`/presets/loop-settings`):

- Each set, last seen first, with its app, source, device, and first and last seen times.
- Click one to see:
  1. what changed from the previous set of the same app (the next one of that app down the page);
  2. the settings as sent, indented to inspect.
- Delete a set there.

What changed (`sitemodels.LoopSettingsChanges`; in the API, `changes` with the previous set's `previous_id`) lists
each setting whose value differs, was added or was removed, old then new:

- **Names:** the setting's path, joined by dots (example: `overridePresets.Running.targetRange`).
- **Preset lists:** matched by each preset's name, so adding one does not change the rest. A preset without a name, or
  with an earlier preset's name, is matched by its place.
- **Glucose** (Loop's `minimumBGGuard`, `preMealTargetRange` and presets' `targetRange`; Trio's presets' `target`):
  in the site's units.
  - Loop sends it in its profile's units, which the set does not keep. Thus the units come from its size: a value
    over 30 is mg/dL.
  - AAPS's configuration has none.
- **Other values:** as sent. Text as it is, anything else as JSON.

## Truncating

- Truncating a site (My Account's Truncate, or control's) deletes the sets no profile or device report refers to
  (examples: their device reports were thinned or deleted, or a profile was deleted).
- It keeps each app's last seen set: its current settings, which AAPS sends only now and then.
- The job every 10 minutes does not delete sets. Exception: if a site is over 1.5 times its DB size limit, the job
  deletes them as a truncate does.
- Deleting a site deletes its whole DB, with `loop_settings`.

## History and limits

- The records that refer to a set show, in time order, when it was in force, while they are kept. After they are
  deleted, `first_seen_at` and `last_seen_at` keep the outline.
- A client that sends different settings on every upload adds a row each time. Uploads are rate limited, and a row is
  at most 8 KB, but a cap on rows a day may be worth adding.
