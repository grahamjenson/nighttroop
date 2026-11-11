#!/bin/bash
# Shows a pull request the screenshots its description lists, once its tests pass. Its description ends with a
# "## Screenshots" section listing them one a line, as "- site/summaries/summaries": a .bdd file's path in tests/bdd
# without .bdd, then the name its SCREENSHOT step gave. Each listed one the tests saved in .test/savedScreenshots is
# pushed to the pr-<number> branch of the screenshots repo, in place of what it had, and one comment on the pull
# request shows them, edited on each run. A description without the section is left alone.
#
# The description is anyone's text, so it is only read from PR_BODY, and only names of screenshots the tests saved are
# pushed. DRY_RUN=1 prints the comment instead of pushing and commenting.
#
# Env: PR_BODY, PR_NUMBER, HEAD_SHA, GITHUB_REPOSITORY, GH_TOKEN (commenting), SCREENSHOTS_TOKEN (pushing to the
# screenshots repo), SCREENSHOTS_REPO (default grahamjenson/nighttroop-screenshots), SCREENSHOTS_REMOTE (its git URL)
set -euo pipefail

saved=.test/savedScreenshots
screenshots_repo=${SCREENSHOTS_REPO:-grahamjenson/nighttroop-screenshots}
screenshots_remote=${SCREENSHOTS_REMOTE:-https://github.com/$screenshots_repo.git}
marker='<!-- nighttroop-screenshots -->'
max=40

# The section's list items: the lines after "## Screenshots" until the next heading, without bullets, backticks or .png
listed=$(printf '%s\n' "${PR_BODY:-}" | tr -d '\r' | awk '
	/^##[[:space:]]+Screenshots[[:space:]]*$/ { inside = 1; found = 1; next }
	inside && /^#/ { inside = 0 }
	inside && /^[[:space:]]*[-*][[:space:]]+/ { sub(/^[[:space:]]*[-*][[:space:]]+/, ""); gsub(/`/, ""); sub(/[[:space:]]+$/, ""); sub(/\.png$/, ""); print }
	END { if (!found) exit 3 }') || {
	[ $? -eq 3 ] && { echo "The description has no ## Screenshots section"; exit 0; }
	exit 1
}

found=() missing=()
declare -A seen
while IFS= read -r name; do
	case $name in "" | None | none) continue ;; esac
	[ -n "${seen[$name]:-}" ] && continue
	seen[$name]=1
	if [ ${#found[@]} -ge $max ]; then
		echo "::warning::Only the first $max screenshots listed are shown"
		break
	fi
	# A name is a path of plain segments, so it can only be a file the tests saved
	if [[ $name =~ ^[A-Za-z0-9_-]+(/[A-Za-z0-9_.-]+)*$ && $name != *..* && -f $saved/$name.png ]]; then
		found+=("$name")
	else
		missing+=("$name")
	fi
done <<<"$listed"

commit=""
if [ ${#found[@]} -gt 0 ] && [ -z "${DRY_RUN:-}" ]; then
	if [ -z "${SCREENSHOTS_TOKEN:-}" ]; then
		echo "::warning::SCREENSHOTS_TOKEN is not set, so the screenshots cannot be pushed to $screenshots_repo"
		exit 0
	fi
	# A new branch each run, with only this run's screenshots, so the screenshots repo keeps no old ones
	work=$(mktemp -d)
	for name in "${found[@]}"; do
		mkdir -p "$work/$(dirname "$name")"
		cp "$saved/$name.png" "$work/$name.png"
	done
	git -C "$work" init -q
	git -C "$work" checkout -q --orphan "pr-$PR_NUMBER"
	git -C "$work" add -A
	git -C "$work" -c user.name="nighttroop CI" -c user.email="ci@nighttroop.invalid" commit -q \
		-m "Screenshots of $GITHUB_REPOSITORY#$PR_NUMBER at $HEAD_SHA"
	auth=$(printf 'x-access-token:%s' "$SCREENSHOTS_TOKEN" | base64 -w0)
	git -C "$work" -c "http.https://github.com/.extraheader=AUTHORIZATION: basic $auth" push -q -f \
		"$screenshots_remote" "HEAD:refs/heads/pr-$PR_NUMBER"
	commit=$(git -C "$work" rev-parse HEAD)
fi

body=$(mktemp)
{
	echo "$marker"
	echo "### Screenshots"
	echo
	echo "The screenshots listed in the description's \`## Screenshots\`, taken by the tests of ${HEAD_SHA:0:7}."
	[ ${#found[@]} -eq 0 ] && [ ${#missing[@]} -eq 0 ] && { echo; echo "None listed."; }
	for name in "${found[@]}"; do
		echo
		echo "**$name**"
		echo
		# By the commit, not the branch, as a branch's raw files are cached for minutes
		echo "![$name](https://raw.githubusercontent.com/$screenshots_repo/${commit:-COMMIT}/$name.png)"
	done
	if [ ${#missing[@]} -gt 0 ]; then
		echo
		echo "Listed, but the tests took no screenshot of that name:"
		for name in "${missing[@]}"; do echo "- \`${name//\`/}\`"; done
	fi
} >"$body"

if [ -n "${DRY_RUN:-}" ]; then
	cat "$body"
	exit 0
fi

# The comment this has made before, edited rather than added to
comments="repos/$GITHUB_REPOSITORY/issues/$PR_NUMBER/comments"
existing=$(gh api --paginate "$comments" \
	--jq ".[] | select(.user.login == \"github-actions[bot]\" and (.body | startswith(\"$marker\"))) | .id" | head -n1)
if [ -n "$existing" ]; then
	gh api -X PATCH "repos/$GITHUB_REPOSITORY/issues/comments/$existing" -F "body=@$body" >/dev/null
else
	gh api -X POST "$comments" -F "body=@$body" >/dev/null
fi
echo "Showed ${#found[@]} screenshots on $GITHUB_REPOSITORY#$PR_NUMBER, ${#missing[@]} listed not found"
