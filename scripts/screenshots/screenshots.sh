#!/bin/sh
# Regenerates the golden screenshots in tests/screenshots by running the
# browser suite in a Linux container that matches CI, with fonts set up to
# render like macOS, so captures made on a Mac are never committed.
# Usage: scripts/screenshots/screenshots.sh [extra go test args, e.g. -run TestSite/screenshots.bdd]
set -e
cd "$(dirname "$0")/../.."
docker build -q -t nighttroop-screenshots -f scripts/screenshots/Dockerfile.screenshots scripts/screenshots
rm -rf .test/savedScreenshots
docker run --rm -v "$PWD":/src \
	-v nighttroop-go-mod:/root/go/pkg/mod \
	-v nighttroop-go-build:/root/.cache/go-build \
	nighttroop-screenshots go test -count=1 ./tests/bdd "$@"
# The suite writes its captures to .test/savedScreenshots (bind mounted from
# the container); promote them to the committed golden files.
mkdir -p tests/screenshots
cp -R .test/savedScreenshots/. tests/screenshots/
echo "updated $(find tests/screenshots -name '*.png' | wc -l | tr -d ' ') screenshots in tests/screenshots"
