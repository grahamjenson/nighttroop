# xDrip4iOS and Nighttroop

How xDrip4iOS ([xdripswift](https://github.com/JohanDegraeve/xdripswift), read at `ba37f26`, October 2026) talks to
Nightscout, and how well each part works with Nighttroop. Paths below are in that repository; the tests are in
`tests/api/nightscout/xdrip4ios/`.

xDrip4iOS is a CGM app. As a **master** it uploads readings, calibrations, treatments and the uploader's battery, and
keeps its treatments in sync both ways. As a **Nightscout follower** it downloads readings. In both modes it reads the
profile and the latest device statuses, to show a loop's status. It uses only REST API v1, and no Socket.IO.

## Setting up

- **URL**: the site's Nightscout host, e.g. `https://<site>.ns.<host>`. The app removes a trailing `/` and any
  `/api/v1`, and takes a port from it (`SettingsViewNightscoutSettingsViewModel`).
- **Token**: the PWD phone's Nightscout token to upload, or a read only device's token to only follow.
- **API_SECRET**: leave it empty. Nighttroop has no API secret, and the app only ever sends the secret's SHA1, which
  Nighttroop takes only as the SHA1 of a device's token (#512), so the token works here too and anything else is a 401.
- **Units**: set the token's identity's units to mg/dL. The follower reads `sgv` as mg/dL whatever the `units`, and
  Nighttroop returns readings in the identity's (then the site's) units, so on an mmol/L site 6.5 mmol/L is read as
  6.5 mg/dL. xDrip4iOS is not one of Add Device's clients, which set mg/dL (#524), so add it as Other and an admin sets
  its units from its row on Users & Devices (#523).

## Authentication

- Every request has `?token=<token>` when a token is set (`NightscoutSyncManager.nightscoutRequest` appends it
  unescaped; `uploadDataAndGetResponse` and `Endpoint.getEndpointForLatestNSEntries` use `URLQueryItem`).
- With an API_SECRET set, every request also has an `api-secret` (or `API-SECRET`) header of its lowercase hex SHA1
  (`String.sha1()`). Nighttroop reads the token in the query first, so a stray secret doesn't get in the way.
- The settings' Test button (`SettingsViewNightscoutSettingsViewModel.testNightscoutCredentials`) sends no query. It
  sends `api-secret: <SHA1 of the secret>`, or with no secret, `api-secret: <token>` as it is. Nighttroop takes a token
  in `api-secret`, so a token alone passes.
- A URL with user info (`https://secret@host`) sets the API_SECRET from it.
- No JWT, no `Authorization` header, no `If-Modified-Since`.

## Calls

All requests send `Content-Type: application/json` and `Accept: application/json`, except the follower's download and
the status check. Success is any 2xx unless noted.

### Connection checks

| When | Request | Reads |
| --- | --- | --- |
| Test button, and when the URL, secret or token change (`testNightscoutCredentials` in both files above) | `GET /api/v1/experiments/test` | status only: 2xx verified; 401 "API_SECRET is not valid" or "Token is not valid"; 400 shows the body |
| Follower service status, every 60 s on its settings page (`FollowerServiceStatusMonitor`) | `GET /api/v1/status.json`, no credentials | `status` (`ok` is operational) |

In Nightscout `experiments/test` needs an admin. In Nighttroop it needs an identity that can upload readings, so a PWD
phone or read and write device passes and a read only token gets a 401.

### Readings (master)

`uploadBgReadingsToNightscout`, after each new reading: `POST /api/v1/entries`, an array of at most 300, oldest first,
from the last uploaded one (at most 7 days back). Each is `BgReading.dictionaryRepresentationForNightscoutUpload`:

```json
{"_id": "<24 random letters and digits>", "device": "<transmitter>", "date": 1793612476000,
 "dateString": "2026-11-02T09:41:16.000Z", "sysTime": "2026-11-02T09:41:16.000Z", "type": "sgv",
 "sgv": 118, "direction": "Flat", "filtered": 118000, "unfiltered": 118000, "noise": 1}
```

`sgv` is a whole mg/dL with no `units`. The response body isn't read. A 500 whose `description.code` is 66 (Mongo's
duplicate key) also counts as success; any other failure leaves the readings to be sent again next time, so a batch
sent twice must succeed, which Nighttroop does by ignoring the duplicates.

### Calibrations (master)

`uploadCalibrationsToNightscout`: `POST /api/v1/entries` with a `cal` record for each new calibration, then an `mbg`
record for each (`Calibration+Nightscout.swift`):

```json
{"_id": "..", "device": "..", "date": 1793612476000, "dateString": "..", "sysTime": "..", "type": "cal",
 "slope": 1012.5, "intercept": -2034.1, "scale": 1}
{"_id": "..", "device": "..", "date": 1793612476000, "dateString": "..", "sysTime": "..", "type": "mbg", "mbg": 126}
```

Nighttroop keeps the `mbg` finger pricks and accepts the `cal` records without storing them.

### Uploader battery (master)

`uploadTransmitterBatteryInfoToNightscout`, after each reading upload when the transmitter's or phone's battery
has changed: `POST /api/v1/devicestatus` with
`{"uploader": {"name": "transmitter", "battery": 85}}`. For a Dexcom, `battery` is the phone's battery (0 to 100) and
the transmitter's goes under its own key, e.g. `"batteryVoltage": 301`. Nighttroop keeps `uploader.battery` as the
device status's `uploaderBattery`.

### Sensor start (master)

`uploadActiveSensorToNightscout` (and `uploadLibreLinkUpSensorStartToNightscout` for a LibreLinkUp follower):
`POST /api/v1/treatments` with `{"_id": "..", "eventType": "Sensor Start", "created_at": "..", "enteredBy": "xDrip4iOS"}`.
The response isn't read. Nighttroop keeps its `_id`, and a retry is ignored as a duplicate.

### Treatment sync (master, and a follower that uploads)

`syncWithNightscout`, at most every 10 s, takes the last 50 local treatments and:

1. **Uploads new ones** (`uploadTreatmentsToNightscout`): `POST /api/v1/treatments`, an array of
   `TreatmentEntry.dictionaryRepresentationForNightscoutUpload`. Each has `enteredBy`, `eventTime` (ISO with ms, and no
   `created_at`) and:

   | Treatment | Fields |
   | --- | --- |
   | Insulin | `eventType: "Bolus"`, `insulin` (units) |
   | Carbs | `eventType: "Carbs"`, `carbs` (g) |
   | Exercise | `eventType: "Exercise"`, `duration` (minutes) |
   | BG check | `eventType: "BG Check"`, `glucose` (mg/dL), `units: "mg/dl"`, `glucoseType: "Finger"`, or for an mmol/L user `"Finger: 6.7 mmol/L"` |
   | Basal | `eventType: "Temp Basal"`, `rate`, `duration` |
   | Note | `eventType: "Note"`, `notes` |
   | Site change, sensor start, pump battery change | the `eventType` only |

   A downloaded treatment keeps the `eventType` it had. Treatments with the same `_id` before a `-`, such as a bolus
   with carbs from Nightscout, are sent as one. The response must be a JSON array; `TreatmentNSResponse.fromNightscout`
   turns each document into one local treatment per field it has (`carbs`, `insulin`, `glucose` with `units`, `rate`
   with `duration`, `duration` of an Exercise, or the event type), and `matchesTreatmentEntry` matches one to an upload
   by `created_at` to the millisecond, type, value and notes, to learn its `_id`. A treatment not matched is uploaded
   again next sync.
2. **Uploads edits** (`updateTreatmentToNightscout`): `PUT /api/v1/treatments`, one treatment, with its `_id` in the
   body.
3. **Downloads** (`getLatestTreatmentsNSResponses`): `GET /api/v1/treatments?find[created_at][$gte]=<24 hours ago>`,
   no `count`. It needs `_id` (string) and `created_at` (ISO, with or without fractional seconds), adds treatments it
   doesn't know, and takes any change to a known one's value, time, `enteredBy` or `notes`.
4. **Checks for deletions** (`reconcileRemoteTreatmentDeletions`): for each uploaded treatment of the last 24 hours
   missing from 3, one at a time, `GET /api/v1/treatments?find[_id]=<id>&count=1`. Only a 2xx with `[]` hides it in the
   app; anything else keeps it.
5. **Deletes** (`deleteTreatmentAtNightscout`): `DELETE /api/v1/treatments/<id>`, or a `PUT` without the deleted part
   when it shares its `_id` with others.

### Readings (follower)

`NightscoutFollowManager.performDownload`, every reading interval:
`GET /api/v1/entries/sgv.json?count=<n>&find[date][$gte]=<ms>&token=<token>` (`Endpoint+Nightscout.swift`), with only
an `api-secret` header when a secret is set. `find[date][$gte]` is the newest stored reading's time, or a day ago; `n`
is that time at one reading every 45 s, plus 1, so up to 1921. It needs a 200 and a JSON array, and reads only `sgv`
(a number, mg/dL) and `date` (a number, ms) of each (`FollowerBgReading.init(json:)`). Nighttroop returns at most 1000.

### Profile

`updateProfile`, each sync and at most every 30 s: `GET /api/v1/profile`, decoded with `JSONDecoder` as
`[NightscoutProfileResponse]` (`NightscoutProfileModels.swift`), and the first is used if its `startDate` is newer. It
requires `_id`, `defaultProfile`, `startDate` (a string) and `store`; and in the default profile (or else the first)
`basal`, `carbratio`, `sens`, `target_low` and `target_high` as arrays of `{time: string, value: number, timeAsSeconds:
integer}`, `timezone` (string), `dia` (number) and `units` (string, mg/dL when it lowercases to `mg/dl`). A missing one
fails the whole decode.

### Device status

`updateDeviceStatus`, each sync: `GET /api/v1/devicestatus` (no `count`), decoded as `[NightscoutDeviceStatusResponse]`
(`NightscoutDeviceStatusModels.swift`). Every field is optional, but a value of the wrong type fails the whole decode:
`mills`, `uploaderBattery`, `utcOffset`, `uploader.battery`, `pump.battery.percent` and `duration`s are integers. It
reads `created_at`, `device`, `isCharging`, `uploader`, `uploaderBattery`, `pump` (reservoir, battery, status,
suspended) and the `openaps` or `loop` IOB, COB, predictions and enacted temp basal.

### History import

`NightscoutImportService`, when the user imports history, a day at a time: `GET /api/v1/entries/sgv.json`,
`/api/v1/treatments.json` and `/api/v1/devicestatus.json` with `find[date]` or `find[created_at]` `$gte` and `$lt`
and a `count` of 2000, 1000 and 1000, and `GET /api/v1/profile`. A response as long as the `count` is split in half
and asked again. Its decoding is lenient.

### Deleting and rewriting readings (master)

When the user deletes a reading or smooths or rebuilds readings: `DELETE /api/v1/entries.json?find[type]=sgv` with
`find[date][$gte]` and `find[date][$lte]`, or with up to 50 `find[date][$in][]`, reading `n`, `deletedCount` or
`result.n`; then the new values are `POST`ed to `/api/v1/entries` without `_id`, to replace the stored ones.

## Quirks

- `enteredBy` comes back as sent, `xDrip4iOS`, as Nightscout does.
- Carbs come back as a `Carb Correction`; the app reads carbs from `carbs` whatever the event type.
- An `sgv` under 30 with no `units` is read by Nighttroop as mmol/L. A CGM's readings don't go that low.
- The history import asks for 2000 entries, more than Nighttroop's 1000, so it doesn't split a day with more.

## Status

| Interaction | Nighttroop | Test |
| --- | --- | --- |
| Token in the query, or in `api-secret` for the Test button | Works | `auth.http` |
| API_SECRET's SHA1 in `api-secret` | Works when the API_SECRET is the token (#512); any other secret is a 401 | `auth.http`, `tests/api/nightscout/xdrip/` |
| `GET /api/v1/experiments/test` | Fixed in #513 (was a 400) | `auth.http` |
| `GET /api/v1/status.json` without credentials | Not supported, a 401 (TODO P4) | |
| Upload readings, and the same batch again | Works | `uploader.http` |
| Upload calibrations (`cal` and `mbg`) | Fixed in #512 (`cal` failed the batch) | `uploader.http` |
| Uploader battery in `uploader.battery` | Fixed in #512 (was dropped) | `uploader.http` |
| Sensor start, and a retry | Works | `treatments.http` |
| Upload treatments with `eventTime`, matched by `created_at` | Fixed in #513 (`created_at` was the time received) | `treatments.http` |
| Upload `Carbs` | Fixed in #513 (was a 400) | `treatments.http` |
| Upload a BG check from an mmol/L user (`Finger: 6.7 mmol/L`) | Fixed in #513 (was a 400) | `treatments.http` |
| BG check matched and downloaded with the same value | Not supported (TODO P2) | `treatments.http` |
| Update a treatment with `PUT` | Works | `treatments.http` |
| Download the last 24 hours of treatments | Works | `treatments.http` |
| `find[_id]` of a deleted treatment is `[]` | Fixed in #513 (was a 404) | `treatments.http`, `tests/socket/nightscout/db_writes.sio` |
| Delete a treatment | Works | `treatments.http` |
| Follower readings with `count` and `find[date][$gte]` | Works, with the identity in mg/dL, set by an admin (#523); not one of Add Device's mg/dL clients (TODO P1) | `follower.http` |
| Profile | Works | `follower.http` |
| Device status | Works | `follower.http` |
| History import | Works, but a day with over 1000 readings loses its oldest, as a search gives at most 1000 by design | |
| Delete readings by a query | Not supported (TODO P2) | |
| Rewrite readings by posting them again | Not supported, the stored ones are kept (TODO P1) | |
