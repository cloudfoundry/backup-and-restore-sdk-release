#!/usr/bin/env bash

set -euo pipefail

pushd "/goproject"
go build ./cmd/database-backup-restore
mv database-backup-restore /database-backup-restore
popd

for i in {1..5}; do
/mysql -u ${MYSQL_USERNAME} -p${MYSQL_PASSWORD} -h ${MYSQL_HOSTNAME} -P ${MYSQL_PORT} -e 'SELECT "successfully connected to mysql"' && break || sleep 15
done

if [ ! -z "$MYSQL_CA_CERT_PATH" ]; then
export MYSQL_CA_CERT="$( cat "${MYSQL_CA_CERT_PATH}" )"
fi

if [ ! -z "$MYSQL_CLIENT_CERT_PATH" ]; then
export MYSQL_CLIENT_CERT="$( cat "${MYSQL_CLIENT_CERT_PATH}" )"
fi

if [ ! -z "$MYSQL_CLIENT_KEY_PATH" ]; then
export MYSQL_CLIENT_KEY="$( cat "${MYSQL_CLIENT_KEY_PATH}" )"
fi

/goproject/scripts/run-system-tests.bash
