# iAPS and Nighttroop

## Overview

[iAPS](https://github.com/Artificial-Pancreas/iAPS) is an iOS closed loop app (oref). Trio grew from it.

- iAPS uploads, downloads and authenticates as Trio does. See [Trio and Nighttroop](trio.md). This page gives only
  the differences.
- No Socket.IO, API v2 or API v3.
- Checked against iAPS's `main` branch:
  - `FreeAPS/Sources/Services/Network/NightscoutAPI.swift`, called from `NightscoutManager.swift`.
  - `FreeAPS/Sources/Models/NightscoutTreatment.swift`, the treatment it uploads.
- Tests: `tests/api/nightscout/iaps/`, and `tests/api/nightscout/trio/iaps.http` for the fetches.

## Set up

As for Trio: use Add Device on the site, enter the site URL in iAPS, and enter the device's token as the API secret.

## Uploads

Each treatment is a `NigtscoutTreatment`, with `enteredBy: "iAPS"`.

| Upload | Fields that Trio does not send |
| --- | --- |
| Carbs (`CarbsStorage`) | `created_at` is when the carbs are eaten (`actualDate`). `creation_date` is when they were entered. `id`, `fpuID` |
| Carbs from the pump (`PumpHistoryStorage`) | `creation_date` is the same as `created_at` |
| Bolus | `bolus`: the pump event as an object |
| Temp basal | `raw_rate` and `raw_duration`: the pump events as objects |
| Manual glucose | `{ eventType: "BG Check", glucose (a string), glucoseType: "Manual", units: "mg/dl" or "mmol", notes: "iAPS User" }` |
| Override | `{ eventType: "Exercise", created_at, duration, enteredBy: "iAPS", notes (its name) }` |

Nighttroop:

- Keeps `creation_date` and gives it back. It writes `creation_date` as `toISOString` does:
  `2026-11-01T12:01:00.000Z`.
- Does not keep `fpuID`, which iAPS now sends empty, or `bolus`, `raw_rate` and `raw_duration`, which iAPS does not read back.

## Deletes

When the user deletes a record in iAPS, iAPS deletes it by query. Each request sends `api-secret`.

| Record (function) | Request: `DELETE /api/v1/treatments.json?...` |
| --- | --- |
| Carbs (`deleteCarbs`) | `find[carbs][$exists]=true&find[creation_date][$eq]=<ISO date>` |
| Insulin (`deleteInsulin`) | `find[bolus][$exists]=true&find[created_at][$eq]=<ISO date>` |
| Manual glucose (`deleteManualGlucose`) | `find[glucose][$exists]=true&find[created_at][$eq]=<ISO date>` |
| Override, changed or ended (`deleteOverride`) | `find[eventType]=Exercise&find[enteredBy]=iAPS&find[created_at][$gte]=<date - 2 s or 5 min>&find[created_at][$lte]=<date + 2 s or 5 min>` |
| Override, cancelled (`deleteNSoverride`) | `find[eventType]=Exercise&count=1&find[enteredBy]=iAPS` |
| Remote commands (`deleteAnnouncements`) | `find[eventType]=Announcement&find[created_at][$gte]=<now>` |
| All overrides (`deleteAllNSoverrrides`, for developers) | `find[eventType]=Exercise` |

- Nightscout deletes all that the query finds. It does not use `count`.
- Nighttroop:
  - Takes `find[bolus][$exists]` as "has insulin", and `find[glucose][$exists]` as "has a glucose".
  - Finds carbs by the `creation_date` it keeps.
  - Uses `count` as a limit: `deleteNSoverride` deletes only the newest override of iAPS, as its code comment says.
    Older overrides, and overrides entered in Nightscout, stay.
  - Refuses `deleteAllNSoverrrides` with a 400. A delete by type alone could delete all the site's overrides.
  - Refuses a delete with a `count` that is not 1 to 10, or with `$gte` and `$lte` more than 2 days apart.
  - Refuses the treatment filters (`carbs`, `bolus`, `glucose`, `creation_date`, `enteredBy`) on
    `/api/v1/entries.json`. A reading search does not use them.
  - Soft deletes what the search finds, at most 1000 a request.

## Quirks

- iAPS sends `glucoseType: "Manual"` with a manual glucose. Nighttroop does not take this glucose type, so it does
  not store the record.
- iAPS percent-encodes its `enteredBy` values in `find[enteredBy][$ne]` before `URLComponents` encodes them again.
  See [Trio](trio.md#quirks).

## Status on Nighttroop

| Interaction | Nighttroop |
| --- | --- |
| Uploads, fetches and announcements | As for [Trio](trio.md#status-on-nighttroop) |
| `creation_date` | Fixed: kept and given back |
| `fpuID`, `bolus`, `raw_rate`, `raw_duration` | Not kept. iAPS sends `fpuID` empty, and does not read the others back |
| Manual glucose upload (`glucoseType: "Manual"`) | Not stored |
| Deletes of carbs, insulin, manual glucose, overrides and announcements | Fixed (were a 400) |
| `deleteAllNSoverrrides` | Refused (400) |
