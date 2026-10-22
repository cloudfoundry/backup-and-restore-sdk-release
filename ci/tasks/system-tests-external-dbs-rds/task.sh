#!/usr/bin/env bash
set -euo pipefail

[ -z "${DEBUG:-}" ] || set -x

# shellcheck source-path=SCRIPTDIR source=../shared/local-bosh.sh
source backup-and-restore-sdk-release-ci/ci/tasks/shared/local-bosh.sh

aws_config="$(mktemp)"
cat > "${aws_config}" <<EOF
[profile pipeline]
role_arn = ${AWS_ROLE_ARN}
credential_source = Environment
region = ${AWS_REGION}
EOF

aws_pipeline() {
  AWS_CONFIG_FILE="${aws_config}" aws --profile pipeline "$@"
}

# The RDS security group has no ingress rules of its own. Everything this
# task and its containers send leaves through the worker's outbound IP, so
# allow just that, for as long as the task runs.
egress_ip="$(curl -fsS https://checkip.amazonaws.com)"
ip_permissions="$(cat <<EOF
[
  {"IpProtocol": "tcp", "FromPort": 5432, "ToPort": 5432, "IpRanges": [{"CidrIp": "${egress_ip}/32"}]},
  {"IpProtocol": "tcp", "FromPort": 3306, "ToPort": 3306, "IpRanges": [{"CidrIp": "${egress_ip}/32"}]}
]
EOF
)"

revoke_ingress() {
  aws_pipeline ec2 revoke-security-group-ingress \
    --group-id "${RDS_SECURITY_GROUP_ID}" \
    --ip-permissions "${ip_permissions}" > /dev/null \
    || echo "WARNING: could not revoke ingress for ${egress_ip}; terraform destroy removes the security group anyway" >&2
}

echo "Allowing ${egress_ip} into ${RDS_SECURITY_GROUP_ID}"
aws_pipeline ec2 authorize-security-group-ingress \
  --group-id "${RDS_SECURITY_GROUP_ID}" \
  --ip-permissions "${ip_permissions}" > /dev/null
trap revoke_ingress EXIT

# The new rule can take a few seconds to apply, so retry each endpoint.
wait_for_endpoint() {
  local label="$1" host="$2" port="$3"
  for _ in 1 2 3 4 5 6; do
    if nc -zw5 "${host}" "${port}"; then
      echo "${label} (${host}:${port}) is reachable"
      return 0
    fi
    sleep 5
  done
  echo "ERROR: can't reach ${label} (${host}:${port}). Check the RDS security group." >&2
  return 1
}

wait_for_endpoint "Postgres 15" "${POSTGRES_15_HOST}" "${POSTGRES_PORT}"
wait_for_endpoint "Postgres 16" "${POSTGRES_16_HOST}" "${POSTGRES_PORT}"
wait_for_endpoint "MariaDB 10.6" "${MARIADB_10_6_HOST}" "${MYSQL_PORT}"

rds_ca_cert="$(curl -fsSL "https://truststore.pki.rds.amazonaws.com/${AWS_REGION}/${AWS_REGION}-bundle.pem")"

start_local_bosh

upload_stemcell stemcell/*.tgz
upload_release dev-release/*.tgz

deploy "${SDK_DEPLOYMENT}" "${ci_dir}/manifests/database-backup-restorer.yml" \
  -o "$(release_ops backup-and-restore-sdk)"

db_suite="${ci_dir}/tasks/sdk-system-db/task.sh"

run_suite "postgres 15 tls" \
  POSTGRES_HOSTNAME="${POSTGRES_15_HOST}" \
  POSTGRES_PASSWORD="${POSTGRES_15_PASSWORD}" \
  POSTGRES_CA_CERT="${rds_ca_cert}" \
  TEST_SUITE_NAME=postgresql_tls \
  "${db_suite}"

run_suite "postgres 16" \
  POSTGRES_HOSTNAME="${POSTGRES_16_HOST}" \
  POSTGRES_PASSWORD="${POSTGRES_16_PASSWORD}" \
  TEST_SUITE_NAME=postgresql \
  "${db_suite}"

run_suite "postgres 16 tls" \
  POSTGRES_HOSTNAME="${POSTGRES_16_HOST}" \
  POSTGRES_PASSWORD="${POSTGRES_16_PASSWORD}" \
  POSTGRES_CA_CERT="${rds_ca_cert}" \
  TEST_SUITE_NAME=postgresql_tls \
  "${db_suite}"

run_suite "mariadb 10.6" \
  MYSQL_HOSTNAME="${MARIADB_10_6_HOST}" \
  MYSQL_PASSWORD="${MARIADB_10_6_PASSWORD}" \
  MYSQL_CA_CERT="${rds_ca_cert}" \
  TEST_TLS_MUTUAL_TLS=false \
  TEST_TLS_VERIFY_IDENTITY=false \
  TEST_SUITE_NAME=mysql \
  "${db_suite}"

check_suites
