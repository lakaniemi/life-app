#!/usr/bin/env bash
# Compares app/version.json with the latest app release (the newest app-v*
# tag) and reports whether the current commit is a new release.
#
# Usage: check-app-version.sh
#
# Prints release=true or release=false, to $GITHUB_OUTPUT when it is set. Fails
# when the file is malformed or the bump is inconsistent, so a bad bump fails
# on the pull request instead of after the merge. Needs the tags, so check out
# with fetch-depth: 0.
#
#   same version, same versionCode       -> release=false
#   higher version and higher versionCode -> release=true
#   anything else                         -> error
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

version=$(jq -r '.version' app/version.json)
code=$(jq -r '.versionCode' app/version.json)

if ! jq -e '(.version | type == "string") and (.versionCode | type == "number")' app/version.json > /dev/null; then
  echo "::error file=app/version.json::version must be a JSON string and versionCode a JSON number, because latest.json copies them as they are."
  exit 1
fi
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "::error file=app/version.json::version must look like 1.2.3, got '$version'."
  exit 1
fi
if [[ ! "$code" =~ ^[1-9][0-9]*$ ]]; then
  echo "::error file=app/version.json::versionCode must be a positive integer, got '$code'."
  exit 1
fi

report() {
  echo "release=$1" >> "${GITHUB_OUTPUT:-/dev/stdout}"
}

latest_tag=$(git tag --list 'app-v*' --sort=-v:refname | head -n 1)
if [[ -z "$latest_tag" ]]; then
  echo "No app release yet, so $version ($code) will be the first."
  report true
  exit 0
fi

previous_version="${latest_tag#app-v}"
previous_code=$(git show "$latest_tag:app/version.json" | jq -r '.versionCode')

if [[ "$version" == "$previous_version" ]]; then
  if [[ "$code" != "$previous_code" ]]; then
    echo "::error file=app/version.json::versionCode changed to $code but version is still $version. Bump both, or neither."
    exit 1
  fi
  echo "$version is already released ($latest_tag)."
  report false
  exit 0
fi

highest=$(printf '%s\n%s\n' "$previous_version" "$version" | sort -V | tail -n 1)
if [[ "$highest" != "$version" ]]; then
  echo "::error file=app/version.json::version $version is lower than the latest release $previous_version."
  exit 1
fi
if ((code <= previous_code)); then
  echo "::error file=app/version.json::versionCode $code must be higher than the latest release's $previous_code. Android refuses to install a lower one over a higher one."
  exit 1
fi

echo "$version ($code) is newer than $latest_tag, so it will be released."
report true
