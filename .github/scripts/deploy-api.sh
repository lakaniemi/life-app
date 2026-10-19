#!/usr/bin/env bash
# Asks the server to deploy whatever :production points to, then waits until
# /health reports the expected commit.
#
# Usage: DEPLOY_WEBHOOK_TOKEN=… deploy-api.sh <expected-sha>
#
# The hook only means "check :production now": the server picks the image,
# runs migrations and swaps the API on its own, so the token can't choose
# what's deployed. It's called once; a failed deploy fails the run, which is
# re-run by hand.
set -euo pipefail

hook_url="https://webhooks.lakaniemi.com/deploy"
health_url="https://life-app-api.lakaniemi.com/health"
timeout_seconds=300
poll_interval_seconds=5

expected="${1:?usage: deploy-api.sh <expected-sha>}"
: "${DEPLOY_WEBHOOK_TOKEN:?DEPLOY_WEBHOOK_TOKEN is not set}"

# -f is left out so a 401 can be told apart from an unreachable hook.
status=$(curl -sS -o /dev/null -w '%{http_code}' --max-time 30 -X POST \
  -H "Authorization: Bearer $DEPLOY_WEBHOOK_TOKEN" "$hook_url") || status=000
case "$status" in
  2??) echo "Deploy hook accepted ($status)" ;;
  401)
    echo "::error::Deploy hook rejected the token (401). Check the DEPLOY_WEBHOOK_TOKEN secret."
    exit 1
    ;;
  *)
    echo "::error::Deploy hook unreachable or failed (HTTP $status)."
    exit 1
    ;;
esac

deadline=$((SECONDS + timeout_seconds))
while ((SECONDS < deadline)); do
  # Connection errors, 5xx and unparsable bodies all mean "not yet".
  version=$(curl -fsS --max-time 10 "$health_url" 2>/dev/null | jq -r '.version // empty' 2>/dev/null) || version=""
  if [[ "$version" == "$expected" ]]; then
    echo "Deployed: /health reports $version"
    exit 0
  fi
  echo "Waiting: /health reports ${version:-nothing}, want $expected"
  sleep "$poll_interval_seconds"
done

echo "::error::/health didn't report $expected within ${timeout_seconds}s. Check the server's deploy logs, then re-run this job."
exit 1
