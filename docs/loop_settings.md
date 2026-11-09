# Loop settings

A loop app's settings, such as Loop's maximum bolus or Trio's remote control keys, change rarely. They come in on
records that are sent often:

| App  | Sent in                                     | Field                                          | How often                               |
|------|---------------------------------------------|------------------------------------------------|-----------------------------------------|
| Loop | a profile (`/api/v1/profile`)               | `loopSettings`                                 | when its therapy settings are saved     |
| Trio | a profile                                   | its remote settings, at the top of the profile | when a schedule is saved                |
| AAPS | each device status (`/api/v1/devicestatus`) | `configuration`                                | about hourly, and `{}` in the others    |

They used to be copied onto each record: every `profile_presets` row, and every device report. Now each different set
is kept once in `loop_settings`, and each record refers to the set it was sent with.

The `openaps` and `loop` objects in a device status are not settings. They are the loop's state at that cycle: its
IOB, COB, predictions and what it enacted, with the effective ISF, CR and targets of the cycle. They change every
cycle and stay on `device_reports`, in `loop_status_json`, kept whole only on each device's newest report: the
truncate every 10 minutes removes the others' debug output and predictions.

## The table

`loop_settings` has a row for each different set of an app's settings:

- `id`: the SHA-256, in hex, of the app and the settings. The same settings from the same app are always the same
  row, whenever and however often they are sent.
- `app`: `loop`, `trio` or `aaps`.
- `source`: whether they came in a `profile` or a `devicestatus`.
- `settings_json`: the settings as the app's model writes them. Writing them through the model leaves out unknown
  fields and puts the known ones in the same order each time, and Go writes a map's keys sorted, so the same settings
  are always the same text. A Go upgrade that wrote numbers differently would at worst add a second row for the same
  settings; it can never mix up two different sets.
- `first_seen_at` and `last_seen_at`: the earliest and latest times of the records they came in.
- `created_by_device_name`: the device that first sent them.

## Writing: extracting them from a Nightscout write

`recordLoopSettings` (`server/nightscoutapi/nscollections/loop_settings.go`) runs when a profile or device status is
created, in the same transaction:

1. It writes the app's settings through its model, and works out their ID.
2. It upserts the row: a new set is added, and a set already kept only widens when it was seen. So repeat uploads,
   retries, backfills and records arriving out of order all land on the same row, and A, then B, then A again is two
   rows, not three.
3. It gives back the ID, which the record keeps: `profile_presets.loop_settings_id` and `remote_settings_id`, and
   `device_reports.configuration_id`.

Settings that are empty (`{}`, as AAPS sends in most statuses) or missing are none sent, so the record refers to no
row. Settings over 8 KB are not kept, and the record is still kept, without them.

## Reading: putting the records back together

A Nightscout read still gives a profile its `loopSettings` and Trio's remote settings, and a device status its
`configuration`, as they were sent. The converters (`DBToNightscoutProfileStore`, `DBToNightscoutDeviceStatus`) look
up the row the record refers to by its ID, one primary key lookup for each record that has one.

The reference is not a foreign key. A set of settings can be deleted, from the site's Loop Settings page, and the
records sent with it are kept: they are read back without those settings, as if none had been sent. Nothing else
changes about them.

## The site

The Data hub's Presets group has a Loop Settings page (`/presets/loop-settings`): each set, the last seen first, with
its app, what it came in, the device and when it was first and last seen. Clicking one shows the settings as they were
sent, indented to inspect. A set can be deleted there.

## Afterwards

- What changed between two sets, shown on the page.
- When settings were in force can be read from the records that refer to them, in time order, while those records
  are kept; `first_seen_at` and `last_seen_at` keep the outline after they are deleted.
- Sets no record refers to any more could be cleared when a site is truncated. Deleting a site deletes its whole DB,
  `loop_settings` with it.
- Limits: a client that sends different settings on every upload adds a row each time. Uploads are already rate
  limited, and a row is at most 8 KB, but a cap on rows a day may be worth adding.
