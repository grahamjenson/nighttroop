# Test scripts

The API, Socket.IO and browser tests are scripts, read by one parser in one package (`tests/dsl/flow`):

- `.http` files (`tests/api`) send API requests,
- `.sio` files (`tests/socket`) also open Socket.IO connections and live update streams,
- `.bdd` files (`tests/bdd`) drive the pages in Chromium,
- `.flow` files mix all three, e.g. `tests/bdd/site/live_mix.flow`.

The package imports nothing from nighttroop, so it can be used on its own: a `flow.Runner` is given its HTTP
client, a Socket.IO client for CONNECT (`NewSocket`, nighttroop's `socketio.NewClient` here) and a page for the
browser steps.

## Writing one

A line `HTTP` or `BROWSE` switches the lines after it:

- **HTTP steps** are written as in a `.sio` file. Each starts with `NAME`, then one of:
  - an HTTP request (`GET <url>`, `POST <url>`…, with `H` headers and a `B` JSON body);
  - `CONNECT <url>` (with `NS` and a `B` auth), `EMIT <event>` (with `B` or `ARGS`, and `ACK`), or `DISCONNECT`;
  - `STREAM <url>`, which opens a page's live updates (`…/events`) with its `H` headers;
  - `WAIT <event>` and `COLLECT <event>`, for an event from the socket or a stream, such as `created`;
  - or `RUN_BACKGROUND_JOBS`.
- **Browser steps** are written as in a `.bdd` file: `GOTO`, `CLICK`, `INPUT`, `TYPE` and the rest.
- **Asserts and `SET`** after a step check or keep what it got. In the browser, `SAVE_TEXT <locator> <key>` keeps an element's text as `{{ .key }}`, and `SAVE_VALUE <locator> <key>` an input's value, though it is hidden, as the TOTP secret behind a new site's QR code (`#public--site-info--totp-secret`) is.
- **Where a file starts:** a `.bdd` file starts with browser steps and any other with HTTP ones. A `.steps` file goes on as the file that `INCLUDE`s it was, and a switch in an included file lasts until it ends.
- **Values** are filled in from `{{ }}`:
  - `{{ .name.field }}` is from an earlier step, and `{{ dig "name.0.field" }}` too;
  - `{{ site_url "abcdefgh" }}` is a site's address, and `{{ server_url }}` the server's;
  - `{{ totp .create_site.totp_secret }}` is the code an authenticator app shows now, made when the step runs, so a later step's is the code then. A dev or test server takes the code its dev TOTP secret makes, as well as `123123`, so a script making codes runs on one as on staging. A code signs an identity in once, so a script signing the same identity in again within 30 seconds waits for the next code;
  - `{{ env "PASSWORD" }}` is an environment variable.

## Running one against staging

`cmd/nighttroop-script` runs scripts against a server running elsewhere, as staging, to make sites with real data and check them:

```
go run ./cmd/nighttroop-script --host staging.nighttroop.nz --root cmd/staging -test.v new_site.flow
```

- **URLs:** a URL on `localhost`, or a host under it such as `abcdefgh.ns.localhost`, goes to that name under `--host`, over `--scheme` (https by default), so a script runs as written in a test and against staging.
- **Values and secrets:** `--var KEY=VALUE` sets `{{ .KEY }}`. Give passwords and TOTP secrets in the environment, for `{{ env }}`, so they are never written in a script.
- **Background jobs:** a script with `RUN_BACKGROUND_JOBS` is refused, as only a test's own server can run its jobs at once. On staging, data is written through the API instead, as a phone would.
- **Browser steps** need Playwright's Chromium (`go run github.com/mxschmitt/playwright-go/cmd/playwright install chromium`). `--headed` shows the browser, `--timeout` is how long each step waits, and screenshots, `SCREENSHOT`'s and a failed script's, go to `--out`.
- **Running tests:** each script runs as a test, so `-test.v` prints every step and its response, and `-test.run` picks which scripts run.
- **A screenshot of every browser step**, into `allScreenshots`, is taken with `FLOW_STEP_SCREENSHOTS=1`, in a script run or `go test ./tests/bdd/`, to see where a test went wrong. It about doubles a browser test's time, so it is off otherwise, even with `-v`.

`cmd/staging` has three:

- **`create_site.flow`** creates a site from the home page as a person would: it saves the admin's password and TOTP secret from the page, checks a code made from the secret, signs in with another and screenshots the credentials and the summary:

```
go run ./cmd/nighttroop-script --host staging.nighttroop.nz --root cmd/staging --out screenshots create_site.flow
```

- **`new_site.flow`** makes a site with an hour of readings, a meal bolus and a phone's status, and prints the site's credentials in its `create_site` response.
- **`sign_in.flow`** signs in to a site's page with them and screenshots its summary:

```
SITE=abcdefgh PASSWORD=... TOTP_SECRET=... go run ./cmd/nighttroop-script --host staging.nighttroop.nz --root cmd/staging sign_in.flow
```
