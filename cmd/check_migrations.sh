#!/usr/bin/env bash
# Fails if a migration added is not numbered after the last on the base branch. A DB only runs the migrations numbered
# after the version it reached, so one numbered at or before it, as when another pull request's migration merged
# first, would never run on DBs already migrated past it.
#
# Also fails if a migration that is on the base branch was changed, deleted or renamed, once the migrations are frozen:
# once data/sql/MIGRATIONS_FROZEN is on the base branch, which it is from the first production deploy. A DB already at a
# migration's version never runs it again, so a change to one would reach new DBs only, and every DB made before would
# keep the old schema; see docs/migrations.md. Until then 000001 can still change, as staging starts again from it.
#
#   cmd/check_migrations.sh [base]    e.g. cmd/check_migrations.sh origin/main (the default)
set -euo pipefail

base=${1:-origin/main}
misordered=""
# The migration numbers in a folder at a commit
numbers() { git ls-tree --name-only "$1" "$2/" | sed -n 's|.*/\([0-9]*\)_.*|\1|p' | sort -u; }
# --no-renames, so a migration renamed to a new number is checked as one added. A number the base branch already has
# is not new: a second migration of it fails go generate, and a change to one fails below once frozen.
for dir in $(git diff --name-only --no-renames --diff-filter=A "$base...HEAD" -- 'data/sql/*/migrations/*.sql' | xargs -r -n1 dirname | sort -u); do
	last=$(numbers "$base" "$dir" | tail -1)
	for number in $(comm -13 <(numbers "$base" "$dir") <(numbers HEAD "$dir")); do
		if [ -n "$last" ] && [ "$((10#$number))" -le "$((10#$last))" ]; then
			misordered+="$dir/${number}_* (the last on $base is $last)"$'\n'
		fi
	done
done
if [ -n "$misordered" ]; then
	echo "These migrations are not numbered after the last on $base, so DBs already past that number would never run them:"
	printf '%s' "$misordered"
	echo "Number them after the last on $base; see docs/migrations.md."
	exit 1
fi

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
