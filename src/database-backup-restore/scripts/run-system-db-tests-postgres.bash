#!/usr/bin/env bash

set -euo pipefail

SRC_DIR="$(cd "$( dirname "$0" )/.." && pwd)"

for i in {1..5}; do
  # Wait for the database to be ready
  PGPASSWORD=${POSTGRES_PASSWORD} psql -U ${POSTGRES_USERNAME} -h ${POSTGRES_HOSTNAME} -p ${POSTGRES_PORT} -c "SELECT CAST('successfully connected' AS text) AS healthcheck" && break || sleep 15
done

pushd "$SRC_DIR"
  go build ./cmd/database-backup-restore
  mv database-backup-restore /usr/local/bin/database-backup-restore

  export TEST_TLS=true
  export TEST_TLS_VERIFY_IDENTITY=false
  export TEST_SSL_USER_REQUIRES_SSL=true


  source scripts/system-db-tests-vars.bash

  export PG_DUMP_9_4_PATH="$(which pg_dump)"
  export PG_RESTORE_9_4_PATH="$(which pg_restore)"

  export PG_DUMP_9_6_PATH="$(which pg_dump)"
  export PG_RESTORE_9_6_PATH="$(which pg_restore)"

  export PG_DUMP_10_PATH="$(which pg_dump)"
  export PG_RESTORE_10_PATH="$(which pg_restore)"

  export PG_DUMP_11_PATH="$(which pg_dump)"
  export PG_RESTORE_11_PATH="$(which pg_restore)"

  export PG_DUMP_13_PATH="$(which pg_dump)"
  export PG_RESTORE_13_PATH="$(which pg_restore)"

  export PG_CLIENT_PATH="$(which psql)"

  ginkgo -mod vendor -r -v "system_tests/postgresql" -trace
popd
