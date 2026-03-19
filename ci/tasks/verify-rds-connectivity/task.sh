#!/usr/bin/env bash
set -euo pipefail

ssh_key="$(mktemp)"
echo -e "${BOSH_GW_PRIVATE_KEY}" > "$ssh_key"
chmod 0400 "$ssh_key"

ssh_opts=(
  -o StrictHostKeyChecking=no
  -o UserKnownHostsFile=/dev/null
  -o ConnectTimeout=10
  -o LogLevel=ERROR
  -i "$ssh_key"
)

check_connectivity() {
  local label="$1" host="$2" port="$3"

  echo -n "Checking ${label} (${host}:${port})... "
  if ssh "${ssh_opts[@]}" "${BOSH_GW_USER}@${BOSH_GW_HOST}" "nc -zw5 ${host} ${port}" 2>&1; then
    echo "OK"
  else
    echo "FAILED"
    return 1
  fi
}

failed=0

check_connectivity "Postgres 13" "$POSTGRES_13_HOST" "$POSTGRES_PORT" || failed=1
check_connectivity "Postgres 15" "$POSTGRES_15_HOST" "$POSTGRES_PORT" || failed=1
check_connectivity "Postgres 16" "$POSTGRES_16_HOST" "$POSTGRES_PORT" || failed=1
check_connectivity "MariaDB 10.6" "$MARIADB_10_6_HOST" "$MARIADB_PORT" || failed=1

if [ "$failed" -ne 0 ]; then
  echo ""
  echo "ERROR: One or more RDS endpoints are not reachable from the jumpbox."
  echo "Check that the RDS security groups allow inbound traffic on the database ports."
  exit 1
fi

echo ""
echo "All RDS endpoints are reachable."
