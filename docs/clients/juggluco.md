# Juggluco

[Juggluco](https://github.com/j-kaltes/Juggluco) is an Android (and Wear OS) app that reads Libre, Dexcom and other
sensors directly. It can upload its glucose readings and the amounts entered in it (insulin, carbs and other
labelled numbers) to a Nightscout server. This page is how that uploader talks to Nightscout, checked against
Juggluco's source at commit `d8233ce` (2026-10-29), for anyone checking it works with Nighttroop.

Juggluco can also act as a local Nightscout-like web server for watches and other apps. That is not covered here.

The uploader is native C++ in `Common/src/main/cpp/net/watchserver/` (`uploader.cpp`, `uploadtreatment.cpp`,
`common.cpp`, `datestring.hpp`), which calls Java in `Common/src/main/java/tk/glucodata/NightPost.java` for HTTP.
Juggluco only writes: it never reads glucose, treatments, profiles or device statuses back. It does not upload
device statuses, and it does not use Socket.IO.

## Setting it up

In Settings, Nightscout upload, the user enters:

- **Nightscout URL**: used exactly as typed, with paths like `/api/v1/entries` appended
  (`makeuploadurls`, `uploader.cpp:66`). A trailing `/` gives `//api/...`.
- **Secret**: kept as typed (`setNightUploader`, `settings/javasettings.cpp:1482`).
- **test V3**: a checkbox, only on phones (`NightPost.config`). Wear OS always uses v1.
- **Send amounts**: whether amounts are uploaded as treatments.

With Nighttroop, tick **test V3**, use the site's Nightscout host as the URL, with no trailing slash, e.g.
`https://<site>.ns.<host>`, and a device's Nightscout token as the secret.

## Authentication

The mode decides how it authenticates (`makeuploadsecret`, `uploader.cpp:74`; `NightPost.upload`, `NightPost.java:228`):

- **v1 (default)**: every request has `api-secret: <lowercase SHA1 hex of the secret>`. There is no token in the
  query.
- **v3**: no `api-secret`. `NightPost.gettoken` (`NightPost.java:190`) first calls
  `GET <url>/api/v2/authorization/request/<secret>`, with the secret as typed in the path and no auth header. It needs
  a 200, then reads `token` (a string) and `exp` (a long, in seconds). Every later request has
  `Authorization: Bearer <token>` until the time is past `exp * 1000`, when it asks again. If asking fails, it sends
  an empty `Authorization` header and the request fails.

Every request also has `Content-Type: application/json`, and uploads a `Content-Length`. Connect timeout is 10 s and
read timeout 60 s.

## When it uploads

A background thread (`uploaderthread`, `uploader.cpp:355`) uploads glucose when a new reading arrives
(`wakestreamuploader`, from `wakewithcurrent` in `g.cpp`), and glucose then amounts when the network comes back
(`backupjava.cpp:560`) or the user taps **Send now** (`wakeuploader`). Its waits running out upload nothing, so
after a failure it tries again on the next wake, from what was not yet sent. Readings and amounts older than 30 days
(`nighttimeback`) are never sent. **Resend data** (`resetuploader`) sends everything from the last 30 days again.

Only the status code of a response is used, and the body is read only for the log. In Java, a 4xx or 5xx throws
on reading the body, so it counts as a failure.

## Glucose

### v1: `POST <url>/api/v1/entries`

`uploadCGM` (`uploader.cpp:267`) sends each sensor's new readings as one JSON array of up to 5000 readings. Only a
**200** counts as sent; then it moves on to the next sensor. Each reading is written by `mkuploaditem`
(`uploader.cpp:155`):

```json
{"type":"sgv","device":"3MH000TEST","dateString":"2026-11-01T23:00:00.000+13:00","date":1793527200000,"sgv":123,
 "delta":-1.500,"direction":"FortyFiveDown","noise":1,"filtered":123000,"unfiltered":123000,"rssi":100}
```

| Field | Type | Value |
|---|---|---|
| `type` | string | always `sgv` |
| `device` | string | the sensor's short name, e.g. `3MH000TEST` |
| `dateString` | string | local time with its offset, from `Tdatestring` (`datestring.hpp:5`), always `.000` milliseconds |
| `date` | integer | ms since the epoch, whole seconds |
| `sgv` | integer | mg/dL, calibrated if the user calibrated in Juggluco. No `units` field |
| `delta` | number | mg/dL per 5 minutes, three decimals; `0.000` when there is no trend |
| `direction` | string | `DoubleUp`, `SingleUp`, `FortyFiveUp`, `Flat`, `FortyFiveDown` or `DoubleDown`; `""` when there is no trend, which Nighttroop leaves out |
| `noise` | integer | always 1 |
| `filtered`, `unfiltered` | integer | `sgv * 1000` |
| `rssi` | integer | always 100 |

There is no `sysTime`, `utcOffset`, `_id` or `identifier` in v1.

### v3: `POST <url>/api/v3/entries`

`uploadCGM3` (`uploader.cpp:193`) posts one reading at a time, written by `writev3entry(..., server=false)`
(`common.cpp:46`). A **200 or 201** counts as sent; anything else stops the upload at that reading.

```json
{"app":"Juggluco","device":"3MH000TEST","date":1793527200000,"sgv":123,"delta":-1.500,"direction":"FortyFiveDown",
 "type":"sgv","utcOffset":0,"identifier":"54534554-3030-3048-50ee-1234eeeeeeee","_id":"5453455430303048501234ee"}
```

`date`, `sgv`, `delta`, `direction` and `device` are as in v1. `utcOffset` is always 0. `identifier` is a UUID-shaped
hex string made from the sensor name and the reading's number (`mkv3streamid`), and `_id` a 24 hex character one
(`mkv1streamid`). Both stay the same when a reading is sent again.

## Amounts (treatments)

Only sent with **Send amounts** on, and only amounts of a kind the user mapped to a Nightscout category. Each amount
is sent alone. For each new or changed amount, `uploadtreatments` (`uploadtreatment.cpp:112`) first **deletes** it
by its id, ignoring the answer, then sends it. An amount removed in Juggluco is deleted, and that delete must be a
**200**: otherwise the upload stops, and amounts stay stuck until it works.

Every amount has `eventType` `<none>` (Nightscout careportal's "no event type"). Its other fields depend on the kind:

- long-acting insulin: `"notes":"Long-Acting"`, `insulin` (units) and `insulinType` (the label, e.g. `Levemir`);
- rapid-acting insulin: `"notes":"Rapid-Acting"`, `insulin` and `insulinType`;
- carbs: `carbs` (grams);
- anything else: `"notes":"<label> <value>"`, e.g. `Walk 30`.

Insulin and carbs entered at the same time are separate treatments with the same `date`.

### v1

- Delete: `DELETE <url>/api/v1/treatments/<_id>` (`deletetreatment`, `uploadtreatment.cpp:44`).
- Send: `PUT <url>/api/v1/treatments` with one object, not an array; only a **200** counts
  (`sendtreatment`, `uploadtreatment.cpp:94`; `writetreatment`, `common.cpp:161`):

```json
{"_id":"ba0e12bbbbbbbbbbbbbbbbbb","date":1793527500000,"eventType":"<none>","enteredBy":"Juggluco",
 "created_at":"2026-11-01T10:05:00.000Z","notes":"Rapid-Acting","carbs":null,"insulin":4.5,"insulinType":"Novorapid"}
```

`_id` is 24 hex characters: `ba<kind>e<number>` padded with `b` (`mkid`), or, for amounts made after the phone's
Juggluco ID was set, `<kind><number, 7 digits><Juggluco ID, 16 hex>` (`mkidid`). Carbs come with `"insulin":null`,
insulin with `"carbs":null`, and other amounts with both null.

### v3

- Delete: `DELETE <url>/api/v3/treatments/<identifier>` (`deletetreatment3`, `uploadtreatment.cpp:57`).
- Send: `POST <url>/api/v3/treatments`; a **200 or 201** counts (`sendtreatment3`; `writetreatmentV3`,
  `uploadtreatment.cpp:175`):

```json
{"identifier":"0fffffff-ffff-ffff-ffff-12ffffffffff","_id":"ba0e12bbbbbbbbbbbbbbbbbb","date":1793527500000,
 "utcOffset":0,"eventType":"<none>","app":"Juggluco","notes":"Rapid-Acting","insulin":4.5,"insulinType":"Novorapid"}
```

`identifier` is UUID-shaped hex from the kind and number (`mkidV3`, or `mkididV3` with the Juggluco ID). There is no
`created_at` or `enteredBy`, and no null fields: carbs are only `"carbs":30`.

So an amount changed in Juggluco, even its time, is deleted, then posted again with the same identifier.

## Quirks

- The secret is SHA1 hashed in v1 but sent as typed, in the URL path, in v3.
- A trend of 2 mg/dL per minute or faster downwards is always `DoubleDown`: `getdeltaindex` (`glucose.cpp:556`)
  compares with `DOUBLETHRESHOLD` instead of its negative, so `SingleDown` is never sent.
- A zone west of UTC with half hours is written with two minus signs in `dateString`, e.g. `-03:-30` for
  Newfoundland, as `Tdatestring` takes the remainder of negative minutes.
- `sgv` has no `units`. Nighttroop guesses an unlabelled value of 30 or less is mmol/L; Libre and Dexcom readings
  start at 39 or 40 mg/dL, but a reading calibrated in Juggluco could be lower.
- v1 sends up to 5000 readings in one request, as one array, and retries the same array until it is accepted.
- A JWT that stops working before its `exp` (e.g. after a password reset) is sent until `exp` anyway.

## Nighttroop status

| Interaction | Status |
|---|---|
| `GET /api/v2/authorization/request/<token>`, reading `token` and `exp` | Works |
| v3 `POST /api/v3/entries`, one reading, 201 then 200 when sent again | Works |
| v3 entries keep Juggluco's `identifier` | Fixed in #508: a create that names its `identifier` keeps it |
| v3 `POST /api/v3/treatments` with `eventType` `<none>` | Fixed in #508: `<none>` is a Note, keeping insulin and carbs (was a 400) |
| v3 `DELETE /api/v3/treatments/<identifier>` of an amount, 200 | Fixed in #508: was a 404, as the identifier was not kept, which stopped the amounts uploading |
| v3 insulin and carbs at the same time stored as two treatments | Fixed in #508: with an identifier, only the same identifier is a duplicate |
| v3 amount changed in Juggluco (delete, then post with the same identifier) | Fixed in #508: a deleted document is replaced whatever changed |
| v1 `api-secret` with the SHA1 of the secret | Works since #512: a 40 hex `api-secret` is the SHA1 of a device's token, not an admin's or one with TOTP |
| v1 `POST /api/v1/entries` with Juggluco's readings, 200 | Works |
| v1 `dateString` with `-03:-30` | Fixed in #508: read as `-03:30` (was a 400) |
| v1 `POST /api/v1/entries` with more than 100 readings | Works: a v1 POST takes up to 5000 (it was a 400 over 100) |
| v1 `DELETE /api/v1/treatments/<_id>`, 200 | Works |
| v1 `PUT /api/v1/treatments` of one amount, 200, and a changed one after its delete | Works; fixed in #508 for a changed time |

Tests: `tests/api/nightscout/juggluco/upload_v3.http` and `upload_v1.http`. The v1 tests send the token's SHA1 in
`api-secret`, as Juggluco does.
