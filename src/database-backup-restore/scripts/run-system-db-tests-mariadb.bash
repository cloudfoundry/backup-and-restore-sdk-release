#!/usr/bin/env bash

set -euo pipefail

SRC_DIR="$(cd "$( dirname "$0" )/.." && pwd)"

pushd "$SRC_DIR"
  source scripts/system-db-tests-vars.bash

  for i in {1..5}; do
    # Wait for the database to be ready
    ${MYSQL_CLIENT_5_7_PATH} -u ${MYSQL_USERNAME} -p${MYSQL_PASSWORD} -h ${MYSQL_HOSTNAME} -P ${MYSQL_PORT} -e 'SELECT "successfully connected to mysql"' && break || sleep 15
  done

  go build ./cmd/database-backup-restore
  mv database-backup-restore /usr/local/bin/database-backup-restore

  export MYSQL_CA_CERT_PATH="/tls-certs/ca-cert.pem"
  export MYSQL_CLIENT_CERT_PATH="/tls-certs/client-cert.pem"
  export MYSQL_CLIENT_KEY_PATH="/tls-certs/client-key.pem"

  export MYSQL_CA_CERT="$( cat "${MYSQL_CA_CERT_PATH}" )"
  export MYSQL_CLIENT_CERT="$( cat "${MYSQL_CLIENT_CERT_PATH}" )"
  export MYSQL_CLIENT_KEY="$( cat "${MYSQL_CLIENT_KEY_PATH}" )"

  export TEST_TLS=true
  export TEST_TLS_MUTUAL_TLS=false
  export TEST_TLS_VERIFY_IDENTITY=false
  export TEST_SSL_USER_REQUIRES_SSL=true

  ginkgo -mod vendor -r -v "system_tests/mysql" -trace
popd
