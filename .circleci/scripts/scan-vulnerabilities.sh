#! /bin/bash
# Scan the insights-cli image at tip SHA, then latest git tag.
set -eo pipefail

repo="us-docker.pkg.dev/fairwinds-ops/oss/insights-cli"
sha="${CIRCLE_SHA1:-}"
rel=""
git fetch --tags --force --depth=1 origin >/dev/null 2>&1 || true
if git describe --tags --abbrev=0 >/dev/null 2>&1; then
  rel=$(git describe --tags --abbrev=0)
fi

if [[ -n "${GCP_ARTIFACTREADWRITE_JSON_KEY:-}" ]]; then
  echo "$GCP_ARTIFACTREADWRITE_JSON_KEY" | base64 -d | docker login -u _json_key --password-stdin us-docker.pkg.dev || true
fi

pulled=""
for cand in "${repo}:${sha}" "${repo}:${rel}"; do
  [[ -n "${cand##*:}" ]] || continue
  echo "Trying $cand"
  if docker pull "$cand"; then
    pulled=$cand
    break
  fi
done

if [[ -z "$pulled" ]]; then
  echo "No insights-cli image for sha=${sha:-none} tag=${rel:-none}" >&2
  exit 1
fi

echo "Scanning $pulled"
set +e
trivy image --exit-code 123 --severity CRITICAL "$pulled"
code=$?
set -e
if [[ $code -eq 123 ]]; then
  list="- ${pulled}"
  printf '%s\n' "$list" > .scan-vulns-list
  printf 'export CRITICAL_VULNERABILITIES_LIST=%s\n' "$(printf '%s' "$list" | jq -Rs .)" >> "${BASH_ENV:-/dev/null}"
  exit 1
fi
if [[ $code -ne 0 ]]; then
  exit "$code"
fi
echo "Image scan found no CRITICAL vulnerabilities."
