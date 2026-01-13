#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

fly -t "${CONCOURSE_TARGET:-bosh-ecosystem}" \
  set-pipeline -p backup-and-restore-sdk-release \
  -c ci/pipeline.yml
