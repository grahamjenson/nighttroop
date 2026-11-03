# xDrip+ (Android)

How [xDrip+](https://github.com/NightscoutFoundation/xDrip) talks to Nightscout, and how that works with Nighttroop.
It was read from xDrip+ at commit `c42cf55c`. Paths below are under `app/src/main/java/com/eveningoutpost/dexdrip/`.

xDrip+ talks to Nightscout in two ways:

- **Nightscout Sync (REST-API)** uploads readings, treatments and device status, and can download treatments
  (`utilitymodels/NightscoutUploader.java`, driven by `utilitymodels/UploaderTask.java` and `UploaderQueue.java`).
- **Nightscout Follower** is a data source that reads another site's readings, treatments and device status
  (`cgm/nsfollow/NightscoutFollow.java`).

It does not use Socket.IO or API v3. (Its "MongoDB" upload writes to a database directly, which Nighttroop has no
equivalent of.) The tests are in `tests/api/nightscout/xdrip/`.

## Authentication

Both put the secret in the URL as user info, and send its SHA1, in lowercase hex, in an `api-secret` header:

- Uploader: the base URL is `https://<secret>@<site>/api/v1/` (several can be given, separated by spaces). The
  user info is cut out of the URL (`baseURI.replaceFirst("//[^@]+@", "//")`, `NightscoutUploader.java:499`) and sent as
  `Hashing.sha1()` of it (`:508`, and `:407` for downloads). A URL without `/v1/` and without a secret uses the old
  API with no `api-secret` at all (`doLegacyRESTUploadTo`, `:544`).
- Follower: the URL is `https://<secret>@<site>`; `/api/v1/` is added when missing (`cgm/nsfollow/utils/NightscoutUrl.java`
  `getURI`) and `getHashedSecret` is the SHA1 of the user info. Every call sends `api-secret` and
  `User-Agent: xDrip+ <version>` (`NightscoutFollow.java:48-62`).

Neither sends a `token` query or `Authorization` header, so a Nighttroop token is only ever seen as its SHA1. For
Nighttroop the secret is a device's Nightscout token, which is URL safe: `https://<token>@<subdomain>.ns.<host>/api/v1/`.
Nighttroop keeps, for each identity with a token, the SHA256 of the token's SHA1 (`site_identities.nightscout_token_digest`),
and the Nightscout API looks up an `api-secret` of 40 characters by it. A password reset gives a new token and digest.
A wrong secret gets 401. A read only token's SHA1 reads, and gets 403 on a write.

## Uploader

`UploaderTask` collects what is queued (readings, meter readings, calibrations and treatments, at most 300 of each)
and calls `NightscoutUploader.uploadRest`. For each base URL it calls, in order, all with `api-secret`:

### `GET status.json` (`doStatusUpdate`, `:216`)

When first used, then once a day (an hour after an error). The body is stored and only `version` is read
(`getNightscoutVersion`, `:206`): a version starting `0.6`, `0.7` or `0.8` stops treatment downloads. Nighttroop's
status has no `version`, which xDrip+ reads as "Unknown", which is allowed. `X-Powered-By` containing `Express` or
`ASP.NET`, or `Accept-Encoding` with `gzip`, makes it gzip later request bodies (`checkGzipSupport`, `:1069`);
Nighttroop sends neither, so bodies are not gzipped.

### `POST entries` (`doRESTUploadTo`, `:557`)

One JSON array of every queued record, readings first, `Content-Type: application/json`:

| Type | Fields (`populateV1API...`) |
| --- | --- |
| `sgv` (`:662`) | `device` (`xDrip-<collection method>`, e.g. `xDrip-DexcomG5`, with the source when `nightscout_device_append_source_info`), `date` (ms), `dateString` and `sysTime` (`yyyy-MM-dd'T'HH:mm:ss.SSSZ` in the phone's zone, e.g. `2026-11-02T23:00:00.000+1300`), `sgv` (whole mg/dL), `delta` (mg/dL per 5 min, 3 decimals), `direction` (`DoubleUp` ... `DoubleDown`, `NONE`, `NOT COMPUTABLE` or `NotComputable`), `type`, `filtered` and `unfiltered` (raw x 1000), `rssi` (100), `noise` (1 to 4) |
| `mbg` (`:710`, `:727`) | `device`, `type`, `date`, `dateString`, `mbg` (mg/dL, may have decimals), `sysTime`. A calibration is also sent as an `mbg` unless a blood test is within a minute of it |
| `cal` (`:745`) | `device`, `type`, `date`, `dateString`, `slope`, `intercept`, `scale`, `sysTime`; not sent when the slope is 0 |

Only the status code is read. Anything but 2xx throws, so nothing is marked done and the whole batch, growing, is
sent again on the next reading until its queue entries are a week old; the device status and treatments after it are
not sent meanwhile.

### `POST devicestatus` (`postDeviceStatus`, `:1119`)

After a good entries upload, one object (not an array) for each battery: the phone always, a bridge (`bridge_battery`),
a Wi-Fi uploader, and with `send_ob1dex_tx_battery_to_nightscout` the Dexcom transmitter
(`utilitymodels/NightscoutBatteryDevice.java`):

```json
{"device": "Google Pixel 7", "uploader": {"battery": 87, "type": "PHONE"}}
{"device": "G6 Transmitter", "uploader": {"days": 50, "daysEstimate": "50 / 100", "status": "OK", "voltagea": 285,
 "voltageb": 275, "temperature": 30, "battery": "50 days (voltage: 285/275)", "lastQueried": 1793613019000,
 "type": "DEXCOM_TRANSMITTER"}}
```

The transmitter's `battery` is text. A failure is logged and ignored.

### Treatments (`postTreatments`, `:795`)

With `send_treatments_to_nightscout`. Every insert and update is a `PUT treatments` of one object
(`populateV1APITreatmentEntry`, `:773`, then `:884`):

```json
{"timestamp": 1793612539000, "eventType": "<none>", "enteredBy": "xdrip", "notes": "pizza",
 "uuid": "8c5f0c3a-1b2d-4e6f-9a7b-0c1d2e3f4a5b", "carbs": 45.0, "insulin": 4.5,
 "insulinInjections": "[{\"units\":4.5,\"insulin\":\"Novorapid\"}]", "created_at": "2026-11-02T09:42:19Z",
 "sysTime": "2026-11-02T22:42:19.000+1300", "_id": "8c5f0c3a1b2d4e6f9a7b0c1d"}
```

- `eventType` is `<none>` for every bolus, carbs and note (`models/Treatments.java:228`), or `Sensor Start` /
  `Sensor Stop` (`:399`, `:424`).
- `_id` is `uuid_to_id(uuid)` (`:245`): the uuid without dashes, cut to 24.
- `insulinInjections` is a JSON string, only for multiple injections; `notes` is left out when empty; `carbs` and
  `insulin` are always there, 0 when none. `enteredBy` is `xdrip` or `xdrip pos:<n>`. Treatments that came from
  Nightscout (`enteredBy` ending `via Nightscout`) are not sent back.
- They are sent one at a time; the first that is not 2xx stops the rest, and a 404 shows "Please ensure careportal
  plugin is enabled". The upload is still a success, so every queued treatment is marked done.

A delete finds the treatment first, `GET treatments.json?find[uuid]=<uuid>`, takes `_id` of the first result, and sends
`DELETE treatments/<_id>`. When the lookup finds nothing it gives up without retrying.

### `POST activity` (`postHeartRate`, `postStepsCount`, `postMotionTracking`, `:923`)

Only in engineering mode with `use_pebble_health`. Arrays of up to 500 of `{"type": "hr-bpm", "timeStamp", "created_at",
"bpm", "accuracy"}`, `{"type": "steps-total", ..., "steps"}` and `{"type": "motion-class", ..., "class"}`.
Nighttroop keeps the `type` as the activity type, a heart rate's `bpm` as its heart rate (0 to 300) and a motion's
`class` as its name; `accuracy` is dropped.

### Treatment download (`doRESTtreatmentDownload`, `:354`)

With `cloud_storage_api_download_enable`, after each upload and at most once a minute. `GET treatments` with no query
and the header `BROKEN-If-Modified-Since` (named so on purpose, so it is ignored). A 304 skips it. On a 200 it keeps
`Last-Modified` and skips the body when `Etag` is the same as last time. The array is read by
`NightscoutTreatments.processTreatmentResponse`: `_id` (required), `uuid`, `eventType`, `enteredBy` (one starting
`xdrip` is its own), `created_at` (the time, ISO 8601, required), `carbs`, `insulin`, `insulinInjections`, `notes`, and
for a finger stick `glucoseType` `Finger`, `glucose` and `units` (`mmol` or mg/dL). A treatment matching one it has by
uuid, `_id`, or time within a minute with the same amounts is not added again.

## Follower

`NightscoutFollow.work`, about every reading interval. The base URL's paths are absolute, all `GET` with
`api-secret` and the `User-Agent`, parsed with Gson (whole numbers tolerate decimals and text, a `{}` is an empty list):

| Call | When | Reads |
| --- | --- | --- |
| `/api/v1/entries.json?count=10&rr=<now ms>` | no reading yet | each `Entry`: `date` (ms, else `sysTime` or `dateString`), `sgv` (mg/dL, read as a whole number), `filtered`, `unfiltered`, `noise`, and `_id`, `delta`, `direction`, `device`, `rssi`, `type`, which are not used |
| `/api/v1/entries.json?count=2880&find[date][$gt]=<ms>&rr=<now ms>` | since the latest reading, at most a day back | the same |
| `/api/v1/treatments` | `nsfollow_download_treatments`, at most once a minute | as the uploader's treatment download |
| `/api/v1/devicestatus.json?count=1` | every 5 minutes | `uploaderBattery`, else `uploader.battery`; `isCharging`; `pump.battery.percent`; `pump.reservoir` |

`rr` busts caches. Every entry is taken as a reading, whatever its `type`, unless one is within 10 seconds of it. A
4xx on device status turns it off for 6 hours.

## Quirks

- **Units.** xDrip+ sends mg/dL with no `units`; Nighttroop takes a value over 30 as mg/dL, so readings of 30 mg/dL or
  less would be taken as mmol/L (xDrip+ sends 39 for a sensor's LOW). The follower reads `sgv` as mg/dL, but Nighttroop
  gives it in the identity's units, then the site's. A follower on an mmol/L site shows 6.5 mmol/L as 6 mg/dL, so set
  the site, or the identity the follower uses, to mg/dL. A device added for xDrip+ in the site's Add Device has its
  units set to mg/dL (#524), and an admin can set an existing device's units from its row (#523).
- `dateString` has an offset without a colon (`+1300`), and wins over `date` when there is no `utcOffset`; both are
  the same time.
- The returned treatment has `enteredBy` and `uuid` as sent, so xDrip+ matches its own by them, but no
  `insulinInjections`, and eventType `Note` for `<none>`.
- An `mbg` entry has no `sgv` (#514), as in Nightscout, but the follower takes every entry as a reading
  (`cgm/nsfollow/EntryProcessor.java`), so a finger stick becomes a reading of 0, as it would from Nightscout.
- A search gives at most 1000, newest first; the follower asks for 2880.
- `rr` is an unknown query parameter, which is logged and ignored.

## Status in Nighttroop

| Interaction | Status |
| --- | --- |
| `api-secret` as the SHA1 of a device token (uploader and follower) | Fixed in #512 |
| Old API (`/api/` without `/v1/`, no secret, `doLegacyRESTUploadTo`) | Not supported: Nighttroop refuses writes without a token |
| `GET status.json` | Works (no `version`, read as allowed) |
| `POST entries` of `sgv` and `mbg` | Works, duplicates ignored |
| `POST entries` with a `cal` | Fixed in #512: the calibration is left out, so the batch is not refused |
| `POST devicestatus`, phone and bridge battery | Fixed in #512: `uploader.battery` was only kept for Loop |
| `POST devicestatus`, Dexcom transmitter (text `battery`) | Fixed in #512: the battery is left out rather than the status refused |
| `PUT treatments` with eventType `<none>` (bolus, carbs, note) | Fixed in #512: kept as a Note |
| `PUT treatments`, `Sensor Start`, and again to update | Works |
| `PUT treatments`, `Sensor Stop` | Not supported: refused, and the treatments queued after it are lost (TODO P1) |
| `GET treatments.json?find[uuid]=` then `DELETE treatments/<_id>` | Fixed in #512: `find[uuid]` was ignored, so the newest treatment was deleted |
| `GET treatments` download, `Last-Modified` and `Etag` | Works |
| `POST activity`, steps | Works |
| `POST activity`, heart rate and motion | Fixed in this PR (only the time was kept) |
| Follower `entries.json?count=10` and `find[date][$gt]` | Works when the follower's identity is in mg/dL, as a device added for xDrip+ is (#524); in mmol/L the values are wrong |
| Follower `treatments` | Works |
| Follower `devicestatus.json?count=1`, `uploaderBattery` | Works |
| Follower takes `mbg` entries as readings | As in Nightscout: since #514 they carry only `mbg`, so the follower reads `sgv` 0 |
| `enteredBy` and `uuid` given back as sent | `enteredBy` fixed in #547; `uuid` fixed in this PR |
| Socket.IO | Not used |
