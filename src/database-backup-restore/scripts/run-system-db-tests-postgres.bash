#!/usr/bin/env bash

set -euo pipefail

SRC_DIR="$(cd "$( dirname "$0" )/.." && pwd)"

pushd "$SRC_DIR"
  PG_BINARY="/var/vcap/packages/database-backup-restorer-postgres-13/bin/psql"

  for i in {1..5}; do
    # Wait for the database to be ready
    PGPASSWORD=${POSTGRES_PASSWORD} ${PG_BINARY} -U ${POSTGRES_USERNAME} -h ${POSTGRES_HOSTNAME} -p ${POSTGRES_PORT} -c "SELECT CAST('successfully connected' AS text) AS healthcheck" && break || sleep 15
  done

  export TEST_TLS=true
  export TEST_TLS_VERIFY_IDENTITY=false
  export TEST_SSL_USER_REQUIRES_SSL=true

  ginkgo -mod vendor -r -v "system_tests/postgresql" -trace
popd
