# xDrip4iOS

How xDrip4iOS talks to Nightscout, and how each part works with Nighttroop.

## Overview

- Source: [xdripswift](https://github.com/JohanDegraeve/xdripswift) at `ba37f26` (October 2026). Paths are in that
  repository.
- Tests: `tests/api/nightscout/xdrip4ios/`.
- xDrip4iOS is a CGM app. It uses only REST API v1. Not used: Socket.IO.
- Modes:
  - **master**: uploads readings, calibrations, treatments and the uploader's battery. Syncs treatments both ways.
  - **Nightscout follower**: downloads readings.
  - Both modes read the profile and the latest device statuses, to show a loop's status.

## Set up

1. On Users & Devices, add a device for xDrip4iOS. Its units are mg/dL (#524).
2. Open its credentials. They show these steps, with its URL and token to copy.
3. Set **URL** to the site's Nightscout host, for example `https://<site>.ns.<host>`.
   - The app removes a trailing `/` and any `/api/v1`, and takes a port from it
     (`SettingsViewNightscoutSettingsViewModel`).
4. Set **Token** to the device's Nightscout token.
   - Master: a PWD phone or a read and write device. Follow only: a read only device.
5. Leave **API_SECRET** empty, or put the token in it.
   - Nighttroop has no API secret. The app sends only its SHA1, which Nighttroop takes only as the SHA1 of a device's
     token (#512). Anything else gets 401.
6. Set **Units** to mg/dL.
   - The follower reads `sgv` as mg/dL, whatever the `units`. The Nightscout API always gives mg/dL, whatever the
     identity's or site's units.
   - A device added for xDrip4iOS also has the Nightscout API write mg/dL as `mg/dl`. An admin can set this for
     another identity with `units_mgdl_label` through the settings endpoint.
   - If you use another device, its first upload of a treatment in `mg/dl` sets its label to `mg/dl`.

## Authentication

- **Query:** `?token=<token>` on each request when a token is set. `NightscoutSyncManager.nightscoutRequest` appends
  it unescaped. `uploadDataAndGetResponse` and `Endpoint.getEndpointForLatestNSEntries` use `URLQueryItem`.
- **Header:** with an API_SECRET, also `api-secret` (or `API-SECRET`) with its lowercase hex SHA1 (`String.sha1()`).
  Nighttroop reads the query token first, so a stray secret does not interfere.
- **Test button** (`SettingsViewNightscoutSettingsViewModel.testNightscoutCredentials`): no query.
  - Sends `api-secret: <SHA1 of the secret>`, or with no secret, `api-secret: <token>` as it is.
  - Nighttroop accepts a token in `api-secret`, so a token alone passes.
- A URL with user info (`https://secret@host`) sets the API_SECRET.
- Not used: JWT, the `Authorization` header, `If-Modified-Since`.

## What syncs

- All requests send `Content-Type: application/json` and `Accept: application/json`, except the follower readings
  download and the status check.
- Success is any 2xx, unless noted.

| Data | Mode | When |
| --- | --- | --- |
| Connection test | — | Test button; a change of URL, secret or token |
| Service status | Follower | Every 60 s on its settings page |
| Readings upload | Master | After each new reading |
| Calibrations | Master | Each new calibration |
| Sensor start | Master, LibreLinkUp follower | — |
| Uploader battery | Master | After each reading upload, if the transmitter's or phone's battery changed |
| Treatments | Master, or a follower that uploads | At most every 10 s |
| Readings download | Follower | Each reading interval |
| Profile | Both | Each sync, at most every 30 s |
| Device status | Both | Each sync |
| History | — | The user imports history |
| Reading deletes and rewrites | Master | The user deletes, smooths or rebuilds readings |

### Connection checks

- **Test** (`testNightscoutCredentials`): `GET /api/v1/experiments/test`. Reads the status only.
  - 2xx: verified.
  - 401: "API_SECRET is not valid" or "Token is not valid".
  - 400: shows the body.
  - Nightscout needs an admin. Nighttroop needs an identity that can upload readings: a PWD phone or read and write
    device passes, and a read only token gets 401.
- **Service status** (`FollowerServiceStatusMonitor`): `GET /api/v1/status.json`, no credentials. Reads `status`
  (`ok` is operational).

### Readings upload

`uploadBgReadingsToNightscout`: `POST /api/v1/entries`, an array of at most 300, oldest first, from the last uploaded
one (at most 7 days back). Each is `BgReading.dictionaryRepresentationForNightscoutUpload`:

```json
{"_id": "<24 random letters and digits>", "device": "<transmitter>", "date": 1793612476000,
 "dateString": "2026-11-02T09:41:16.000Z", "sysTime": "2026-11-02T09:41:16.000Z", "type": "sgv",
 "sgv": 118, "direction": "Flat", "filtered": 118000, "unfiltered": 118000, "noise": 1}
```

- `sgv`: whole mg/dL, no `units`.
- The response body is not read.
- A 500 with `description.code` 66 (Mongo's duplicate key) is also a success.
- After any other failure, the readings are sent again. A batch sent twice must succeed: Nighttroop ignores the
  duplicates.

### Calibrations

`uploadCalibrationsToNightscout` (`Calibration+Nightscout.swift`): `POST /api/v1/entries`, a `cal` record for each
new calibration, then an `mbg` record for each:

```json
{"_id": "..", "device": "..", "date": 1793612476000, "dateString": "..", "sysTime": "..", "type": "cal",
 "slope": 1012.5, "intercept": -2034.1, "scale": 1}
{"_id": "..", "device": "..", "date": 1793612476000, "dateString": "..", "sysTime": "..", "type": "mbg", "mbg": 126}
```

Nighttroop keeps the `mbg` finger pricks. It accepts the `cal` records, but does not store them.

### Uploader battery

`uploadTransmitterBatteryInfoToNightscout`: `POST /api/v1/devicestatus` with
`{"uploader": {"name": "transmitter", "battery": 85}}`.

- For a Dexcom, `battery` is the phone's battery (0 to 100). The transmitter's has its own key, for example
  `"batteryVoltage": 301`.
- Nighttroop keeps `uploader.battery` as the device status's `uploaderBattery`.

### Sensor start

`uploadActiveSensorToNightscout` (and `uploadLibreLinkUpSensorStartToNightscout` for a LibreLinkUp follower):
`POST /api/v1/treatments` with
`{"_id": "..", "eventType": "Sensor Start", "created_at": "..", "enteredBy": "xDrip4iOS"}`.

- The response is not read.
- Nighttroop keeps its `_id`, and ignores a retry as a duplicate.

### Treatment sync

`syncWithNightscout` takes the last 50 local treatments, then:

1. **Upload new ones** (`uploadTreatmentsToNightscout`): `POST /api/v1/treatments`, an array of
   `TreatmentEntry.dictionaryRepresentationForNightscoutUpload`. Each has `enteredBy`, `eventTime` (ISO with ms, no
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

   - A downloaded treatment keeps its `eventType`.
   - Treatments with the same `_id` before a `-` (for example a bolus with carbs from Nightscout) are sent as one.
   - The response must be a JSON array. `TreatmentNSResponse.fromNightscout` makes one local treatment for each field
     a document has: `carbs`, `insulin`, `glucose` with `units`, `rate` with `duration`, the `duration` of an
     Exercise, or the event type.
   - `matchesTreatmentEntry` matches each to an upload, to learn its `_id`. It compares `created_at` to the
     millisecond, type, value and notes. A treatment that does not match is uploaded again at the next sync.
2. **Upload edits** (`updateTreatmentToNightscout`): `PUT /api/v1/treatments`, one treatment, with its `_id` in the
   body.
3. **Download** (`getLatestTreatmentsNSResponses`): `GET /api/v1/treatments?find[created_at][$gte]=<24 hours ago>`,
   no `count`.
   - Needs `_id` (string) and `created_at` (ISO, with or without fractional seconds).
   - Adds unknown treatments. Takes any change to a known one's value, time, `enteredBy` or `notes`.
4. **Check for deletions** (`reconcileRemoteTreatmentDeletions`): for each uploaded treatment of the last 24 hours
   that step 3 did not return, one at a time: `GET /api/v1/treatments?find[_id]=<id>&count=1`.
   - Only a 2xx with `[]` hides it in the app. Any other response keeps it.
5. **Delete** (`deleteTreatmentAtNightscout`): `DELETE /api/v1/treatments/<id>`. If it shares its `_id` with others, a
   `PUT` without the deleted part.

### Readings download

`NightscoutFollowManager.performDownload` (`Endpoint+Nightscout.swift`):
`GET /api/v1/entries/sgv.json?count=<n>&find[date][$gte]=<ms>&token=<token>`.

- Header: only `api-secret`, when a secret is set.
- `find[date][$gte]`: the newest stored reading's time, or one day ago.
- `n`: that time span at one reading every 45 s, plus 1. Thus up to 1921.
- Needs a 200 and a JSON array.
- Reads only `sgv` (a number, mg/dL) and `date` (a number, ms) of each (`FollowerBgReading.init(json:)`).
- Nighttroop returns at most 1000.

### Profile

`updateProfile`: `GET /api/v1/profile`, decoded with `JSONDecoder` as `[NightscoutProfileResponse]`
(`NightscoutProfileModels.swift`). The app uses the first, if its `startDate` is newer.

- Nighttroop, as Nightscout, gives the newest 10 stores, newest first.
- Required: `_id`, `defaultProfile`, `startDate` (a string) and `store`.
- Required in the default profile (or else the first):
  - `basal`, `carbratio`, `sens`, `target_low`, `target_high`: arrays of
    `{time: string, value: number, timeAsSeconds: integer}`
  - `timezone` (string), `dia` (number)
  - `units` (string, mg/dL when it lowercases to `mg/dl`)
- If one is missing, the whole decode fails.

### Device status

`updateDeviceStatus`: `GET /api/v1/devicestatus`, no `count`. Decoded as `[NightscoutDeviceStatusResponse]`
(`NightscoutDeviceStatusModels.swift`).

- Every field is optional. A value of the wrong type fails the whole decode.
- Integers: `mills`, `uploaderBattery`, `utcOffset`, `uploader.battery`, `pump.battery.percent` and each `duration`.
- Reads: `created_at`, `device`, `isCharging`, `uploader`, `uploaderBattery`, `pump` (reservoir, battery, status,
  suspended), and the `openaps` or `loop` IOB, COB, predictions and enacted temp basal.

### History import

`NightscoutImportService`, one day at a time:

- `GET /api/v1/entries/sgv.json`, `/api/v1/treatments.json` and `/api/v1/devicestatus.json`, with `find[date]` or
  `find[created_at]` `$gte` and `$lt`, and a `count` of 2000, 1000 and 1000.
- `GET /api/v1/profile`.
- If a response is as long as the `count`, the app splits it in half and asks again. Its decoding is lenient.

### Reading deletes and rewrites

- **Delete a reading** (`deleteBgReadingFromNightscout`): `DELETE /api/v1/entries.json?find[type]=sgv` with
  `find[date][$gte]` and `find[date][$lte]`. Reads `n`, `deletedCount` or `result.n`.
- **Smooth or rebuild readings** (`replaceBgReadingsInNightscout`):
  1. Deletes the readings a rebuild hides, with up to 50 `find[date][$in][]` in each request.
  2. Posts the new values to `/api/v1/entries` as uploaded, but without `_id`. It relies on the Nightscout upsert on
     `sysTime` and `type` to replace the stored ones.
- Nighttroop:
  - Gives the stored reading of that date, type and device the new values. A reading sent again with the same values
    stays as it is.
  - Soft deletes the readings a delete finds, at most 1000 a request, and returns `deletedCount`.
  - Accepts only these filters (and `token`). Any other filter, or no date filter, gets 400. Thus a filter it does
    not understand never deletes more than was asked.

## Known problems and limits

- `enteredBy` comes back as sent, `xDrip4iOS`, as in Nightscout.
- Carbs come back as a `Carb Correction`. The app reads carbs from `carbs`, whatever the event type.
- Nighttroop reads an `sgv` under 30 with no `units` as mmol/L. A CGM's readings do not go that low.
- The history import asks for 2000 entries, more than Nighttroop's 1000. Thus it does not split a day with more.

## Status in Nighttroop

Tests are in `tests/api/nightscout/xdrip4ios/`, unless a path is given.

- **Works:**
  - Token in the query, or in `api-secret` for the Test button (`auth.http`).
  - The API_SECRET's SHA1 in `api-secret`, when the API_SECRET is the token (#512). Any other secret gets 401
    (`auth.http`, `tests/api/nightscout/xdrip/`).
  - Readings upload, and the same batch again (`uploader.http`).
  - Sensor start and a retry; `PUT` update; download of the last 24 hours; delete (`treatments.http`).
  - Follower readings with `count` and `find[date][$gte]`, with a device added for xDrip4iOS (in mg/dL), or an
    identity an admin sets to mg/dL (#523) (`follower.http`,
    `tests/api/nighttroop/settings/identity_clients_units.http`).
  - Profile and device status (`follower.http`).
  - Delete readings by a query: soft deleted, at most 1000 a request. It was 405 (`deletes.http`).
  - Rewrite readings by a new post: the stored reading takes the new values. It was kept as it was (`rewrite.http`).
  - History import, but a day with over 1000 readings loses its oldest: a search gives at most 1000 by design. No
    test.
- **Fixed:**
  - #513: `GET /api/v1/experiments/test`, was 400 (`auth.http`).
  - #512: calibrations, `cal` and `mbg`. A `cal` failed the batch (`uploader.http`).
  - #512: uploader battery in `uploader.battery`, was dropped (`uploader.http`).
  - #513: treatments with `eventTime`, matched by `created_at`. `created_at` was the time received
    (`treatments.http`).
  - #513: `Carbs`, and a BG check from an mmol/L user (`Finger: 6.7 mmol/L`). Both were 400 (`treatments.http`).
  - #513: `find[_id]` of a deleted treatment is `[]`, was 404 (`treatments.http`,
    `tests/socket/nightscout/db_writes.sio`).
  - A BG check matched and downloaded with the same value. A device added for xDrip4iOS has the Nightscout API write
    mg/dL as `mg/dl`, so it comes back as the `120` `mg/dl` it sent
    (`tests/api/nighttroop/settings/units_labels.http`).
  - A BG check from a device not added for xDrip4iOS was sent again every 15 s. It came back as `mg/dL`, so the
    app read 110 as 1982 and did not match it. Now the device's first `mg/dl` upload sets its label to `mg/dl`
    (`tests/api/nighttroop/settings/units_label_learned.http`).
- **Will not fix:** `GET /api/v1/status.json` without credentials gets 401, as from a Nightscout with
  `AUTH_DEFAULT_ROLES=denied`. The follower service status shows a fetch error. A 200 would let LoopFollow connect
  without a token (`auth.http`).
