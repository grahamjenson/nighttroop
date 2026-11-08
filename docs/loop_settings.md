# Loop settings

A loop app's settings, such as Loop's maximum bolus or Trio's remote control keys, change rarely. They come in on
records that are sent often:

| App  | Sent in                                   | Field                                          |
|------|-------------------------------------------|------------------------------------------------|
| Loop | a profile (`/api/v1/profile`)             | `loopSettings`                                 |
| Trio | a profile                                 | its remote settings, at the top of the profile |
| AAPS | each device status (`/api/v1/devicestatus`) | `configuration`                              |

Until now they were kept on those records. A profile store copied `loopSettings` and Trio's remote settings onto every
one of its `profile_presets` rows. Each device report kept AAPS's configuration in `other_attributes`, every 5
minutes. Each copy is the same until something changes, and there is no way to see when something did.

The `openaps` and `loop` objects in a device status are not settings. They are the loop's state at that cycle: its
IOB, COB, predictions and what it enacted. They change every cycle and stay on `device_reports`.

## The table

`loop_settings` has a row for each change of an app's settings:

- `app`: `loop`, `trio` or `aaps`.
- `source`: whether they came in a `profile` or a `devicestatus`.
- `settings_json`: the settings as the app's model writes them. Writing them through the model leaves out unknown
  fields and puts the known ones in the same order each time, so two uploads of the same settings are the same text.
- `occurred_at`: when the record they came in was from.
- `created_by_device_name`: the device the record came from.

An app's settings in force at a time are its last row at or before that time (`GetLoopSettingsAt`).

## Writing: extracting them from a Nightscout write

This is done; see `server/nightscoutapi/nscollections/loop_settings.go`.

When a profile or device status is created, `recordLoopSettings` takes the app's settings out of it. It compares them
with the app's settings in force at the record's time. It adds a row only when they differ, and gives back the row
they are either way.

- **A repeat upload** adds nothing.
- **A change** adds a row from that time.
- **A change back** adds a row too, so the history shows both changes.
- **A backfill**, such as a profile from last week sent today, is compared with the settings in force last week, not
  today's, so it adds a row only when it really differs from that time.
- **Settings over 8 KB** are not kept, but the record they came in is still kept.
- **Empty settings**, such as the `configuration: {}` AAPS sends in some statuses, count as none sent, not a change.

It runs in the same transaction as the record, so a failed write keeps neither.

## Reading: putting the records back together (next)

A Nightscout read must still return a profile with its `loopSettings`, and a device status with its `configuration`,
as they were sent. That makes them a combination of records again. Guessing which settings were in force at the
record's time would be wrong in two cases:

- **Sent without settings:** a profile sent without `loopSettings`, after one sent with them, would come back with
  them.
- **A later backfill:** a backfill of earlier settings, sent after a profile, would change what that profile reads
  back.

So each record will point at the row it was sent with. `recordLoopSettings` already gives that row:

1. Add `loop_settings_id` (and `remote_settings_id` for Trio) to `profile_presets`, and `loop_settings_id` to
   `device_reports`. Each is `REFERENCES loop_settings(id)`, set from what `recordLoopSettings` gives back.
2. Read them back with a `LEFT JOIN loop_settings` in the profile and device report queries, so a read is still one
   query. The converters (`DBToNightscoutProfileStore`, `DBToNightscoutDeviceStatus`) take the joined JSON where they
   take the copied JSON today.
3. Remove the copies: `profile_presets.loop_settings_json` and `remote_settings_json`, and `configuration_settings`
   in `device_reports.other_attributes`. The migrations are not frozen yet, so this edits `000001`.
4. Read the newest settings for push notifications (`profile_push.go`) with `GetLoopSettingsAt(app, now)`, rather
   than from the newest profile.

A record whose settings were too large has no row and reads back without them, as one over the limit is refused now.

## Afterwards

- A site page listing each change of each app's settings, with what changed.
- Deleting all of a site's data deletes its `loop_settings` too. Deleting one profile or device status leaves them, as
  other records refer to the same rows.
- Limits: a client that sends different settings on every upload adds a row each time. Uploads are already
  rate limited, and a row is at most 8 KB, but a cap on rows a day may be worth adding.
