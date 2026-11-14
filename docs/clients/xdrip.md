# xDrip+ (Android)

How [xDrip+](https://github.com/NightscoutFoundation/xDrip) talks to Nightscout, and how it works with Nighttroop.

## Overview

- Source: commit `c42cf55c`. Paths are under `app/src/main/java/com/eveningoutpost/dexdrip/`.
- Tests: `tests/api/nightscout/xdrip/`.
- Not used: Socket.IO, API v3.
- Not supported: the "MongoDB" upload (writes to a database directly; Nighttroop has no equivalent).

- Modes:
  - **Nightscout Sync (REST-API)**: uploads readings, treatments and device status, and downloads treatments
    (`utilitymodels/NightscoutUploader.java`, from `utilitymodels/UploaderTask.java` and `UploaderQueue.java`).
  - **Nightscout Follower**: a data source. Reads another site's readings, treatments and device status
    (`cgm/nsfollow/NightscoutFollow.java`).

## Set up

1. Add a device for xDrip+. Its units are mg/dL (#524).
2. Copy the device's Nightscout token (URL safe).
3. Uploader: set the base URL to `https://<token>@<subdomain>.ns.<host>/api/v1/`. Separate more URLs with spaces.
4. Follower: set the URL to `https://<token>@<site>`. xDrip+ adds a missing `/api/v1/`
   (`cgm/nsfollow/utils/NightscoutUrl.java` `getURI`).

## Authentication

- **xDrip+ sends:**
  - the user info removed from the URL (`baseURI.replaceFirst("//[^@]+@", "//")`, `NightscoutUploader.java:499`)
  - its SHA1, lowercase hex, in `api-secret`: uploader `Hashing.sha1()` (`:508`; `:407` for downloads), follower
    `getHashedSecret`
  - follower only: `User-Agent: xDrip+ <version>` (`NightscoutFollow.java:48-62`)
  - no `token` query, no `Authorization` header. Nighttroop sees the token only as its SHA1.
- **Nighttroop checks:**
  - It keeps the SHA256 of each token's SHA1 (`site_identities.nightscout_token_digest`), and finds a 40 character
    `api-secret` by it.
  - A password reset gives a new token and digest.
  - Wrong secret: 401 in JSON. Read only token: reads, 403 on a write.

## Uploads (Nightscout Sync)

`UploaderTask` collects queued readings, meter readings, calibrations and treatments (at most 300 of each) and calls
`NightscoutUploader.uploadRest`. It sends the requests below, in this order, to each base URL, with `api-secret`.

### Status check

`GET status.json` (`doStatusUpdate`, `:216`), at first use, then daily (an hour after an error):

- Reads only `version` (`getNightscoutVersion`, `:206`); keeps the body. `0.6`, `0.7` or `0.8` stops treatment
  downloads.
- Nighttroop sends no `version`. xDrip+ reads "Unknown", which is allowed.
- Gzip: on `X-Powered-By` with `Express` or `ASP.NET`, or `Accept-Encoding` with `gzip` (`checkGzipSupport`,
  `:1069`). Nighttroop sends neither, so bodies are not gzipped.

### Entries

`POST entries` (`doRESTUploadTo`, `:557`): one JSON array of all queued records, readings first,
`Content-Type: application/json`:

| Type | Fields (`populateV1API...`) |
| --- | --- |
| `sgv` (`:662`) | `device`, `date` (ms), `dateString`, `sysTime`, `sgv` (whole mg/dL), `delta` (mg/dL per 5 min, 3 decimals), `direction`, `type`, `filtered` and `unfiltered` (raw x 1000), `rssi` (100), `noise` (1 to 4) |
| `mbg` (`:710`, `:727`) | `device`, `type`, `date`, `dateString`, `mbg` (mg/dL, can have decimals), `sysTime` |
| `cal` (`:745`) | `device`, `type`, `date`, `dateString`, `slope`, `intercept`, `scale`, `sysTime`. Not sent when the slope is 0 |

- `device`: `xDrip-<collection method>` (for example `xDrip-DexcomG5`), plus the source with
  `nightscout_device_append_source_info`.
- `dateString`, `sysTime`: `yyyy-MM-dd'T'HH:mm:ss.SSSZ` in the phone's zone (for example
  `2026-11-02T23:00:00.000+1300`).
- `direction`: `DoubleUp` ... `DoubleDown`, `NONE`, `NOT COMPUTABLE` or `NotComputable`.
- A calibration is also sent as an `mbg`, unless a blood test is within a minute of it.
- **Response:** only the status code is read. On a non-2xx, nothing is marked done. The whole batch, growing, is sent
  again with the next reading, until its queue entries are a week old. Until then, the device status and treatments
  are not sent.

### Device status

`POST devicestatus` (`postDeviceStatus`, `:1119`), after a good entries upload. One object (not an array) for each
battery (`utilitymodels/NightscoutBatteryDevice.java`): the phone (always), a bridge (`bridge_battery`), a Wi-Fi
uploader, the Dexcom transmitter (`send_ob1dex_tx_battery_to_nightscout`).

```json
{"device": "Test Phone", "uploader": {"battery": 87, "type": "PHONE"}}
{"device": "G6 Transmitter", "uploader": {"days": 50, "daysEstimate": "50 / 100", "status": "OK", "voltagea": 285,
 "voltageb": 275, "temperature": 30, "battery": "50 days (voltage: 285/275)", "lastQueried": 1793613019000,
 "type": "DEXCOM_TRANSMITTER"}}
```

The transmitter's `battery` is text. A failure is logged and ignored.

Watch faces and followers read the newest numeric `battery` of the last 30 minutes, from any device:
`/api/v2/properties` (`upbat`) and `/pebble` (`battery`).

### Treatments

With `send_treatments_to_nightscout`, each insert and update is a `PUT treatments` of one object (`postTreatments`,
`:795`; `populateV1APITreatmentEntry`, `:773`, then `:884`):

```json
{"timestamp": 1793612539000, "eventType": "<none>", "enteredBy": "xdrip", "notes": "pizza",
 "uuid": "8c5f0c3a-1b2d-4e6f-9a7b-0c1d2e3f4a5b", "carbs": 45.0, "insulin": 4.5,
 "insulinInjections": "[{\"units\":4.5,\"insulin\":\"Novorapid\"}]", "created_at": "2026-11-02T09:42:19Z",
 "sysTime": "2026-11-02T22:42:19.000+1300", "_id": "8c5f0c3a1b2d4e6f9a7b0c1d"}
```

- `eventType`: `<none>` for bolus, carbs and note (`models/Treatments.java:228`); `Sensor Start` / `Sensor Stop`
  (`:399`, `:424`).
- `_id`: `uuid_to_id(uuid)` (`:245`), the uuid without dashes, cut to 24.
- `timestamp`: ms. `created_at`: whole seconds. Nighttroop keeps the time from `timestamp`, to the ms.
- `insulinInjections`: a JSON string, only for multiple injections.
- `notes`: left out when empty. `carbs`, `insulin`: always there, 0 when none.
- `enteredBy`: `xdrip` or `xdrip pos:<n>`. Treatments from Nightscout (`enteredBy` ends `via Nightscout`) are not
  sent back.
- **Errors:** sent one at a time; the first non-2xx stops the rest. A 404 shows "Please ensure careportal plugin is
  enabled". The upload still counts as a success, so all queued treatments are marked done.
- **Delete:**
  1. `GET treatments.json?find[uuid]=<uuid>`.
  2. Take `_id` of the first result.
  3. `DELETE treatments/<_id>`. If the lookup finds nothing, xDrip+ stops without a retry.

### Activity

`POST activity` (`postHeartRate`, `postStepsCount`, `postMotionTracking`, `:923`), only in engineering mode with
`use_pebble_health`. Arrays of up to 500 of `{"type": "hr-bpm", "timeStamp", "created_at", "bpm", "accuracy"}`,
`{"type": "steps-total", ..., "steps"}` and `{"type": "motion-class", ..., "class"}`.

Nighttroop keeps `type` as the activity type, `bpm` as the heart rate (0 to 300), `class` as the motion's name. It
drops `accuracy`.

### Treatment download

- When: with `cloud_storage_api_download_enable`, after each upload, at most once a minute
  (`doRESTtreatmentDownload`, `:354`).
- Request: `GET treatments`, no query, header `BROKEN-If-Modified-Since` (named so on purpose, to be ignored).
- 304: skipped. 200: keeps `Last-Modified`; skips the body if `Etag` is unchanged.
- `NightscoutTreatments.processTreatmentResponse` reads: `_id` (required), `uuid`, `eventType`, `enteredBy` (starts
  with `xdrip`: its own), `created_at` (ISO 8601, required), `carbs`, `insulin`, `insulinInjections`, `notes`; for a
  finger stick `glucoseType` `Finger`, `glucose`, `units` (`mmol` or mg/dL).
- Not added again when it matches a known treatment by uuid, `_id`, or time within a minute with the same amounts.

## Downloads (Nightscout Follower)

`NightscoutFollow.work`, about each reading interval. All `GET`, absolute paths, with `api-secret` and `User-Agent`.
Gson parses: whole numbers accept decimals and text; `{}` is an empty list.

| Call | When | Reads |
| --- | --- | --- |
| `/api/v1/entries.json?count=10&rr=<now ms>` | No reading yet | Entries (below) |
| `/api/v1/entries.json?count=2880&find[date][$gt]=<ms>&rr=<now ms>` | Since the latest reading, at most a day back | Entries |
| `/api/v1/treatments` | `nsfollow_download_treatments`, at most once a minute | As the [treatment download](#treatment-download) |
| `/api/v1/devicestatus.json?count=1` | Every 5 minutes | `uploaderBattery`, else `uploader.battery`; `isCharging`; `pump.battery.percent`; `pump.reservoir` |

- **Entries:** uses `date` (ms, else `sysTime` or `dateString`), `sgv` (mg/dL, whole number), `filtered`,
  `unfiltered`, `noise`. Reads but does not use `_id`, `delta`, `direction`, `device`, `rssi`, `type`.
- Every entry is a reading, whatever its `type`, unless another is within 10 seconds.
- `rr` stops caches.
- A 4xx on device status turns it off for 6 hours.

## Known problems and limits

- **Units:** xDrip+ sends mg/dL without `units`. Nighttroop takes an entry without `units` as mg/dL, as Nightscout
  does, whatever the value: 39 (a sensor's LOW) and lower stay low (`low_sgv.http`). The follower reads `sgv` as
  mg/dL; the Nightscout API always gives mg/dL, whatever the identity's or site's units.
- **Times:** `dateString` has an offset without a colon (`+1300`). It wins over `date` when there is no `utcOffset`.
  Both are the same time.
- **AAPS:** gets an xDrip+ blood test (`mbg`) as a BG Check from API v3.
- **Returned treatments:** no `insulinInjections`. For `<none>`, the eventType of its amounts: `Meal Bolus` (insulin
  and carbs), `Carb Correction`, `Correction Bolus`, or `Note` (no amounts). xDrip+ finds its own by `uuid`.
- **Gaps:** a reading search gives at most 2880, newest first, so the follower gets all of them after a gap. It was
  1000: a follower offline for over 1000 readings never read the older ones.
- **`rr`:** an unknown query parameter, logged and ignored. Not a filter, unlike an unknown `find[...]`, which gives
  nothing.

## Status in Nighttroop

- **Works:**
  - `GET status.json`.
  - Entries of `sgv` and `mbg`, duplicates ignored.
  - `Sensor Start`, and again to update. `Sensor Stop`, kept as a sensor change (was refused, and the treatments
    queued after it were lost).
  - Treatment download with `Last-Modified` and `Etag`. Activity steps.
  - Follower `entries.json?count=10` and `find[date][$gt]`, with the identity in mg/dL, as a device added for xDrip+
    is (#524). In mmol/L the values are wrong.
  - Follower `treatments`, and `devicestatus.json?count=1` with `uploaderBattery`.
  - Follower `mbg` entries: as in Nightscout. Since #514 they have no `sgv`. The follower takes every entry as a
    reading (`cgm/nsfollow/EntryProcessor.java`), so a finger stick is a reading of 0.
- **Fixed, authentication:** `api-secret` as the SHA1 of a device token (#512).
- **Fixed, entries:**
  - A careportal `BG Check` sent back as an `mbg`, device `careportal via Nightscout`: kept once. It was kept twice,
    as a reading with another device.
  - A `cal` (#512): left out, so the batch is not refused.
  - A reading dated more than 15 minutes after now (X10): left out, the rest stored. It was kept, so followers showed it
    as the current reading.
  - A reading of 0 or less, or over 1000 mg/dL: left out, the rest stored. The batch was refused and sent again with
    each reading, so nothing more uploaded.
  - Only readings of 0 (a failing sensor): `200 []`, nothing stored. It was refused, so the device status and
    treatments (a bolus) after it were not sent.
  - An unknown type: left out, the rest stored (up to 10 a batch). A batch with nothing else stored is still refused.
    The site's client errors show a batch with an entry left out, with the counts and the reasons. Every v1 batch,
    such as activity, now uses this rule. Treatments (`PUT`, one at a time) and device statuses (one object) are not
    batches, so one with a problem is still refused.
- **Fixed, device status (#512):**
  - Phone and bridge battery: `uploader.battery` was only kept for Loop.
  - Dexcom transmitter (text `battery`): the battery is left out, not the status refused. The site's client errors
    show it. Its other uploader fields
    (`days`, `daysEstimate`, `status`, `voltagea`, `voltageb`, `resistance`, `resistance_status`,
    `temperature`, `lastQueried`, `type`) come back as sent.
  - `upbat` in `/api/v2/properties` and the `/pebble` battery: only a loop's status was read, so they were missing.
    Test: `xdrip/battery.http`.
- **Fixed, treatments:**
  - eventType `<none>` (bolus, carbs, note) (#512): kept as a Note.
  - `<none>` with insulin or carbs: returned as `Meal Bolus`, `Carb Correction` or `Correction Bolus`. As a Note,
    LoopFollow showed a dot only. Test: `xdrip/none_amounts.http`.
  - A careportal treatment with no id, in the same second as a `<none>` bolus: kept apart. It was merged into the
    bolus, taking its `enteredBy` and notes.
  - `notes` with HTML (for example `<b>bold</b>`): given back as sent. Tags were removed, so xDrip+ appended the
    changed note to its own each download.
  - `find[uuid]=`, then `DELETE treatments/<_id>` (#512): `find[uuid]` was ignored, so the newest treatment was
    deleted. Any other unknown `find[...]` gives nothing, so deletes nothing.
  - Delete of a downloaded treatment (for example a careportal note, Undo last): xDrip+ deletes only a 24 character
    `_id` (a MongoDB ObjectId). Nighttroop's had 8, so the delete was never sent and the treatment came back. Records
    without an `_id` now get an ObjectId-shaped one.
  - `enteredBy` (#547) and `uuid` given back as sent, so xDrip+ matches its own.
- **Fixed, activity:** heart rate and motion (only the time was kept).
- **Not supported:** the old API (a URL without `/v1/` and without a secret, so no `api-secret`,
  `doLegacyRESTUploadTo`, `:544`). Writes need a token.
- **Not used:** Socket.IO.
