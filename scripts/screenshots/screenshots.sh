#!/bin/sh
# Runs generateScreenshots.sh inside the Playwright Linux container, so the
# golden screenshots made on a Mac are the same as the ones CI makes.
set -e
cd "$(dirname "$0")/../.."
docker build -q -t nighttroop-screenshots -f scripts/screenshots/Dockerfile.screenshots scripts/screenshots
docker run --rm -v "$PWD":/src \
	-v nighttroop-go-mod:/root/go/pkg/mod \
	-v nighttroop-go-build:/root/.cache/go-build \
	nighttroop-screenshots sh scripts/screenshots/generateScreenshots.sh
