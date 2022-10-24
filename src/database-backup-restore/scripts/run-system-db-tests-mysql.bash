#!/usr/bin/env bash

set -euo pipefail

SRC_DIR="$(cd "$( dirname "$0" )/.." && pwd)"

pushd "$SRC_DIR"
  MYSQL_BINARY="/var/vcap/packages/database-backup-restorer-mysql-5.7/bin/mysql"

  for i in {1..5}; do
    # Wait for the database to be ready
    ${MYSQL_BINARY} -u ${MYSQL_USERNAME} -p${MYSQL_PASSWORD} -h ${MYSQL_HOSTNAME} -P ${MYSQL_PORT} -e 'SELECT "successfully connected to mysql"' && break || sleep 15
  done

  export MYSQL_CA_CERT_PATH="/tls-certs/ca.pem"
  export MYSQL_CLIENT_CERT_PATH="/tls-certs/server-cert.pem"
  export MYSQL_CLIENT_KEY_PATH="/tls-certs/server-key.pem"

  export MYSQL_CA_CERT="$( cat "${MYSQL_CA_CERT_PATH}" )"
  export MYSQL_CLIENT_CERT="$( cat "${MYSQL_CLIENT_CERT_PATH}" )"
  export MYSQL_CLIENT_KEY="$( cat "${MYSQL_CLIENT_KEY_PATH}" )"

  export TEST_TLS=true
  export TEST_TLS_VERIFY_IDENTITY=false
  export TEST_SSL_USER_REQUIRES_SSL=true

  export RUN_TESTS_WITHOUT_BOSH=true
  ginkgo -mod vendor -r -v "system_tests/mysql" -trace
popd
