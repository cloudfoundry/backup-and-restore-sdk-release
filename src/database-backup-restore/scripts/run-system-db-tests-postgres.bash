#!/usr/bin/env bash

set -euo pipefail

SRC_DIR="$(cd "$( dirname "$0" )/.." && pwd)"

pushd "$SRC_DIR"
  source scripts/system-db-tests-vars.bash

  for i in {1..5}; do
    # Wait for the database to be ready
    PGPASSWORD=${POSTGRES_PASSWORD} ${PG_CLIENT_PATH} -U ${POSTGRES_USERNAME} -h ${POSTGRES_HOSTNAME} -p ${POSTGRES_PORT} -c "SELECT CAST('successfully connected' AS text) AS healthcheck" && break || sleep 15
  done

  go build ./cmd/database-backup-restore
  mv database-backup-restore /usr/local/bin/database-backup-restore

  export TEST_TLS=true
  export TEST_TLS_VERIFY_IDENTITY=false
  export TEST_SSL_USER_REQUIRES_SSL=true

  ginkgo -mod vendor -r -v "system_tests/postgresql" -trace
popd
