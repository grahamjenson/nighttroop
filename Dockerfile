# The server is one static binary, as its SQLite is pure Go. It runs on Alpine for its CA certificates and busybox
# wget, which the health check uses. The keyset tool is built beside it, so a server needs no Go to make keys, and
# dbs_are_valid, which the dev server checks its databases with.
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/nighttroop main.go && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/create_keysets scripts/keysets/create_keysets.go && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/dbs_are_valid ./scripts/dbs_are_valid

# DB-IP's free IP to City Lite (CC BY 4.0), this month's or last month's, for the city each sign in is from. If it
# cannot be downloaded the image is still built, and the server knows only countries, from its built in database.
FROM alpine:3.24 AS geoip
ARG DBIP_URL=https://download.db-ip.com/free
RUN mkdir /out && cd /out && \
    for month in $(date -u +%Y-%m) $(date -u -d "@$(( $(date -u +%s) - $(date -u +%-d) * 86400 ))" +%Y-%m); do \
      wget -q -O city.mmdb.gz "$DBIP_URL/dbip-city-lite-$month.mmdb.gz" && gunzip city.mmdb.gz && \
        mv city.mmdb dbip-city-lite.mmdb && echo "DB-IP IP to City Lite $month" && break; \
      rm -f city.mmdb.gz city.mmdb; \
    done; true

FROM alpine:3.24 AS server
RUN adduser -D -H -u 10001 nighttroop && \
    mkdir /app /data && chown nighttroop:nighttroop /data
COPY --from=build /out/ /usr/local/bin/
COPY --from=geoip /out/ /usr/local/share/nighttroop/
# The databases and their local backups live in /data; /app is empty, so no stray .env is read
ENV MODE=production DATA_DIR=/data HOST_PORT=3000 GEOIP_DB_PATH=/usr/local/share/nighttroop/dbip-city-lite.mmdb
WORKDIR /app
USER nighttroop
EXPOSE 3000
VOLUME /data
# The server only answers its own host name, so the check sends it. The start period covers restoring from S3.
HEALTHCHECK --interval=10s --timeout=5s --start-period=120s --retries=3 \
    CMD wget -q -O /dev/null --header "Host: ${HOST_NAME}" "http://127.0.0.1:${HOST_PORT}/health" || exit 1
ENTRYPOINT ["nighttroop"]

# The dev server on Fly.io (docs/deploy/dev_flyio), built with --build-target dev. It starts as root to take over its
# volume and wipe its databases when they no longer fit the migrations, then runs the server as nighttroop.
FROM server AS dev
USER root
RUN apk add --no-cache su-exec
COPY scripts/dbs_are_valid/dev_start.sh /usr/local/bin/dev_start
ENTRYPOINT ["dev_start"]

# The server, the image built by default
FROM server
