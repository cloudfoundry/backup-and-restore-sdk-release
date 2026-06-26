#!/usr/bin/env bash
set -eu -o pipefail

if [[ -n "${DEBUG:-}" ]]; then
  set -x
fi

REPO_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"

concourse_target="${CONCOURSE_TARGET:-bosh}"
fly="${FLY_CLI:-fly}"

pipeline_name="backup-and-restore-sdk-release"
pipeline_config="${REPO_ROOT}/ci/pipeline.yml"

echo "Validating..."
"${fly}" validate-pipeline --config "${pipeline_config}" # TODO: add back '--strict'
echo ""

"${fly}" -t "${concourse_target}" \
  set-pipeline \
    --pipeline "${pipeline_name}" \
    --config "${pipeline_config}"
