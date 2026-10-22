#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

# shellcheck source-path=SCRIPTDIR source=../shared/local-bosh.sh
source backup-and-restore-sdk-release-ci/ci/tasks/shared/local-bosh.sh

start_local_bosh

upload_stemcell stemcell/*.tgz
upload_release dev-release/*.tgz
upload_release bpm-release/*.tgz

deploy "${SDK_DEPLOYMENT}" "${ci_dir}/manifests/database-backup-restorer.yml" \
  -o "$(release_ops backup-and-restore-sdk)"

# Deploy one at a time: parallel deploys that compile the same release
# against the same stemcell fight over compilation locks.
for version in ${POSTGRES_VERSIONS}; do
  deploy "postgres-${version}" "${ci_dir}/manifests/postgres.yml" \
    -o "$(release_ops bpm)" \
    -v postgres_version="${version}" \
    -v db_username="${POSTGRES_USERNAME}" \
    -v db_password="${POSTGRES_PASSWORD}" \
    -v availability_zone=z1
done

for version in ${POSTGRES_VERSIONS}; do
  run_suite "postgres ${version}" \
    POSTGRES_HOSTNAME="$(instance_ip "postgres-${version}")" \
    TEST_SUITE_NAME=postgresql \
    "${ci_dir}/tasks/sdk-system-db/task.sh"
done

check_suites
