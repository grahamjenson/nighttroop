# Development

This document describes the setup needed to develop Nighttroop.


# Submitting changes

# Testing locally

Docker runs our local test s3 server `minio` (the only external dependency)

```
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
```

```
brew install docker colima docker-compose golang-migrate tailwindcss go-air templ sqlc
go install github.com/jmattheis/goverter/cmd/goverter@v1.9.3
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