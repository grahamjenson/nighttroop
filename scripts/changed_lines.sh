#!/bin/sh
# Counts hand-written lines changed against a base ref, ignoring generated
# files (z_ prefix, _temp.go, .gen.go, _gen.sql, _templ.go, _gen.css).
# Usage: scripts/changed_lines.sh [base-ref] [limit]
base="${1:-origin/main}"
limit="${2:-100}"
total=$(git diff --numstat "$base...HEAD" \
	| grep -Ev '(	|/)z_[^/]*$|_temp\.go$|\.gen\.go$|_gen\.sql$|_templ\.go$|_gen\.css$|^-' \
	| awk '{ s += $1 + $2 } END { print s + 0 }')
echo "Changed lines (excluding generated): $total / $limit"
[ "$total" -le "$limit" ]
