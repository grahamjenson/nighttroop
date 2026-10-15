#!/bin/sh
# Runs generateScreenshots.sh inside a Linux container built like the CI
# runner, so the golden screenshots made on a Mac are the same as CI makes.
# The image is pinned to amd64: arm64 Chromium antialiases a few pixels
# differently, so on Apple silicon it runs under Rosetta.
set -e
cd "$(dirname "$0")/../.."
docker build -q --platform linux/amd64 -t nighttroop-screenshots -f scripts/screenshots/Dockerfile.screenshots scripts/screenshots
docker run --rm --platform linux/amd64 -v "$PWD":/src \
	-v nighttroop-go-mod:/root/go/pkg/mod \
	-v nighttroop-go-build:/root/.cache/go-build \
	nighttroop-screenshots sh scripts/screenshots/generateScreenshots.sh
