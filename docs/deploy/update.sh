#!/usr/bin/env bash
# Updates Nighttroop run with hostinger.compose.yml, and rolls back if the new version is not healthy.
#
#   sudo ./update.sh [image]    e.g. sudo ./update.sh ghcr.io/grahamjenson/nighttroop:v1.2.3
#
# 1. Pulls the new image, before anything is stopped
# 2. Stops Nighttroop, which uploads a last backup to S3, and copies ./data into ./backups
# 3. Starts the new image and waits for its health check, which asks the server's /health
# 4. If it is not healthy, puts back ./data from that copy and starts the previous image again
#
# HEALTH_TIMEOUT (seconds, default 180) is how long the new version has to be healthy; KEEP_BACKUPS (default 5) is how
# many copies of ./data, and previous images, are kept. COMPOSE_FILE is the compose file, as for docker compose.
set -euo pipefail

cd "$(dirname "$0")"
export COMPOSE_FILE=${COMPOSE_FILE:-hostinger.compose.yml}
COMPOSE=(docker compose)
SERVICE=nighttroop
NEW_IMAGE=${1:-ghcr.io/grahamjenson/nighttroop:latest}
HEALTH_TIMEOUT=${HEALTH_TIMEOUT:-180}
KEEP_BACKUPS=${KEEP_BACKUPS:-5}
STAMP=$(date -u +%Y%m%dT%H%M%SZ)
BACKUP=backups/data-$STAMP.tar.gz
ROLLBACK_IMAGE=nighttroop-rollback:$STAMP

log() { echo "[$(date -u +%H:%M:%S)] $*"; }

# set_image writes the image the compose file runs into .env
set_image() {
	if grep -q '^NIGHTTROOP_IMAGE=' .env; then
		sed -i "s|^NIGHTTROOP_IMAGE=.*|NIGHTTROOP_IMAGE=$1|" .env
	else
		echo "NIGHTTROOP_IMAGE=$1" >>.env
	fi
}

# wait_healthy waits for the container's health check to pass, failing once it is unhealthy or the time is up
wait_healthy() {
	local container status
	container=$("${COMPOSE[@]}" ps -q "$SERVICE")
	for ((waited = 0; waited < HEALTH_TIMEOUT; waited += 5)); do
		status=$(docker inspect -f '{{.State.Health.Status}}' "$container" 2>/dev/null || echo missing)
		case $status in
		healthy) return 0 ;;
		unhealthy) return 1 ;;
		esac
		sleep 5
	done
	return 1
}

# prune keeps the newest KEEP_BACKUPS copies of ./data and previous images
prune() {
	printf '%s\n' backups/data-*.tar.gz | sort -r | tail -n +$((KEEP_BACKUPS + 1)) | xargs -r rm --
	docker images nighttroop-rollback --format '{{.Tag}}' | sort -r | tail -n +$((KEEP_BACKUPS + 1)) |
		xargs -r -I{} docker rmi nighttroop-rollback:{} >/dev/null
}

if [ "$(id -u)" != 0 ]; then
	echo "Run as root, e.g. with sudo, so ./data is copied with its owner" >&2
	exit 1
fi
current=$("${COMPOSE[@]}" ps -q "$SERVICE")
if [ -z "$current" ]; then
	echo "Nighttroop is not running here; start it with: docker compose -f $COMPOSE_FILE up -d" >&2
	exit 1
fi
old_image=$(docker inspect -f '{{.Image}}' "$current")

log "Pulling $NEW_IMAGE"
docker pull "$NEW_IMAGE" || docker image inspect "$NEW_IMAGE" >/dev/null
if [ "$(docker image inspect -f '{{.Id}}' "$NEW_IMAGE")" = "$old_image" ]; then
	log "Already running $NEW_IMAGE"
	exit 0
fi
# The running image gets its own name, so it can start again whatever its tag points to by then
docker tag "$old_image" "$ROLLBACK_IMAGE"

log "Stopping Nighttroop, which backs up to S3 first"
"${COMPOSE[@]}" stop "$SERVICE"
log "Copying ./data to $BACKUP"
mkdir -p backups
tar -czf "$BACKUP" data

log "Starting $NEW_IMAGE"
set_image "$NEW_IMAGE"
"${COMPOSE[@]}" up -d "$SERVICE"
if wait_healthy; then
	log "Updated: $NEW_IMAGE is healthy"
	prune
	exit 0
fi

log "$NEW_IMAGE is not healthy, rolling back. Its last logs:"
"${COMPOSE[@]}" logs --tail 50 "$SERVICE" || true
"${COMPOSE[@]}" stop "$SERVICE"
mv data "data-failed-$STAMP"
tar -xzf "$BACKUP"
set_image "$ROLLBACK_IMAGE"
"${COMPOSE[@]}" up -d "$SERVICE"
if wait_healthy; then
	log "Rolled back to the previous version ($ROLLBACK_IMAGE) and its data; the failed version's data is in data-failed-$STAMP"
	exit 1
fi
log "The previous version is not healthy either; see: docker compose -f $COMPOSE_FILE logs $SERVICE"
exit 2
