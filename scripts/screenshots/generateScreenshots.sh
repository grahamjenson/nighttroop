#!/bin/sh
# Deletes tests/screenshots and regenerates it from the SCREENSHOT steps of
# the browser suite. CI runs this directly and then fails if the result
# differs from what is committed. On a Mac run scripts/screenshots/screenshots.sh
# instead, which runs this inside a Linux container so the files match CI.
set -e
cd "$(dirname "$0")/../.."
rm -rf .test/savedScreenshots tests/screenshots
go test -count=1 ./tests/bdd
mkdir -p tests/screenshots
cp -R .test/savedScreenshots/. tests/screenshots/
echo "generated $(find tests/screenshots -name '*.png' | wc -l | tr -d ' ') screenshots in tests/screenshots"
