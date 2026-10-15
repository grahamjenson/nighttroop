#!/bin/sh
# Deletes tests/screenshots and regenerates it by running the browser suite
# in the same Linux container locally and in CI, so the committed golden
# files and the ones CI makes are byte for byte comparable.
# Usage: scripts/screenshots/generateScreenshots.sh
set -e
cd "$(dirname "$0")/../.."
docker build -q -t nighttroop-screenshots -f scripts/screenshots/Dockerfile.screenshots scripts/screenshots
rm -rf .test/savedScreenshots tests/screenshots
docker run --rm -v "$PWD":/src \
	-v nighttroop-go-mod:/root/go/pkg/mod \
	-v nighttroop-go-build:/root/.cache/go-build \
	nighttroop-screenshots go test -count=1 ./tests/bdd
# The suite writes its captures to .test/savedScreenshots (bind mounted from
# the container); promote them to the committed golden files.
mkdir -p tests/screenshots
cp -R .test/savedScreenshots/. tests/screenshots/
echo "generated $(find tests/screenshots -name '*.png' | wc -l | tr -d ' ') screenshots in tests/screenshots"
