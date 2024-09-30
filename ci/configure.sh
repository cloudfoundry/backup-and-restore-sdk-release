#!/bin/bash

set -euo pipefail

fly -t "${CONCOURSE_TARGET:-bosh-ecosystem}" \
  set-pipeline -p backup-and-restore-sdk-release \
  -c ci/pipelines/backup-and-restore-sdk-release/pipeline.yml
