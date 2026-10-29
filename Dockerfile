# The server is one static binary, as its SQLite is pure Go. It runs on Alpine for its CA certificates and busybox
# wget, which the health check uses. The keyset tools are built beside it, so a server needs no Go to make or read keys.
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/nighttroop main.go && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/create_keysets scripts/keysets/create_keysets.go && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/decrypt_admin_bundle scripts/keysets/decrypt_admin_bundle.go

FROM alpine:3.24
RUN adduser -D -H -u 10001 nighttroop && \
    mkdir /app /data && chown nighttroop:nighttroop /data
COPY --from=build /out/ /usr/local/bin/
# The databases and their local backups live in /data; /app is empty, so no stray .env is read
ENV MODE=production DATA_DIR=/data HOST_PORT=3000
WORKDIR /app
USER nighttroop
EXPOSE 3000
VOLUME /data
# The server only answers its own host name, so the check sends it. The start period covers restoring from S3.
HEALTHCHECK --interval=10s --timeout=5s --start-period=120s --retries=3 \
    CMD wget -q -O /dev/null --header "Host: ${HOST_NAME}" "http://127.0.0.1:${HOST_PORT}/health" || exit 1
ENTRYPOINT ["nighttroop"]
