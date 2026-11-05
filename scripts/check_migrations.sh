#!/usr/bin/env bash
# Fails if a migration that is on the base branch was changed, deleted or renamed, once the migrations are frozen: once
# data/sql/MIGRATIONS_FROZEN is on the base branch, which it is from the first production deploy. A DB already at a
# migration's version never runs it again, so a change to one would reach new DBs only, and every DB made before would
# keep the old schema; see docs/migrations.md. Until then 000001 can still change, as staging starts again from it.
#
#   scripts/check_migrations.sh [base]    e.g. scripts/check_migrations.sh origin/main (the default)
set -euo pipefail

base=${1:-origin/main}
if ! git cat-file -e "$base:data/sql/MIGRATIONS_FROZEN" 2>/dev/null; then
	echo "Migrations are not frozen yet (data/sql/MIGRATIONS_FROZEN is not on $base), so any can change"
	exit 0
fi
changed=$(git diff --name-only --diff-filter=MDR "$base...HEAD" -- 'data/sql/*/migrations/*.sql')
if [ -n "$changed" ]; then
	echo "These migrations are on $base, so DBs have run them, and they cannot change:"
	echo "$changed"
	echo "Add a new migration numbered after the last instead; see docs/migrations.md."
	exit 1
fi
echo "No migration on $base changed"
