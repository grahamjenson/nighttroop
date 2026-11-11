# Nighttroop

[![CI](https://github.com/grahamjenson/nighttroop/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/grahamjenson/nighttroop/actions/workflows/ci.yml)

**Test it out** on the staging server, <https://staging.nighttroop.nz>, which runs the newest `main`. Its data is wiped
now and then, so keep nothing on it you need.

Nighttroop is a multi user rewrite of Nightscout.
Its core goals are to be:

1. Simple         - Single binary deployment, limited and optional external dependencies
1. Cheap Per User - It should consume very little resources per user
1. Multi-user     - A single server should support at least 10,000 users

To run your own Nighttroop look at the `docs/deploy` folder.

**Nighttroop is not a medical device.** It does not provide medical advice and is not a replacement for proper medical
care. Its data and alarms may be late, missing or wrong, so do not make treatment decisions from it; see the
[Terms of Service](server/assets/docs/TERMS_OF_SERVICE.md).

The Dexcom Share and LibreLinkUp syncs (`clients/dexcom_client`, `clients/libre_client`) use unofficial, undocumented
APIs, the ones other open-source followers use. They may break without warning, and using them may be against those
services' terms.

# GoTHSS Framework. (Go + Templ + HTMx + SQLite + SQLc + Tailwind)

The "framework" that Nighttroop uses uses Go and SQLc, relying heavily on generated code build by tools like SQLc, Templ and Tailwind.

# Security 

We generate the `tink-go` library to generate keysets for encryption, hashing, JWT with `go run utils/keysets/create_keysets.go`

These keysets are uploaded to the server to the ENVAR `KEY_SETS`

On initial startup you will have to set the control admin username, password and save and prove the TOTP.



# Licence

Nighttroop is free software under the [GNU Affero General Public License v3.0](LICENSE) (AGPL-3.0). You can use,
change and share it under that licence's terms. If you run a changed Nighttroop for others over a network, the AGPL
asks you to offer them its source.

## Third-party code and data

These are committed to this repository, or built into its server image, under their own licences:

| What | Where | Licence |
| ---- | ----- | ------- |
| [htmx](https://htmx.org) and its SSE extension | `server/assets/static/htmx.js`, `htmx_sse.js` | [0BSD](https://github.com/bigskysoftware/htmx/blob/master/LICENSE) |
| [Floating UI](https://floating-ui.com) core and dom 1.7.0 | `server/assets/static/floating_ui_core.js`, `floating_ui_dom.js` | [MIT](https://github.com/floating-ui/floating-ui/blob/master/LICENSE) |
| [QRCode.js](https://github.com/davidshimjs/qrcodejs) | `server/assets/static/qrcode.js` | [MIT](https://github.com/davidshimjs/qrcodejs/blob/master/LICENSE) |
| [Apache ECharts](https://echarts.apache.org) 5.6.0 | `server/assets/static/echarts.min.js` | [Apache-2.0](https://github.com/apache/echarts/blob/master/LICENSE), with its [NOTICE](https://github.com/apache/echarts/blob/master/NOTICE) |
| [Redoc](https://github.com/Redocly/redoc) 2.5.4 | `server/assets/static/redoc.standalone.js` | [MIT](https://github.com/Redocly/redoc/blob/main/LICENSE); the libraries it bundles are listed in `redoc.standalone.js.LICENSE.txt` |
| [templUI](https://templui.io) 1.11.2 components | `server/components/ui`, `utils/templui`, and their scripts in `server/assets/static` | [MIT](https://github.com/templui/templui/blob/main/LICENSE) |
| [Lucide](https://lucide.dev) icons | `server/components/ui/icon/icon_data.go` | [ISC](https://github.com/lucide-icons/lucide/blob/main/LICENSE) |
| [DB-IP](https://db-ip.com) IP to Country Lite | `utils/geoip/dbip-country-lite.mmdb` | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) |
| [DB-IP](https://db-ip.com) IP to City Lite | downloaded into the server image when it is built (see `Dockerfile`) | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) |

[IP Geolocation by DB-IP](https://db-ip.com): sign in locations are estimated with DB-IP's free Lite databases. The
built in country database is DB-IP's, unchanged. Pages that show a location link to DB-IP, as its licence asks.

The Go modules Nighttroop builds with are listed in `go.mod`, each under its own licence.
