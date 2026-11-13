# Test scripts

The API, Socket.IO and browser tests are scripts, read by one parser (`tests/dsl/flow`):

- `.http` (`tests/api`): API requests.
- `.sio` (`tests/socket`): also Socket.IO connections and live update streams.
- `.bdd` (`tests/bdd`): pages in Chromium.
- `.flow`: all three (example: `tests/bdd/site/live_mix.flow`).

The package imports nothing from nighttroop, so it works alone. A `flow.Runner` gets its HTTP client, a Socket.IO
client for CONNECT (`NewSocket`; here nighttroop's `socketio.NewClient`) and a page for browser steps.

## Writing one

A line `HTTP` or `BROWSE` switches the lines after it. A `.bdd` file starts with browser steps, others with HTTP. A
`.steps` file continues as the file that `INCLUDE`s it was, and a switch in it lasts until it ends.

- **HTTP steps**, as in a `.sio` file: `NAME`, then one of:
  - a request (`GET <url>`, `POST <url>`…, with `H` headers and a `B` JSON body);
  - `CONNECT <url>` (with `NS` and a `B` auth), `EMIT <event>` (with `B` or `ARGS`, and `ACK`), or `DISCONNECT`;
  - `STREAM <url>`: a page's live updates (`…/events`), with its `H` headers;
  - `WAIT <event>` or `COLLECT <event>`: an event from the socket or a stream (example: `created`);
  - `RUN_BACKGROUND_JOBS`.
- **Browser steps**, as in a `.bdd` file: `GOTO`, `CLICK`, `INPUT`, `TYPE` and the rest.
- **Asserts and `SET`** after a step check or keep its result. In the browser:
  - `SAVE_TEXT <locator> <key>` keeps an element's text as `{{ .key }}`;
  - `SAVE_VALUE <locator> <key>` keeps an input's value, even hidden (example: the TOTP secret behind a new site's
    QR code, `#public--site-info--totp-secret`).
- **Values** from `{{ }}`:
  - `{{ .name.field }}` or `{{ dig "name.0.field" }}`: from an earlier step;
  - `{{ site_url "abcdefgh" }}`: a site's address; `{{ server_url }}`: the server's;
  - `{{ env "PASSWORD" }}`: an environment variable;
  - `{{ totp .create_site.totp_secret }}`: the code an authenticator app shows when the step runs.
    - A dev or test server takes its dev TOTP secret's code and `123123`, so a script runs on it as on staging.
    - A code signs an identity in once. Signing it in again within 30 seconds waits for the next code.

## Running one against staging

`cmd/nighttroop-script` runs scripts against another server (example: staging), to make sites with real data and
check them:

```
go run ./cmd/nighttroop-script --host staging.nighttroop.nz --root cmd/staging -test.v new_site.flow
```

- **URLs:** `localhost`, or a host under it (example: `abcdefgh.ns.localhost`), goes to that name under `--host`,
  over `--scheme` (default https). Thus a script runs as written, in a test and against staging.
- **Values:** `--var KEY=VALUE` sets `{{ .KEY }}`. Give passwords and TOTP secrets in the environment, for
  `{{ env }}`, never in a script.
- **Background jobs:** `RUN_BACKGROUND_JOBS` is refused, as only a test's own server can run its jobs at once. On
  staging, write data through the API, as a phone does.
- **Browser:** needs Playwright's Chromium (`go run github.com/mxschmitt/playwright-go/cmd/playwright install chromium`).
  `--headed` shows it. `--timeout` is each step's wait. `--out` gets `SCREENSHOT`'s and a failed script's screenshots.
- **Tests:** each script is a test. `-test.v` prints every step and response. `-test.run` selects scripts.

`cmd/staging` scripts:

- **`create_site.flow`** creates a site from the home page, as a person does. It saves the admin's password and TOTP
  secret from the page, checks a code from the secret, signs in with another, and screenshots the credentials and
  the summary:

  ```
  go run ./cmd/nighttroop-script --host staging.nighttroop.nz --root cmd/staging --out screenshots create_site.flow
  ```

- **`new_site.flow`** makes a site with an hour of readings, a meal bolus and a phone's status. Its `create_site`
  response prints the credentials.
- **`sign_in.flow`** signs in to the site's page with them and screenshots its summary:

  ```
  SITE=abcdefgh PASSWORD=... TOTP_SECRET=... go run ./cmd/nighttroop-script --host staging.nighttroop.nz --root cmd/staging sign_in.flow
  ```

## Test output

- In `go test ./tests/api/`, `./tests/bdd/` and `./tests/socket/`, only a failing flow shows:
  - a red `✗ FLOW FAILED`;
  - its steps, the last one failed;
  - the server's last 200 log lines while it ran, warnings yellow, errors red.
- A failed assert leads with its file, line, and expected and actual values, then the dimmed response or page.
- `NO_COLOR=1` gives plain text.
- `FLOW_STEP_SCREENSHOTS=1` screenshots every browser step into `allScreenshots` (script run or `go test ./tests/bdd/`),
  to see where a test went wrong. It about doubles a browser test's time, so it is off otherwise, even with `-v`.
