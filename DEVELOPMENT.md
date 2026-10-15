# Development

This document describes the setup needed to develop Nighttroop.

# Generated Code

This project uses generated code extensively, at the root if you run `go generate  ./...` it will generate all the code.

If a file has a prefix `z_<file>` or suffix `.gen.go` the code is generated and should not be altered.

The primary generators are:

1. SQLc which
2. goverter
3. go migrate
4. custom scripts like `getters` which live in `gen` folders

# Submitting changes

# Testing locally

Docker runs our local test s3 server `minio` (the only external dependency)

```
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
```

```
brew install docker colima docker-compose golang-migrate tailwindcss go-air templ sqlc
go install github.com/jmattheis/goverter/cmd/goverter@v1.10.0
go install github.com/mickamy/slog-pretty/cmd/spretty@latest


colima start
docker-compose up -d
```

```
air | spretty
```

If you want to wipe data and start again

```
docker-compose restart && rm -rf .dbs && go generate ./... && sleep 2 && air | spretty
```

Then go to `localhost:8080` and create a user, `control.localhost:8080` and login as control.
In Development all initial passwords are `password` and all TOTP is `123123`


# Testing

# BDD

We are using playwright to run integration tests

have to install the driver with

```
go run github.com/mxschmitt/playwright-go/cmd/playwright@v0.6201.0 install chromium
```

# Screenshots

`SCREENSHOT` steps in `tests/bdd` write `tests/screenshots`, which is committed
and rendered on Linux. On macOS they go to `.test/screenshots` instead; to
update the committed set run `scripts/screenshots.sh` (needs Docker).

# Coverage


```
got -coverpkg=./... -coverprofile=coverage.out tests && go tool cover -html=coverage.out
```


# Profiling

```
got -cpuprofile=cpu.pprof gover
```

```
go tool pprof -http=:8080 cpu.pprof 
```


# Intercepting client Calls to add values

```
brew install mitmproxy
```

Run the mitmproxy

```
mitmweb --listen-port 80 --mode reverse:http://localhost:8080 --set keep_host_header=true
```

Then run the server 

```
air | spretty
```

Now you can catch and copy all calls to the server

### AAPS

AAPS run in a simulator through Android Studio. Setup local DNS to point to your IP address (this is handwavy will come up with a standard way later)

create a file AAPS/extras/engineering_mode to enable RandomBG (easiest way is create a screen )

Can use `adb push init.tmp /sdcard/init.txt` 

AAPS doenst support `http` NS URLs out of the box so will need to edit code, mainly the okhttp3 client code that taked the baseURL and adds `https://` to it, you should be able to just replace that.

### Carer TODO

### Loop TODO

### Trio TODO