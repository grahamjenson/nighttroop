#!/bin/sh
# Regenerates tests/screenshots by running the browser suite in a Linux
# container that matches CI, so captures made on macOS are never committed.
# Usage: scripts/screenshots.sh [extra go test args, e.g. -run TestSite/screenshots.bdd]
set -e
cd "$(dirname "$0")/.."
docker build -q -t nighttroop-screenshots -f scripts/Dockerfile.screenshots scripts
docker run --rm -v "$PWD":/src \
	-v nighttroop-go-mod:/root/go/pkg/mod \
	-v nighttroop-go-build:/root/.cache/go-build \
	nighttroop-screenshots go test -count=1 ./tests/bdd "$@"
