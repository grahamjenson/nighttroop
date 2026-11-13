# Juggluco

How the Juggluco Nightscout uploader talks to Nightscout, and how it works with Nighttroop.

## Overview

- [Juggluco](https://github.com/j-kaltes/Juggluco): an Android (and Wear OS) app that reads Libre, Dexcom and other
  sensors directly.
- Uploads: glucose readings, and amounts entered in it (insulin, carbs, other labelled numbers).
- Write only: never reads glucose, treatments, profiles or device statuses back. No device status upload, no
  Socket.IO.
- Not covered: Juggluco as a local Nightscout-like web server for watches and other apps.
- Source: commit `d8233ce` (2026-10-29). Native C++ in `Common/src/main/cpp/net/watchserver/` (`uploader.cpp`,
  `uploadtreatment.cpp`, `common.cpp`, `datestring.hpp`), which calls Java in
  `Common/src/main/java/tk/glucodata/NightPost.java` for HTTP.
- Tests: `tests/api/nightscout/juggluco/upload_v3.http` and `upload_v1.http`. The v1 tests send the token's SHA1 in
  `api-secret`, as Juggluco does.

## Set up

1. In Juggluco, open Settings, Nightscout upload.
2. Set **Nightscout URL** to the site's Nightscout host, for example `https://<site>.ns.<host>`.
   - Do not add a trailing `/`. Juggluco adds paths such as `/api/v1/entries` to the URL exactly as typed, so it
     gives `//api/...` (`makeuploadurls`, `uploader.cpp:66`).
3. Set **Secret** to a device's Nightscout token. Juggluco keeps it as typed (`setNightUploader`,
   `settings/javasettings.cpp:1482`).
4. Tick **test V3**. This checkbox is only on phones (`NightPost.config`). Wear OS always uses v1.
5. To upload amounts as treatments, turn on **Send amounts**.

## Authentication

The mode sets the method (`makeuploadsecret`, `uploader.cpp:74`; `NightPost.upload`, `NightPost.java:228`):

- **v1 (default):** each request has `api-secret: <lowercase SHA1 hex of the secret>`. No token in the query.
- **v3:** no `api-secret`. `NightPost.gettoken` (`NightPost.java:190`) gets a JWT:
  1. `GET <url>/api/v2/authorization/request/<secret>`, with the secret as typed in the path and no auth header.
  2. Needs a 200. Reads `token` (a string) and `exp` (a long, in seconds).
  3. Sends `Authorization: Bearer <token>` on each later request, until the time is past `exp * 1000`. Then it asks
     again.
    - If this fails, Juggluco sends an empty `Authorization` header, and the request fails.

Every request has `Content-Type: application/json`, and an upload has a `Content-Length`. Timeouts: connect 10 s,
read 60 s.

## When it uploads

A background thread (`uploaderthread`, `uploader.cpp:355`) uploads:

| Event | Uploads |
| --- | --- |
| A new reading (`wakestreamuploader`, from `wakewithcurrent` in `g.cpp`) | Glucose |
| The network comes back (`backupjava.cpp:560`) | Glucose, then amounts |
| The user taps **Send now** (`wakeuploader`) | Glucose, then amounts |
| The user taps **Resend data** (`resetuploader`) | Everything from the last 30 days, again |

- A wait that times out uploads nothing. After a failure, the next wake sends what is not yet sent.
- Readings and amounts older than 30 days (`nighttimeback`) are never sent.
- Only the status code is used; the body is read only for the log. In Java, a 4xx or 5xx throws on reading the body,
  so it is a failure.

## Glucose

### v1: `POST <url>/api/v1/entries`

`uploadCGM` (`uploader.cpp:267`) sends each sensor's new readings as one JSON array of up to 5000. Only a **200**
counts as sent; then the next sensor starts. `mkuploaditem` (`uploader.cpp:155`) writes each reading:

```json
{"type":"sgv","device":"3MH000TEST","dateString":"2026-11-01T23:00:00.000+13:00","date":1793527200000,"sgv":123,
 "delta":-1.500,"direction":"FortyFiveDown","noise":1,"filtered":123000,"unfiltered":123000,"rssi":100}
```

| Field | Type | Value |
|---|---|---|
| `type` | string | always `sgv` |
| `device` | string | the sensor's short name, for example `3MH000TEST` |
| `dateString` | string | local time with its offset, from `Tdatestring` (`datestring.hpp:5`), always `.000` milliseconds |
| `date` | integer | ms since the epoch, whole seconds |
| `sgv` | integer | mg/dL, calibrated if the user calibrated in Juggluco. No `units` field |
| `delta` | number | mg/dL per 5 minutes, three decimals. `0.000` when there is no trend |
| `direction` | string | `DoubleUp`, `SingleUp`, `FortyFiveUp`, `Flat`, `FortyFiveDown` or `DoubleDown`. `""` when there is no trend, which Nighttroop leaves out |
| `noise` | integer | always 1 |
| `filtered`, `unfiltered` | integer | `sgv * 1000` |
| `rssi` | integer | always 100 |

v1 has no `sysTime`, `utcOffset`, `_id` or `identifier`.

### v3: `POST <url>/api/v3/entries`

`uploadCGM3` (`uploader.cpp:193`) posts one reading at a time, written by `writev3entry(..., server=false)`
(`common.cpp:46`). A **200 or 201** counts as sent. Anything else stops the upload at that reading.

```json
{"app":"Juggluco","device":"3MH000TEST","date":1793527200000,"sgv":123,"delta":-1.500,"direction":"FortyFiveDown",
 "type":"sgv","utcOffset":0,"identifier":"54534554-3030-3048-50ee-1234eeeeeeee","_id":"5453455430303048501234ee"}
```

- `date`, `sgv`, `delta`, `direction`, `device`: as in v1. `utcOffset`: always 0.
- `identifier`: UUID-shaped hex from the sensor name and the reading's number (`mkv3streamid`). `_id`: 24 hex
  characters (`mkv1streamid`). Both stay the same when a reading is sent again.

## Amounts (treatments)

- Sent only with **Send amounts** on, for kinds the user mapped to a Nightscout category. Each amount is sent alone.
- Insulin and carbs entered at the same time are separate treatments with the same `date`.
- New or changed amount (`uploadtreatments`, `uploadtreatment.cpp:112`): first a delete by its id (response ignored),
  then a send. Thus a change, even of the time, is a delete and a new post with the same identifier.
- Amount removed in Juggluco: a delete that must get a **200**. If not, the upload stops, and amounts stay stuck until
  it works.

Every amount has `eventType` `<none>` (Nightscout careportal's "no event type"). Other fields by kind:

| Kind | Fields |
| --- | --- |
| Long-acting insulin | `"notes":"Long-Acting"`, `insulin` (units), `insulinType` (the label, for example `Levemir`) |
| Rapid-acting insulin | `"notes":"Rapid-Acting"`, `insulin`, `insulinType` |
| Carbs | `carbs` (grams) |
| Anything else | `"notes":"<label> <value>"`, for example `Walk 30` |

### v1

- Delete: `DELETE <url>/api/v1/treatments/<_id>` (`deletetreatment`, `uploadtreatment.cpp:44`).
- Send: `PUT <url>/api/v1/treatments` with one object, not an array. Only a **200** counts (`sendtreatment`,
  `uploadtreatment.cpp:94`; `writetreatment`, `common.cpp:161`).

```json
{"_id":"ba0e12bbbbbbbbbbbbbbbbbb","date":1793527500000,"eventType":"<none>","enteredBy":"Juggluco",
 "created_at":"2026-11-01T10:05:00.000Z","notes":"Rapid-Acting","carbs":null,"insulin":4.5,"insulinType":"Novorapid"}
```

- `_id`: 24 hex characters, `ba<kind>e<number>` padded with `b` (`mkid`). For amounts made after the phone's
  Juggluco ID was set: `<kind><number, 7 digits><Juggluco ID, 16 hex>` (`mkidid`).
- Carbs have `"insulin":null`, insulin has `"carbs":null`, other amounts have both null.

### v3

- Delete: `DELETE <url>/api/v3/treatments/<identifier>` (`deletetreatment3`, `uploadtreatment.cpp:57`).
- Send: `POST <url>/api/v3/treatments`. A **200 or 201** counts (`sendtreatment3`; `writetreatmentV3`,
  `uploadtreatment.cpp:175`).

```json
{"identifier":"0fffffff-ffff-ffff-ffff-12ffffffffff","_id":"ba0e12bbbbbbbbbbbbbbbbbb","date":1793527500000,
 "utcOffset":0,"eventType":"<none>","app":"Juggluco","notes":"Rapid-Acting","insulin":4.5,"insulinType":"Novorapid"}
```

- `identifier`: UUID-shaped hex from the kind and number (`mkidV3`, or `mkididV3` with the Juggluco ID).
- No `created_at`, no `enteredBy`, no null fields. Carbs are only `"carbs":30`.

## Known problems and limits

- **Secret:** v1 sends its SHA1. v3 sends it as typed, in the URL path.
- **No `SingleDown`:** a fall of 2 mg/dL per minute or faster is always `DoubleDown` (`getdeltaindex`,
  `glucose.cpp:556`, compares with `DOUBLETHRESHOLD`, not its negative).
- **Half-hour zones west of UTC:** two minus signs in `dateString`, for example `-03:-30` for Newfoundland
  (`Tdatestring` takes the remainder of negative minutes).
- **Units:** `sgv` has no `units`. Nighttroop takes it as mg/dL, as Nightscout does, whatever the value. Thus a
  reading calibrated in Juggluco below 39 mg/dL stays low.
- **Large v1 batches:** up to 5000 readings in one array, sent again until accepted.
- **JWT:** sent until its `exp`, even if it stops working before (for example after a password reset).

## Status in Nighttroop

- **Works:**
  - `GET /api/v2/authorization/request/<token>`, with `token` and `exp`.
  - v3 `POST /api/v3/entries` of one reading: 201, then 200 when sent again.
  - v1 `api-secret` with the SHA1 of the secret (since #512). A 40 hex `api-secret` is the SHA1 of a device's token,
    not of an admin's or of one with TOTP.
  - v1 `POST /api/v1/entries` of Juggluco's readings: 200.
  - v1 `POST /api/v1/entries` of more than 100 readings: a v1 POST accepts up to 5000 (over 100 was 400).
  - v1 `DELETE /api/v1/treatments/<_id>`: 200.
  - v1 `PUT /api/v1/treatments` of one amount, and of a changed one after its delete: 200. Fixed in #508 for a
    changed time.
- **Fixed in #508:**
  - v3 entries keep Juggluco's `identifier`: a create that names its `identifier` keeps it.
  - v3 `POST /api/v3/treatments` with `eventType` `<none>`: a Note, with insulin and carbs (was 400).
  - v3 `DELETE /api/v3/treatments/<identifier>` of an amount: 200. It was 404 (the identifier was not kept), which
    stopped amounts uploading.
  - v3 insulin and carbs at the same time: two treatments. With an identifier, only the same identifier is a
    duplicate.
  - v3 amount changed in Juggluco (a delete, then a post with the same identifier): a deleted document is replaced,
    whatever changed.
  - v1 `dateString` with `-03:-30`: read as `-03:30` (was 400).
