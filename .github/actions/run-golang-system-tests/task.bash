#!/usr/bin/env bash

set -euo pipefail
set -x

for i in {1..5}; do
/mysql -u ${MYSQL_USERNAME} -p${MYSQL_PASSWORD} -h ${MYSQL_HOSTNAME} -P ${MYSQL_PORT} -e "SELECT 999" && break || sleep 15
done
/goproject/scripts/run-system-tests.bash
