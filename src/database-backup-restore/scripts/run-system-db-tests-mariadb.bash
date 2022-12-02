#!/usr/bin/env bash

set -euo pipefail

BOSH_LOG_LEVEL=none
SRC_DIR="$(cd "$( dirname "$0" )/.." && pwd)"

source "${BOSH_CREDS_SCRIPT}"

STEMCELL_URL="$(curl -L https://bosh.io/stemcells | grep -io "https:\/\/.*warden-boshlite-${STEMCELL_NAME}-go_agent.tgz")"
bosh -n --tty upload-stemcell "${STEMCELL_URL}"

export SDK_DEPLOYMENT="database-backup-restorer-$(head /dev/urandom | md5sum | cut -f1 -d" ")"
export SDK_INSTANCE_GROUP=database-backup-restorer

cat > /tmp/manifest.yml <<EOF
---
name: "$SDK_DEPLOYMENT"
releases:
- name: backup-and-restore-sdk
  version: 'latest'
instance_groups:
- name: database-backup-restorer
  instances: 1
  vm_type: default
  persistent_disk_type: default
  stemcell: default
  networks:
  - name: default
  jobs:
  - name: database-backup-restorer
    release: backup-and-restore-sdk
  azs: [z1]
stemcells:
- alias: default
  os: "$STEMCELL_NAME"
  version: 'latest'
update:
  canaries: 1
  max_in_flight: 1
  canary_watch_time: 1000 - 90000
  update_watch_time: 1000 - 90000
EOF

bosh -n --tty -d "${SDK_DEPLOYMENT}" deploy /tmp/manifest.yml

pushd "$SRC_DIR"
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

  # BOSH in Docker name resolution doesn't play well with docker-compose services
  # however, if we provide a valid IP within the Docker network it communicates flawlessly
  # for that reason, we need to resolve the MYSQL_HOST before running the tests
  export MYSQL_HOSTNAME="$(getent hosts ${MYSQL_HOSTNAME} | awk '{ print $1 }')"

  ginkgo -mod vendor -r -v "system_tests/mysql" -trace
popd

bosh -n --tty -d "${SDK_DEPLOYMENT}" delete-deployment
