#!/usr/bin/env bash
#
# Helpers for system-test tasks that deploy to a BOSH director.
#
# start_local_bosh needs a privileged task on ghcr.io/cloudfoundry/bosh/docker-cpi.
# It starts a Docker CPI director inside the task container. That director only
# lives as long as the task, so deploys and tests run in the same task.
#
# Source this file; it doesn't do anything on its own.

# shellcheck disable=SC2034 # used by the tasks that source this
ci_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
release_ops_dir="$(mktemp -d)"
failed_suites=()

start_local_bosh() {
  # start-bosh is written to be sourced, but not under `set -u`.
  set +u
  # shellcheck disable=SC1091
  source start-bosh
  # shellcheck disable=SC1091
  source /tmp/local-bosh/director/bosh-env
  set -u

  # Deployments are containers on the task's docker network, reachable directly.
  unset BOSH_ALL_PROXY BOSH_GW_HOST BOSH_GW_USER BOSH_GW_PRIVATE_KEY
}

# Points BOSH_ALL_PROXY at BOSH_GW_HOST, for directors behind a jumpbox.
# BOSH_GW_PRIVATE_KEY holds the key itself, as elsewhere in this pipeline.
use_bosh_gateway() {
  local key_file
  key_file="$(mktemp)"
  echo "${BOSH_GW_PRIVATE_KEY}" > "${key_file}"
  chmod 0600 "${key_file}"
  export BOSH_ALL_PROXY="ssh+socks5://${BOSH_GW_USER}@${BOSH_GW_HOST}:22?private-key=${key_file}"
}

upload_stemcell() {
  bosh -n upload-stemcell "$1"
}

# Uploads a release tarball and writes ${release_ops_dir}/<name>.yml, an ops
# file that pins the manifest's release to the uploaded version (and drops the
# bosh.io url it might have).
upload_release() {
  local tarball="$1" release_mf name version

  bosh -n upload-release "${tarball}"

  release_mf="$(tar -xzOf "${tarball}" --wildcards '*release.MF')"
  name="$(bosh int --path /name - <<< "${release_mf}")"
  version="$(bosh int --path /version - <<< "${release_mf}")"

  cat > "${release_ops_dir}/${name}.yml" <<EOF
- type: replace
  path: /releases/name=${name}
  value: {name: ${name}, version: "${version}"}
EOF
}

release_ops() {
  echo "${release_ops_dir}/$1.yml"
}

# deploy <deployment> <manifest> [bosh deploy flags...]
deploy() {
  local deployment="$1" manifest="$2"
  shift 2

  bosh -n -d "${deployment}" deploy "${manifest}" \
    -v deployment-name="${deployment}" \
    "$@"
}

# instance_ip <deployment>: the IP of the deployment's only instance.
instance_ip() {
  bosh -d "$1" instances --json | jq -r '.Tables[0].Rows[0].ips'
}

# run_suite <label> [VAR=value...] <command...>
#
# Runs a test suite with extra env vars and records whether it failed, so one
# failing suite doesn't stop the others. Call check_suites at the end.
run_suite() {
  local label="$1"
  shift

  echo "=== ${label}"
  if env "$@"; then
    echo "=== ${label}: passed"
  else
    echo "=== ${label}: FAILED"
    failed_suites+=("${label}")
  fi
}

check_suites() {
  if (( ${#failed_suites[@]} > 0 )); then
    echo "Failed suites:" >&2
    printf '  %s\n' "${failed_suites[@]}" >&2
    return 1
  fi
}
