#!/bin/sh
# Starts the dev server on Fly.io (docs/deploy/dev_flyio), whose databases are kept on a volume between deploys. It runs
# as root, as Fly mounts the volume owned by root: it gives the volume to the server's user, wipes the databases only
# when a migration change means they no longer fit, then runs the server as that user.
set -e
DATA_DIR="${DATA_DIR:-/data}"
chown nighttroop:nighttroop "$DATA_DIR"
if ! su-exec nighttroop dbs_are_valid "$DATA_DIR"; then
  echo "dev_start: the databases no longer fit the migrations, so they are wiped"
  rm -rf "$DATA_DIR/db" "$DATA_DIR/backup"
fi
exec su-exec nighttroop nighttroop "$@"
