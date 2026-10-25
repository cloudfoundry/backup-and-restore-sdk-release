#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

# shellcheck source-path=SCRIPTDIR source=../shared/local-bosh.sh
source backup-and-restore-sdk-release-ci/ci/tasks/shared/local-bosh.sh

start_local_bosh

upload_stemcell stemcell/*.tgz
upload_stemcell mysql-stemcell/*.tgz
upload_release dev-release/*.tgz

compiled_releases="$(mktemp)"
yq '[.[] | select(.path == "/releases/name=pxc" or .path == "/releases/name=bpm")]' \
  cf-deployment/operations/use-compiled-releases.yml > "${compiled_releases}"
if [[ "$(yq 'length' "${compiled_releases}")" != 2 ]]; then
  echo "cf-deployment's use-compiled-releases.yml is missing pxc or bpm" >&2
  exit 1
fi

deploy "${SDK_DEPLOYMENT}" "${ci_dir}/manifests/database-backup-restorer.yml" \
  -o "$(release_ops backup-and-restore-sdk)"

# The mysql manifest's `databases` TLS certificate is generated here, so the
# tests can be given its CA. Both servers share it.
vars_store="$(mktemp -d)/mysql-vars.yml"

# Deploy one at a time: the first deploy uploads the compiled releases, and
# the second reuses them.
#
# -v values are parsed as YAML, so the version is quoted to keep 8.0 from
# becoming 8.
for version in ${MYSQL_VERSIONS}; do
  deploy "mysql-${version//./-}" "${ci_dir}/manifests/mysql.yml" \
    -o "${compiled_releases}" \
    --vars-store "${vars_store}" \
    -v mysql_version="'${version}'" \
    -v db_password="${MYSQL_PASSWORD}" \
    -v availability_zone=z1
done

for version in ${MYSQL_VERSIONS}; do
  run_suite "mysql ${version}" \
    MYSQL_HOSTNAME="$(instance_ip "mysql-${version//./-}")" \
    MYSQL_CA_CERT="$(bosh int "${vars_store}" --path /databases/ca)" \
    MYSQL_CLIENT_CERT="$(bosh int "${vars_store}" --path /databases/certificate)" \
    MYSQL_CLIENT_KEY="$(bosh int "${vars_store}" --path /databases/private_key)" \
    TEST_SUITE_NAME=mysql \
    "${ci_dir}/tasks/sdk-system-db/task.sh"
done

check_suites
