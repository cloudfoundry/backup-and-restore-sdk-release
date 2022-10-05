#!/usr/bin/env bash

set -euo pipefail

SRC_DIR="$(cd "$( dirname "$0" )/.." && pwd)"

for i in {1..5}; do
  # Wait for the database to be ready
  mysql -u ${MYSQL_USERNAME} -p${MYSQL_PASSWORD} -h ${MYSQL_HOSTNAME} -P ${MYSQL_PORT} -e 'SELECT "successfully connected to mysql"' && break || sleep 15
done

pushd "$SRC_DIR"
  go build ./cmd/database-backup-restore
  mv database-backup-restore /usr/local/bin/database-backup-restore

  export MYSQL_CA_CERT_PATH="${MYSQL_CERTS_PATH}/ca.pem"
  export MYSQL_CLIENT_CERT_PATH="${MYSQL_CERTS_PATH}/server-cert.pem"
  export MYSQL_CLIENT_KEY_PATH="${MYSQL_CERTS_PATH}/server-key.pem"

  export MYSQL_CA_CERT="$( cat "${MYSQL_CA_CERT_PATH}" )"
  export MYSQL_CLIENT_CERT="$( cat "${MYSQL_CLIENT_CERT_PATH}" )"
  export MYSQL_CLIENT_KEY="$( cat "${MYSQL_CLIENT_KEY_PATH}" )"

  export TEST_TLS=true
  export TEST_TLS_VERIFY_IDENTITY=false
  export TEST_SSL_USER_REQUIRES_SSL=true


  source scripts/system-db-tests-vars.bash

  export MYSQL_DUMP_5_7_PATH="$(which mysqldump)"
  export MYSQL_CLIENT_5_7_PATH="$(which mysql)"

  export MYSQL_DUMP_8_0_PATH="$(which mysqldump)"
  export MYSQL_CLIENT_8_0_PATH="$(which mysql)"

  ginkgo -mod vendor -r -v "system_tests/mysql" -trace
popd
