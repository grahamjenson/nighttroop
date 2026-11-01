#!/bin/sh
# Starts the dev server on Fly.io (docs/deploy/dev_flyio), whose databases are kept on a volume between deploys. It runs
# as root, as Fly mounts the volume owned by root: it gives the volume to the server's user, migrates the databases or,
# if one cannot be, deletes them all, then runs the server as that user.
set -e
DATA_DIR="${DATA_DIR:-/data}"
chown nighttroop:nighttroop "$DATA_DIR"
su-exec nighttroop try_migrate_or_delete "$DATA_DIR"
exec su-exec nighttroop nighttroop "$@"
