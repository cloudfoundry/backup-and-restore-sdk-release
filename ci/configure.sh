#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
pipeline_config="${REPO_ROOT}/ci/pipeline.yml"

fly validate-pipeline --strict --config "${pipeline_config}"

fly -t "${CONCOURSE_TARGET:-bosh}" \
  set-pipeline -p backup-and-restore-sdk-release \
  -c "${pipeline_config}"
