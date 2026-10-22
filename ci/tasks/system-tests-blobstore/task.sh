#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

# shellcheck source-path=SCRIPTDIR source=../shared/local-bosh.sh
source backup-and-restore-sdk-release-ci/ci/tasks/shared/local-bosh.sh

if [[ -f bosh-env/env.sh ]]; then
  # shellcheck disable=SC1091
  source bosh-env/env.sh
  use_bosh_gateway
else
  start_local_bosh
fi

upload_stemcell stemcell/*.tgz
upload_release dev-release/*.tgz

deploy_vars="$(mktemp)"
echo "${DEPLOY_VARS}" > "${deploy_vars}"

deploy "${BOSH_DEPLOYMENT}" "${ci_dir}/manifests/${MANIFEST}" \
  -o "$(release_ops backup-and-restore-sdk)" \
  -l "${deploy_vars}"

# The suite sets up its own gateway connection from BOSH_GW_*.
unset BOSH_ALL_PROXY
"${ci_dir}/tasks/sdk-system-blobstore/task.sh"
