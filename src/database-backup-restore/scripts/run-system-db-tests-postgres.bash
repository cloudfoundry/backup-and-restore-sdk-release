#!/usr/bin/env bash

set -euo pipefail

ENABLE_TLS="${ENABLE_TLS:-no}"

SRC_DIR="$(cd "$( dirname "$0" )/.." && pwd)"

pushd "$SRC_DIR"
  PG_BINARY="/var/vcap/packages/database-backup-restorer-postgres-13/bin/psql"

  if [[ "$ENABLE_TLS" == "no" ]]; then

  for i in {1..5}; do
     # Wait for the database to be ready
     PGPASSWORD=${POSTGRES_PASSWORD} ${PG_BINARY} -U ${POSTGRES_USERNAME} -h ${POSTGRES_HOSTNAME} -p ${POSTGRES_PORT} -c "SELECT CAST('successfully connected' AS text) AS healthcheck" && break || sleep 15
  done


  export TEST_TLS=true
  export TEST_TLS_VERIFY_IDENTITY=false
  export TEST_SSL_USER_REQUIRES_SSL=true

  ginkgo -mod vendor -r -v "system_tests/postgresql" -trace

  elif [[ "$ENABLE_TLS" == "yes" ]]; then

  export TEST_TLS=true
  export TEST_TLS_VERIFY_IDENTITY=false
  export TEST_SSL_USER_REQUIRES_SSL=false

  export POSTGRES_CA_CERT_PATH="/tls-certs/ca-cert.pem"
  export POSTGRES_CLIENT_CERT_PATH="/tls-certs/client-cert.pem"
  export POSTGRES_CLIENT_KEY_PATH="/tls-certs/client-key.pem"

  for i in {1..5}; do
     # Wait for the database to be ready
     PGPASSWORD=${POSTGRES_PASSWORD} ${PG_BINARY} -U ${POSTGRES_USERNAME} -h ${POSTGRES_HOSTNAME} -p ${POSTGRES_PORT} -c "SELECT CAST('successfully connected' AS text) AS healthcheck" && break || sleep 15
  done

  export POSTGRES_CA_CERT="$( cat "${POSTGRES_CA_CERT_PATH}" )"
  export POSTGRES_CLIENT_CERT="$( cat "${POSTGRES_CLIENT_CERT_PATH}" )"
  export POSTGRES_CLIENT_KEY="$( cat "${POSTGRES_CLIENT_KEY_PATH}" )"

  ginkgo -mod vendor -r -v "system_tests/postgresql_tls" -trace

  elif [[ "${ENABLE_TLS}" == "mutual" ]]; then

  export TEST_TLS=true
  export TEST_TLS_VERIFY_IDENTITY=false
  export TEST_SSL_USER_REQUIRES_SSL=false

  export POSTGRES_CA_CERT_PATH="/tls-certs/ca-cert.pem"
  export POSTGRES_CLIENT_CERT_PATH="/tls-certs/client-cert.pem"
  export POSTGRES_CLIENT_KEY_PATH="/tls-certs/client-key.pem"

  for i in {1..5}; do
     # Wait for the database to be ready
     PGSSLMODE="verify-full"                  \
     PGREQUIRESSL=1                           \
     PGPASSWORD="${POSTGRES_PASSWORD}"        \
     PGSSLROOTCERT="${POSTGRES_CA_CERT_PATH}" \
     PGSSLKEY="${POSTGRES_CLIENT_KEY_PATH}"   \
     PGSSLCERT="${POSTGRES_CLIENT_CERT_PATH}" \
     ${PG_BINARY} -U ${POSTGRES_USERNAME} -h ${POSTGRES_HOSTNAME} -p ${POSTGRES_PORT} -c "SELECT CAST('successfully connected' AS text) AS healthcheck" && break || sleep 15
  done

  export POSTGRES_CA_CERT="$( cat "${POSTGRES_CA_CERT_PATH}" )"
  export POSTGRES_CLIENT_CERT="$( cat "${POSTGRES_CLIENT_CERT_PATH}" )"
  export POSTGRES_CLIENT_KEY="$( cat "${POSTGRES_CLIENT_KEY_PATH}" )"

  ginkgo -mod vendor -r -v "system_tests/postgresql_mutual_tls" -trace

  fi
popd
