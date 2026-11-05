#!/usr/bin/env bash
# Fails if a migration that is on the base branch was changed, deleted or renamed. A DB already at a migration's
# version never runs it again, so a change to one would reach new DBs only, and every DB made before would keep the old
# schema; see docs/migrations.md. New migrations, numbered after the last, are what change a schema.
#
#   scripts/check_migrations.sh [base]    e.g. scripts/check_migrations.sh origin/main (the default)
set -euo pipefail

base=${1:-origin/main}
changed=$(git diff --name-only --diff-filter=MDR "$base...HEAD" -- 'data/sql/*/migrations/*.sql')
if [ -n "$changed" ]; then
	echo "These migrations are on $base, so DBs have run them, and they cannot change:"
	echo "$changed"
	echo "Add a new migration numbered after the last instead; see docs/migrations.md."
	exit 1
fi
echo "No migration on $base changed"
