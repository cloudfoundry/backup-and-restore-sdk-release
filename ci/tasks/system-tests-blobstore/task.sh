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

deploy_args=()
# The job wants the key as a JSON string. In DEPLOY_VARS, or with -v, bosh
# would parse it into a hash, so pass it as a file.
if [[ -n "${GCP_SERVICE_ACCOUNT_KEY:-}" ]]; then
  gcp_key="$(mktemp)"
  echo "${GCP_SERVICE_ACCOUNT_KEY}" > "${gcp_key}"
  deploy_args+=(--var-file gcp-service-account-key="${gcp_key}")
fi

deploy "${BOSH_DEPLOYMENT}" "${ci_dir}/manifests/${MANIFEST}" \
  -o "$(release_ops backup-and-restore-sdk)" \
  -l "${deploy_vars}" \
  "${deploy_args[@]}"

# The suite sets up its own gateway connection from BOSH_GW_*.
unset BOSH_ALL_PROXY
"${ci_dir}/tasks/sdk-system-blobstore/task.sh"
